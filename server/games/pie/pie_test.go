package pie

import (
	"testing"

	"github.com/adriadam10/mattermost-plugin-juegos/server/engine"
	"github.com/adriadam10/mattermost-plugin-juegos/server/engine/enginetest"
)

func TestVotesAreBlindAndCount(t *testing.T) {
	g := Game{}
	m, env := enginetest.Match(t, g, engine.Async)
	enginetest.Act(t, g, m, env, "a", `{"text":"pie de a"}`)
	enginetest.Act(t, g, m, env, "b", `{"text":"pie de b"}`)
	enginetest.Act(t, g, m, env, "c", `{"text":"pie de c"}`)
	if err := g.Timeout(m, env); err != nil {
		t.Fatal(err)
	}
	s, _ := engine.Load[state](m)
	of := map[string]int{}
	for i, id := range s.Current.Order {
		of[id] = i
	}
	if err := g.Act(m, "a", []byte(`{"option":`+itoa(of["a"])+`}`), env); !engine.IsRefusal(err) {
		t.Fatalf("own vote: got %v, want a refusal", err)
	}
	enginetest.Act(t, g, m, env, "a", `{"option":`+itoa(of["b"])+`}`)
	enginetest.Act(t, g, m, env, "c", `{"option":`+itoa(of["b"])+`}`)
	if err := g.Timeout(m, env); err != nil {
		t.Fatal(err)
	}
	if m.Phase != engine.PhaseFinished || m.Result.Scores["b"] != 2 || len(m.Result.Scores) != 3 {
		t.Fatalf("phase %q, scores %v", m.Phase, m.Result)
	}
}

func TestRoundWithOneCaptionIsVoid(t *testing.T) {
	g := Game{}
	m, env := enginetest.Match(t, g, engine.Async)
	enginetest.Act(t, g, m, env, "a", `{"text":"solo"}`)
	if err := g.Timeout(m, env); err != nil {
		t.Fatal(err)
	}
	if m.Phase != engine.PhaseFinished {
		t.Fatalf("phase %q, want the match over", m.Phase)
	}
}

func itoa(n int) string { return string(rune('0' + n)) }
