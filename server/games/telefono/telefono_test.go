package telefono

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/adriadam10/mattermost-plugin-juegos/server/engine"
	"github.com/adriadam10/mattermost-plugin-juegos/server/engine/enginetest"
)

var tinyPNG = base64.StdEncoding.EncodeToString(append(append([]byte(nil), pngMagic...), "rest"...))

func TestEveryoneTouchesEveryChainOnce(t *testing.T) {
	g := Game{}
	m, _ := enginetest.Match(t, g, engine.Live, "a", "b", "c", "d")
	s, _ := engine.Load[state](m)
	for c, chain := range s.Chains {
		seen := map[string]bool{}
		for _, st := range chain {
			if seen[st.Author] {
				t.Fatalf("chain %d has %s twice", c, st.Author)
			}
			seen[st.Author] = true
		}
	}
	for r := range len(s.Order) {
		seen := map[int]bool{}
		for _, id := range s.Order {
			c := s.task(id, r)
			if seen[c] || s.Chains[c][r].Author != id {
				t.Fatalf("step %d: %s got chain %d", r, id, c)
			}
			seen[c] = true
		}
	}
}

func TestFullGame(t *testing.T) {
	g := Game{}
	players := []string{"a", "b", "c"}
	m, env := enginetest.Match(t, g, engine.Live, players...)
	for step := range len(players) {
		phase := m.Phase
		for _, id := range players {
			move := fmt.Sprintf(`{"text":"frase %s %d"}`, id, step)
			if phase == phaseDraw {
				move = fmt.Sprintf(`{"drawing":%q}`, tinyPNG)
			}
			enginetest.Act(t, g, m, env, id, move)
		}
	}
	if m.Phase != phaseVote {
		t.Fatalf("phase %q after the last step", m.Phase)
	}
	s, _ := engine.Load[state](m)
	// Everyone votes the first step of a chain that is not theirs.
	for _, id := range players {
		for c := range s.Chains {
			if s.Chains[c][0].Author != id {
				enginetest.Act(t, g, m, env, id, fmt.Sprintf(`{"vote":"%d-0"}`, c))
				break
			}
		}
	}
	if m.Phase != engine.PhaseFinished {
		t.Fatalf("phase %q after every vote", m.Phase)
	}
	if len(m.Result.Parts) != len(players) || len(m.Result.Parts[0].Files) != 1 {
		t.Fatalf("parts %+v", m.Result.Parts)
	}
	if len(m.Blobs) != len(players) {
		t.Fatalf("%d blobs, want one drawing per chain", len(m.Blobs))
	}
}

func TestBlankStepPassesOnTheLastThing(t *testing.T) {
	g := Game{}
	m, env := enginetest.Match(t, g, engine.Async, "a", "b", "c")
	enginetest.Act(t, g, m, env, "a", `{"text":"un pulpo"}`)
	if err := g.Timeout(m, env); err != nil {
		t.Fatal(err)
	}
	// Nobody draws: the describers still get a sentence.
	if err := g.Timeout(m, env); err != nil {
		t.Fatal(err)
	}
	s, _ := engine.Load[state](m)
	c := s.task("a", 0)
	describer := s.Chains[c][2].Author
	v := g.View(m, describer, env).(map[string]any)
	prev, _ := v["prev"].(map[string]string)
	if prev["text"] != "un pulpo" {
		t.Fatalf("describer sees %v", v["prev"])
	}
}

func TestRejectsNonPNG(t *testing.T) {
	g := Game{}
	m, env := enginetest.Match(t, g, engine.Async, "a", "b", "c")
	_ = g.Timeout(m, env)
	err := g.Act(m, "a", []byte(`{"drawing":"aG9sYQ=="}`), env)
	if !engine.IsRefusal(err) {
		t.Fatalf("got %v, want a refusal", err)
	}
}
