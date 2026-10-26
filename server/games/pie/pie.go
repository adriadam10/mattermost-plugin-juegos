// Package pie is «Pie de foto»: a random GIF, everybody writes a caption and the
// channel votes blind for the best one.
package pie

import (
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/adriadam10/mattermost-plugin-juegos/server/engine"
)

const (
	phaseWrite  = "escribir"
	phaseVote   = "votar"
	maxCaption  = 140
	contentKind = "pie"
)

type Game struct{}

type round struct {
	Gif      string            `json:"gif"`
	Captions map[string]string `json:"captions"`
	// Authors in the order the captions are shown for voting, so votes and
	// views carry a position and never an author.
	Order []string      `json:"order"`
	Votes engine.Ballot `json:"votes"`
}

type state struct {
	Rounds  int     `json:"rounds"`
	Current round   `json:"current"`
	Done    []round `json:"done"`
}

func (Game) Info() engine.Info {
	return engine.Info{
		ID:          "pie",
		Name:        "Pie de foto",
		Emoji:       "🖼️",
		Description: "Sale un GIF al azar: escribe el mejor pie de foto y el canal vota a ciegas.",
		MinPlayers:  2,
	}
}

func (Game) Start(m *engine.Match, env *engine.Env) error {
	s := &state{Rounds: engine.Rounds(m.Mode, 3, 1)}
	if err := newRound(m, s, env); err != nil {
		return err
	}
	return engine.Save(m, s)
}

func newRound(m *engine.Match, s *state, env *engine.Env) error {
	tags := env.Pick(contentKind, 1)
	if len(tags) == 0 {
		return fmt.Errorf("no GIF tags")
	}
	gif, err := env.Gif(tags[0])
	if err != nil {
		return fmt.Errorf("random GIF: %w", err)
	}
	s.Current = round{Gif: gif, Captions: map[string]string{}, Votes: engine.Ballot{}}
	m.Round++
	env.SetPhase(m, phaseWrite, 1)
	return nil
}

type move struct {
	Text   string `json:"text"`
	Option int    `json:"option"`
}

func (Game) Act(m *engine.Match, userID string, raw json.RawMessage, env *engine.Env) error {
	mv, err := engine.Decode[move](raw)
	if err != nil {
		return err
	}
	s, err := engine.Load[state](m)
	if err != nil {
		return err
	}
	r := &s.Current
	switch m.Phase {
	case phaseWrite:
		text := strings.Join(strings.Fields(mv.Text), " ")
		if text == "" {
			return engine.Refusal("El pie está vacío")
		}
		if utf8.RuneCountInString(text) > maxCaption {
			return engine.Refusal(fmt.Sprintf("Máximo %d caracteres", maxCaption))
		}
		r.Captions[userID] = text
		m.AddPlayer(userID)
	case phaseVote:
		if mv.Option < 0 || mv.Option >= len(r.Order) {
			return engine.Refusal("Ese pie no existe")
		}
		if r.Order[mv.Option] == userID {
			return engine.Refusal("No puedes votar tu propio pie")
		}
		r.Votes[userID] = r.Order[mv.Option]
	default:
		return engine.Refusal("Ahora no toca")
	}
	return engine.Save(m, s)
}

func (Game) Timeout(m *engine.Match, env *engine.Env) error {
	s, err := engine.Load[state](m)
	if err != nil {
		return err
	}
	r := &s.Current
	// With fewer than two captions there is nothing to vote: the round is over.
	if m.Phase == phaseWrite && len(r.Captions) >= 2 {
		for id := range r.Captions {
			r.Order = append(r.Order, id)
		}
		env.Rand.Shuffle(len(r.Order), func(i, j int) { r.Order[i], r.Order[j] = r.Order[j], r.Order[i] })
		env.SetPhase(m, phaseVote, 0.5)
		return engine.Save(m, s)
	}
	s.Done = append(s.Done, *r)
	s.Current = round{}
	if len(s.Done) < s.Rounds {
		if err := newRound(m, s, env); err != nil {
			return err
		}
		return engine.Save(m, s)
	}
	if err := engine.Save(m, s); err != nil {
		return err
	}
	env.Finish(m, scores(s.Done), summary(s.Done, env))
	return nil
}

// scores: a point per vote received; writing a caption is taking part.
func scores(rounds []round) map[string]int {
	pts := map[string]int{}
	for _, r := range rounds {
		for id := range r.Captions {
			engine.TookPart(pts, id)
		}
		for _, author := range r.Votes {
			pts[author]++
		}
	}
	return pts
}

func summary(rounds []round, env *engine.Env) string {
	var b strings.Builder
	for i, r := range rounds {
		fmt.Fprintf(&b, "**Ronda %d** · [GIF](%s)\n", i+1, r.Gif)
		if len(r.Captions) < 2 {
			b.WriteString("_No hubo pies suficientes para votar._\n")
			continue
		}
		tally := r.Votes.Tally()
		for _, id := range engine.Ranked(withCaptions(tally, r.Captions)) {
			fmt.Fprintf(&b, "- %s — @%s (%d votos)\n", r.Captions[id], env.Username(id), tally[id])
		}
	}
	return b.String()
}

// withCaptions gives every author an entry, also those with no votes.
func withCaptions(tally map[string]int, captions map[string]string) map[string]int {
	all := make(map[string]int, len(captions))
	for id := range captions {
		all[id] = tally[id]
	}
	return all
}
