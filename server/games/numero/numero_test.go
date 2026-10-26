package numero

import (
	"testing"

	"github.com/adriadam10/mattermost-plugin-juegos/server/engine"
	"github.com/adriadam10/mattermost-plugin-juegos/server/engine/enginetest"
)

func TestSettleLowestUnique(t *testing.T) {
	r := settle(map[string]int{"a": 1, "b": 1, "c": 2, "d": 5})
	if r.Winner != "c" || r.Number != 2 {
		t.Fatalf("winner %q with %d, want c with 2", r.Winner, r.Number)
	}
	if r := settle(map[string]int{"a": 1, "b": 1, "c": 1}); r.Winner != "" {
		t.Fatalf("no unique number, got winner %q", r.Winner)
	}
	if r := settle(map[string]int{"a": 1, "b": 2}); r.Winner != "" {
		t.Fatalf("too few pickers, got winner %q", r.Winner)
	}
}

func TestLiveMatchPlaysThreeRounds(t *testing.T) {
	g := Game{}
	m, env := enginetest.Match(t, g, engine.Live)
	for range 3 {
		enginetest.Act(t, g, m, env, "a", `{"number":1}`)
		enginetest.Act(t, g, m, env, "b", `{"number":1}`)
		enginetest.Act(t, g, m, env, "c", `{"number":3}`)
		if err := g.Timeout(m, env); err != nil {
			t.Fatal(err)
		}
	}
	if m.Phase != engine.PhaseFinished {
		t.Fatalf("phase %q after three rounds", m.Phase)
	}
	if got := m.Result.Scores; got["c"] != 3 || got["a"] != 0 || len(got) != 3 {
		t.Fatalf("scores %v", got)
	}
}

func TestRefusesOutOfRange(t *testing.T) {
	g := Game{}
	m, env := enginetest.Match(t, g, engine.Async)
	if err := g.Act(m, "a", []byte(`{"number":21}`), env); !engine.IsRefusal(err) {
		t.Fatalf("got %v, want a refusal", err)
	}
}
