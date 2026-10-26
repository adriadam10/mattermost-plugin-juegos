package absurdas

import (
	"fmt"
	"testing"

	"github.com/adriadam10/mattermost-plugin-juegos/server/engine"
	"github.com/adriadam10/mattermost-plugin-juegos/server/engine/enginetest"
)

func TestEveryPlayerGetsTwoPromptsWithDifferentRivals(t *testing.T) {
	g := Game{}
	m, _ := enginetest.Match(t, g, engine.Async, "a", "b", "c", "d")
	s, _ := engine.Load[state](m)
	count := map[string]int{}
	for _, mu := range s.Matchups {
		if mu.Authors[0] == mu.Authors[1] {
			t.Fatalf("a player against themselves: %v", mu.Authors)
		}
		count[mu.Authors[0]]++
		count[mu.Authors[1]]++
	}
	for _, id := range m.Players {
		if count[id] != 2 {
			t.Fatalf("%s has %d prompts", id, count[id])
		}
	}
}

func TestFullRoundWithSweep(t *testing.T) {
	g := Game{}
	players := []string{"a", "b", "c"}
	m, env := enginetest.Match(t, g, engine.Async, players...)
	s, _ := engine.Load[state](m)
	for k, mu := range s.Matchups {
		for _, author := range mu.Authors {
			enginetest.Act(t, g, m, env, author, fmt.Sprintf(`{"matchup":%d,"text":"respuesta de %s"}`, k, author))
		}
	}
	if m.Phase != phaseVote {
		t.Fatalf("phase %q after every answer", m.Phase)
	}
	s, _ = engine.Load[state](m)
	// In each pair the only player left votes for the first answer, plus an
	// outsider from the channel: a sweep of two votes.
	for k, mu := range s.Matchups {
		for _, voter := range append([]string{"x"}, players...) {
			if voter != mu.Authors[0] && voter != mu.Authors[1] {
				enginetest.Act(t, g, m, env, voter, fmt.Sprintf(`{"matchup":%d,"side":0}`, k))
			}
		}
	}
	if m.Phase != engine.PhaseFinished {
		t.Fatalf("phase %q after every vote", m.Phase)
	}
	total := 0
	for _, pts := range m.Result.Scores {
		total += pts
	}
	if want := 3 * (2*votePoints + sweepBonus); total != want {
		t.Fatalf("total points %d, want %d (%v)", total, want, m.Result.Scores)
	}
}

func TestCannotVoteOwnPair(t *testing.T) {
	g := Game{}
	m, env := enginetest.Match(t, g, engine.Async, "a", "b", "c")
	if err := g.Timeout(m, env); err != nil {
		t.Fatal(err)
	}
	s, _ := engine.Load[state](m)
	author := s.Matchups[0].Authors[0]
	if err := g.Act(m, author, []byte(`{"matchup":0,"side":1}`), env); !engine.IsRefusal(err) {
		t.Fatalf("got %v, want a refusal", err)
	}
}

func TestBlankAnswerIsAWalkover(t *testing.T) {
	pts := points(matchup{Authors: [2]string{"a", "b"}, Answers: [2]string{"", "algo"}})
	if pts["b"] != votePoints || pts["a"] != 0 {
		t.Fatalf("points %v", pts)
	}
}
