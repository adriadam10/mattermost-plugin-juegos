// Package enginetest runs games without Mattermost, for their tests.
package enginetest

import (
	"fmt"
	"math/rand/v2"
	"testing"
	"time"

	"github.com/adriadam10/mattermost-plugin-juegos/server/engine"
)

// Services hands out numbered texts and keeps blobs in memory.
type Services struct {
	Blobs map[string][]byte
}

func (s *Services) Pick(kind string, n int) []string {
	out := make([]string, n)
	for i := range out {
		out[i] = fmt.Sprintf("%s %d ___", kind, i+1)
	}
	return out
}

func (s *Services) Gif(tag string) (string, error) { return "https://gif.example/" + tag, nil }

func (s *Services) PutBlob(data []byte) (string, error) {
	if s.Blobs == nil {
		s.Blobs = map[string][]byte{}
	}
	key := fmt.Sprintf("blob-%d", len(s.Blobs))
	s.Blobs[key] = data
	return key, nil
}

func (s *Services) Username(userID string) string { return "u" + userID }

// Env is a fixed-seed environment with phases of a minute.
func Env() *engine.Env {
	return &engine.Env{Services: &Services{}, Now: time.Unix(1_800_000_000, 0), Rand: rand.New(rand.NewPCG(1, 2)), Base: time.Minute}
}

// Match is a new match of game with players, already started.
func Match(t *testing.T, g engine.Game, mode engine.Mode, players ...string) (*engine.Match, *engine.Env) {
	t.Helper()
	m := &engine.Match{ID: "m", Game: g.Info().ID, Mode: mode, Players: players}
	env := Env()
	if err := g.Start(m, env); err != nil {
		t.Fatalf("start: %v", err)
	}
	return m, env
}

// Act makes a move and fails the test if it is not accepted.
func Act(t *testing.T, g engine.Game, m *engine.Match, env *engine.Env, userID, move string) {
	t.Helper()
	if err := g.Act(m, userID, []byte(move), env); err != nil {
		t.Fatalf("%s %s: %v", userID, move, err)
	}
}
