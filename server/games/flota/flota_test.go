package flota

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/adriadam10/mattermost-plugin-juegos/server/engine"
	"github.com/adriadam10/mattermost-plugin-juegos/server/engine/enginetest"
	"github.com/adriadam10/mattermost-plugin-juegos/server/games/duel"
)

func TestValidFleet(t *testing.T) {
	rnd := rand.New(rand.NewPCG(1, 2))
	for range 50 {
		if ships := randomFleet(rnd); !validFleet(ships) {
			t.Fatalf("random fleet not valid: %v", ships)
		}
	}
	ships := randomFleet(rnd)
	bent := append([][]cell{{{0, 0}, {0, 1}, {1, 1}, {1, 2}}}, ships[1:]...)
	if validFleet(bent) {
		t.Fatal("bent ship accepted")
	}
	if validFleet(ships[1:]) {
		t.Fatal("fleet without the carrier accepted")
	}
	touching := [][]cell{{{0, 0}, {0, 1}, {0, 2}, {0, 3}}, {{1, 4}, {2, 4}, {3, 4}}}
	if validPartial(touching) {
		t.Fatal("ships touching at a corner accepted")
	}
}

func placeMove(ships [][]cell) string {
	data, _ := json.Marshal(map[string]any{"ships": ships})
	return string(data)
}

func TestHitShootsAgainAndSinkingAllWins(t *testing.T) {
	g := Game{}
	m, env := enginetest.Match(t, g, engine.Live, "a", "b")
	rnd := rand.New(rand.NewPCG(5, 6))
	fleets := map[string][][]cell{"a": randomFleet(rnd), "b": randomFleet(rnd)}
	enginetest.Act(t, g, m, env, "a", placeMove(fleets["a"]))
	enginetest.Act(t, g, m, env, "b", placeMove(fleets["b"]))
	if m.Phase != duel.PhaseTurn {
		t.Fatalf("phase %q after placing", m.Phase)
	}
	s, _ := engine.Load[state](m)
	shooter, target := s.Seats[s.Turn], s.Seats[1-s.Turn]
	for _, ship := range fleets[target] {
		for _, p := range ship {
			enginetest.Act(t, g, m, env, shooter, fmt.Sprintf(`{"row":%d,"column":%d}`, p[0], p[1]))
		}
	}
	if m.Phase != engine.PhaseFinished || m.Result.Scores[shooter] != 1 || m.Result.Unranked {
		t.Fatalf("phase %q, result %+v", m.Phase, m.Result)
	}
}

func TestPlacementTimeoutPlacesAtRandomAndBotPlays(t *testing.T) {
	g := Game{}
	m, env := enginetest.Match(t, g, engine.Async, "a")
	if err := g.Timeout(m, env); err != nil {
		t.Fatal(err)
	}
	s, _ := engine.Load[state](m)
	if !s.Sides[0].Placed || !s.Sides[1].Placed || m.Phase != duel.PhaseTurn {
		t.Fatalf("phase %q, placed %v %v", m.Phase, s.Sides[0].Placed, s.Sides[1].Placed)
	}
	if engine.IsBot(s.Seats[s.Turn]) {
		t.Fatal("the bot kept the turn")
	}
}

// The hard bot needs fewer shots than one firing at random.
func TestHardBotShootsBetter(t *testing.T) {
	rnd := rand.New(rand.NewPCG(7, 8))
	shots := func(hard bool) int {
		total := 0
		for range 20 {
			sd := side{Ships: randomFleet(rnd)}
			for len(sd.sunkSizes()) < len(fleet) {
				p := botShot(&sd.Shots, sd.leftSizes(), hard, rnd)
				s := state{Sides: [2]side{{}, sd}, Last: [2]int{-1, -1}}
				m := &engine.Match{}
				(Game{}).fire(m, &s, p, enginetest.Env())
				sd = s.Sides[1]
				total++
			}
		}
		return total / 20
	}
	easy, hard := shots(false), shots(true)
	if hard >= easy-15 {
		t.Fatalf("hard bot %d shots, easy %d", hard, easy)
	}
	t.Logf("shots to sink the fleet: hard %d, easy %d", hard, easy)
}
