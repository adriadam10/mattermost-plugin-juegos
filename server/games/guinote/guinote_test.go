package guinote

import (
	"encoding/json"
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/adriadam10/mattermost-plugin-juegos/server/engine"
	"github.com/adriadam10/mattermost-plugin-juegos/server/engine/enginetest"
)

func TestBeatsAndWinning(t *testing.T) {
	trick := []play{{0, "o3"}, {1, "o1"}, {2, "c1"}, {3, "o12"}}
	if w := winning(trick, 'e'); w.Seat != 1 {
		t.Fatalf("ace of the suit led should win, got seat %d", w.Seat)
	}
	trick[2].Card = "e2"
	if w := winning(trick, 'e'); w.Seat != 2 {
		t.Fatalf("the two of trumps should win, got seat %d", w.Seat)
	}
	if !beats("o10", "o11", 'o', 'e') {
		t.Fatal("in guiñote the jack beats the knight")
	}
}

func TestLegalInArrastre(t *testing.T) {
	hand := []card{"o4", "o1", "e2", "c5"}
	// Opponent winning with the three of oros: beat it with the ace.
	if got := legal(hand, []play{{1, "o3"}}, 0, 'e', true); !slices.Equal(got, []card{"o1"}) {
		t.Fatalf("must beat: %v", got)
	}
	// Partner winning: any oros.
	if got := legal(hand, []play{{2, "o3"}, {3, "o5"}}, 0, 'e', true); len(got) != 2 {
		t.Fatalf("partner winning, any of the suit: %v", got)
	}
	// No copas, opponent winning: trump.
	if got := legal([]card{"o4", "e2"}, []play{{1, "c3"}}, 0, 'e', true); !slices.Equal(got, []card{"e2"}) {
		t.Fatalf("must trump: %v", got)
	}
	// Before the arrastre, anything goes.
	if got := legal(hand, []play{{1, "o3"}}, 0, 'e', false); len(got) != len(hand) {
		t.Fatalf("free play: %v", got)
	}
}

// Four bots play whole cotos without breaking a rule: the cards add up and
// every coto ends.
func TestBotsPlayWholeCotos(t *testing.T) {
	for seed := range uint64(30) {
		for _, level := range []string{"facil", "dificil"} {
			g := Game{}
			m := &engine.Match{ID: "m", Game: "guinote", Mode: engine.Live, Options: map[string]string{"nivel": level}}
			env := enginetest.Env()
			env.Rand = rand.New(rand.NewPCG(seed, 9))
			if err := g.Start(m, env); err != nil {
				t.Fatal(err)
			}
			if m.Phase != engine.PhaseFinished {
				t.Fatalf("seed %d %s: bots did not finish, phase %q", seed, level, m.Phase)
			}
			if !m.Result.Unranked || len(m.Result.Scores) != 0 {
				t.Fatalf("bots-only result %+v", m.Result)
			}
		}
	}
}

func TestHumanPlaysWithBotsAndCardsAddUp(t *testing.T) {
	g := Game{}
	m, env := enginetest.Match(t, g, engine.Live, "a")
	for !m.Over() {
		s, _ := engine.Load[state](m)
		h := &s.H
		if h.Trick == nil && len(h.Hands[0]) == handSize && len(h.Played) == 0 {
			if total := len(h.Deck) + 1 + 4*handSize; total != 40 {
				t.Fatalf("deal of %d cards", total)
			}
		}
		v := g.View(m, "a", env).(view)
		if v.Mine != 0 || len(v.Legal) == 0 {
			t.Fatalf("not my turn: %+v", v)
		}
		move, _ := json.Marshal(map[string]any{"card": v.Legal[0]})
		enginetest.Act(t, g, m, env, "a", string(move))
	}
	if !m.Result.Unranked {
		t.Fatal("playing only with bots should not count")
	}
}

func TestTimeoutLosesTheCoto(t *testing.T) {
	g := Game{}
	m, env := enginetest.Match(t, g, engine.Live, "a", "b", "c", "d")
	s, _ := engine.Load[state](m)
	late := s.Seats[s.H.Turn]
	if err := g.Timeout(m, env); err != nil {
		t.Fatal(err)
	}
	if m.Phase != engine.PhaseFinished || m.Result.Scores[late] != 0 || m.Result.Unranked {
		t.Fatalf("result %+v", m.Result)
	}
	winners := 0
	for _, v := range m.Result.Scores {
		winners += v
	}
	if winners != 2 {
		t.Fatalf("the other pair should win: %v", m.Result.Scores)
	}
}

func TestCanteAndChange(t *testing.T) {
	m := &engine.Match{Phase: "turno"}
	env := enginetest.Env()
	s := &state{Seats: [4]string{"a", "bot:1", "c", "bot:3"}}
	s.H = hand{Trump: 'o', Muestra: "o1", Deck: []card{"c2"}, LastWinner: 2, Turn: 2,
		Hands: [4][]card{{"o12", "o10", "o7", "c12", "c10", "b2"}}}
	if err := s.cante(m, 1, 'o', "x", env); !engine.IsRefusal(err) {
		t.Fatalf("the other pair cannot sing: %v", err)
	}
	if err := s.cante(m, 0, 'o', "a", env); err != nil {
		t.Fatal(err)
	}
	if err := s.cante(m, 0, 'c', "a", env); err != nil {
		t.Fatal(err)
	}
	if s.H.Cantes[0] != 60 {
		t.Fatalf("40 + 20, got %d", s.H.Cantes[0])
	}
	if err := s.cante(m, 0, 'o', "a", env); !engine.IsRefusal(err) {
		t.Fatal("sang the forty twice")
	}
	if err := s.change(0, "a"); err != nil || s.H.Muestra != "o7" || !slices.Contains(s.H.Hands[0], "o1") {
		t.Fatalf("change: %v, muestra %s", err, s.H.Muestra)
	}
}

func TestEndHandNeedsThirtyWithoutCantes(t *testing.T) {
	m := &engine.Match{Phase: "turno"}
	env := enginetest.Env()
	s := &state{Seats: [4]string{"a", "b", "c", "d"}}
	s.H = hand{LastWinner: 1, Won: [2]int{25, 105}, Cantes: [2]int{80, 0}}
	s.endHand(m, env)
	if s.Games != [2]int{0, 1} {
		t.Fatalf("games %v: 105 with 25 + 80 of cantes does not win", s.Games)
	}
	s.H = hand{LastWinner: 0, Won: [2]int{60, 70}, Cantes: [2]int{0, 0}}
	s.endHand(m, env)
	if !s.Vueltas || s.Carry != [2]int{60, 70} {
		t.Fatalf("nobody at 101 should go to vueltas: %v %v", s.Vueltas, s.Carry)
	}
}

// A pair of hard bots beats a pair of easy ones most of the time.
func TestHardBotsBeatEasyOnes(t *testing.T) {
	won := 0
	const cotos = 200
	for seed := range uint64(cotos) {
		m := &engine.Match{Phase: "turno"}
		env := enginetest.Env()
		env.Rand = rand.New(rand.NewPCG(seed, 11))
		s := &state{Seats: [4]string{"bot:0", "bot:1", "bot:2", "bot:3"}, Dealer: int(seed % 4)}
		s.deal(env)
		for !m.Over() {
			seat := s.H.Turn
			if s.mayTalk(m, seat) {
				for _, suit := range s.canteSuits(seat) {
					_ = s.cante(m, seat, suit, "bot", env)
				}
				if m.Over() {
					break
				}
			}
			if err := s.play(m, s.H.Turn, botCard(&s.H, s.H.Turn, s.H.Turn%2 == 0, env), env); err != nil {
				t.Fatal(err)
			}
		}
		if s.Games[0] > s.Games[1] {
			won++
		}
	}
	t.Logf("hard pair won %d of %d cotos", won, cotos)
	if won < cotos*6/10 {
		t.Fatalf("hard pair only won %d of %d cotos", won, cotos)
	}
}
