package cuatro

import (
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/adriadam10/mattermost-plugin-juegos/server/engine"
	"github.com/adriadam10/mattermost-plugin-juegos/server/engine/enginetest"
)

func newMatch(t *testing.T, players ...string) (*engine.Match, *engine.Env, *state) {
	t.Helper()
	m, env := enginetest.Match(t, Game{}, engine.Live, players...)
	s, err := engine.Load[state](m)
	if err != nil {
		t.Fatal(err)
	}
	return m, env, s
}

func column(c int) string { return fmt.Sprintf(`{"column":%d}`, c) }

func TestFourInARowWins(t *testing.T) {
	g := Game{}
	m, env, s := newMatch(t, "a", "b")
	first, second := s.Seats[0], s.Seats[1]
	for range 3 {
		enginetest.Act(t, g, m, env, first, column(0))
		enginetest.Act(t, g, m, env, second, column(1))
	}
	enginetest.Act(t, g, m, env, first, column(0))
	if m.Phase != engine.PhaseFinished {
		t.Fatalf("phase %q after a vertical four", m.Phase)
	}
	if m.Result.Scores[first] != 1 || m.Result.Scores[second] != 0 || m.Result.Unranked {
		t.Fatalf("result %+v", m.Result)
	}
}

func TestRefusesOutOfTurnAndFullColumn(t *testing.T) {
	g := Game{}
	m, env, s := newMatch(t, "a", "b")
	if err := g.Act(m, s.Seats[1], []byte(column(0)), env); !engine.IsRefusal(err) {
		t.Fatalf("out of turn: %v", err)
	}
	for i := range rows {
		enginetest.Act(t, g, m, env, s.Seats[i%2], column(3))
	}
	if err := g.Act(m, s.Seats[0], []byte(column(3)), env); !engine.IsRefusal(err) {
		t.Fatalf("full column: %v", err)
	}
}

func TestTimeoutLosesAndBotGamesAreUnranked(t *testing.T) {
	g := Game{}
	m, env, s := newMatch(t, "a")
	human := s.Seats[0]
	if engine.IsBot(human) {
		human = s.Seats[1]
	}
	if err := g.Timeout(m, env); err != nil {
		t.Fatal(err)
	}
	if !m.Result.Unranked || m.Result.Scores[human] != 0 || len(m.Result.Scores) != 1 {
		t.Fatalf("result %+v", m.Result)
	}
}

func TestHardBotTakesAndBlocksWins(t *testing.T) {
	var b board
	b.drop(0, 2)
	b.drop(0, 2)
	b.drop(0, 2)
	rnd := rand.New(rand.NewPCG(1, 2))
	if c := botMove(b, 1, true, rnd); c != 0 {
		t.Fatalf("did not block, played %d", c)
	}
	b.drop(6, 1)
	b.drop(5, 1)
	b.drop(4, 1)
	if c := botMove(b, 1, true, rnd); c != 3 {
		t.Fatalf("did not win, played %d", c)
	}
}

// The hard bot beats one that plays at random.
func TestHardBotBeatsRandom(t *testing.T) {
	rnd := rand.New(rand.NewPCG(3, 4))
	for game := range 5 {
		var b board
		for turn := 0; ; turn++ {
			who := turn%2 + 1
			c := botMove(b, who, who == 1, rnd)
			r := b.drop(c, who)
			if b.line(r, c) != nil {
				if who != 1 {
					t.Fatalf("game %d: random player won", game)
				}
				break
			}
			if b.full() {
				t.Fatalf("game %d: draw", game)
			}
		}
	}
}
