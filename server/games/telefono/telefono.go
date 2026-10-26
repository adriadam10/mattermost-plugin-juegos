// Package telefono is «Teléfono dibujado»: everybody writes a sentence, the next
// player draws it, the next one describes the drawing, and so on. At the end
// every chain is revealed and people vote their favourite moment.
package telefono

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/adriadam10/mattermost-plugin-juegos/server/engine"
)

const (
	phaseWrite    = "escribir"
	phaseDraw     = "dibujar"
	phaseDescribe = "describir"
	phaseVote     = "votar"
	maxText       = 140
	maxDrawing    = 400 << 10
	contentKind   = "telefono"
)

var pngMagic = []byte("\x89PNG\r\n\x1a\n")

type Game struct{}

// step is one link of a chain: a sentence or a drawing.
type step struct {
	Author  string `json:"author"`
	Text    string `json:"text,omitempty"`
	Drawing string `json:"drawing,omitempty"`
	Done    bool   `json:"done"`
}

type state struct {
	// Chain c, step r is done by Order[(c+r) % len(Order)].
	Order   []string          `json:"order"`
	Chains  [][]step          `json:"chains"`
	Suggest map[string]string `json:"suggest"`
	// Voter → "chain-step".
	Votes engine.Ballot `json:"votes"`
}

func (Game) Info() engine.Info {
	return engine.Info{
		ID:          "telefono",
		Name:        "Teléfono dibujado",
		Emoji:       "🎨",
		Description: "Escribe una frase; el siguiente la dibuja, el otro describe el dibujo... y al final se ve en qué ha acabado cada cadena.",
		MinPlayers:  3,
		MaxPlayers:  10,
		Lobby:       true,
	}
}

func (Game) Start(m *engine.Match, env *engine.Env) error {
	order := append([]string(nil), m.Players...)
	env.Rand.Shuffle(len(order), func(i, j int) { order[i], order[j] = order[j], order[i] })
	s := &state{Order: order, Chains: make([][]step, len(order)), Suggest: map[string]string{}, Votes: engine.Ballot{}}
	for c := range s.Chains {
		s.Chains[c] = make([]step, len(order))
		for r := range s.Chains[c] {
			s.Chains[c][r].Author = order[(c+r)%len(order)]
		}
	}
	for i, text := range env.Pick(contentKind, len(order)) {
		s.Suggest[order[i]] = text
	}
	m.Round = 1
	env.SetPhase(m, phaseWrite, 1)
	return engine.Save(m, s)
}

// phaseOf: step 0 is written, then drawings and descriptions alternate.
func phaseOf(r int) (string, float64) {
	switch {
	case r == 0:
		return phaseWrite, 1
	case r%2 == 1:
		return phaseDraw, 2
	default:
		return phaseDescribe, 1
	}
}

// task is the chain userID works on in step r, or -1.
func (s *state) task(userID string, r int) int {
	for i, id := range s.Order {
		if id == userID {
			n := len(s.Order)
			return ((i-r)%n + n) % n
		}
	}
	return -1
}

type move struct {
	Text    string `json:"text"`
	Drawing string `json:"drawing"`
	Vote    string `json:"vote"`
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
	if m.Phase == phaseVote {
		if err := vote(s, userID, mv.Vote); err != nil {
			return err
		}
		if everyoneVoted(s) {
			return finish(m, s, env)
		}
		return engine.Save(m, s)
	}
	r := m.Round - 1
	c := s.task(userID, r)
	if c < 0 {
		return engine.Refusal("No estás en esta partida")
	}
	st := &s.Chains[c][r]
	if m.Phase == phaseDraw {
		key, err := storeDrawing(m, mv.Drawing, env)
		if err != nil {
			return err
		}
		st.Drawing = key
	} else {
		text := strings.Join(strings.Fields(mv.Text), " ")
		if text == "" {
			return engine.Refusal("La frase está vacía")
		}
		if utf8.RuneCountInString(text) > maxText {
			return engine.Refusal(fmt.Sprintf("Máximo %d caracteres", maxText))
		}
		st.Text = text
	}
	st.Done = true
	if roundDone(s, r) {
		next(m, s, env)
	}
	return engine.Save(m, s)
}

func storeDrawing(m *engine.Match, encoded string, env *engine.Env) (string, error) {
	data, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(encoded, "data:image/png;base64,"))
	if err != nil || !bytes.HasPrefix(data, pngMagic) {
		return "", engine.Refusal("El dibujo no es un PNG")
	}
	if len(data) > maxDrawing {
		return "", engine.Refusal("El dibujo es demasiado grande")
	}
	return env.PutBlob(m, data)
}

func roundDone(s *state, r int) bool {
	for c := range s.Chains {
		if !s.Chains[c][r].Done {
			return false
		}
	}
	return true
}

// next moves to the following step, or to the vote after the last one.
func next(m *engine.Match, s *state, env *engine.Env) {
	if m.Round >= len(s.Order) {
		env.SetPhase(m, phaseVote, 1)
		return
	}
	m.Round++
	phase, weight := phaseOf(m.Round - 1)
	env.SetPhase(m, phase, weight)
}

func vote(s *state, userID, option string) error {
	var c, r int
	if _, err := fmt.Sscanf(option, "%d-%d", &c, &r); err != nil || c < 0 || c >= len(s.Chains) || r < 0 || r >= len(s.Chains[c]) {
		return engine.Refusal("Esa opción no existe")
	}
	st := s.Chains[c][r]
	if st.Text == "" && st.Drawing == "" {
		return engine.Refusal("Ese paso quedó en blanco")
	}
	if st.Author == userID {
		return engine.Refusal("No puedes votarte a ti")
	}
	s.Votes[userID] = option
	return nil
}

func everyoneVoted(s *state) bool {
	for _, id := range s.Order {
		if _, ok := s.Votes[id]; !ok {
			return false
		}
	}
	return true
}

func (Game) Timeout(m *engine.Match, env *engine.Env) error {
	s, err := engine.Load[state](m)
	if err != nil {
		return err
	}
	if m.Phase == phaseVote {
		return finish(m, s, env)
	}
	// Whoever did not make it leaves a blank step; the next one gets the last
	// thing the chain has.
	r := m.Round - 1
	for c := range s.Chains {
		s.Chains[c][r].Done = true
	}
	next(m, s, env)
	return engine.Save(m, s)
}

func finish(m *engine.Match, s *state, env *engine.Env) error {
	scores := map[string]int{}
	for _, id := range s.Order {
		engine.TookPart(scores, id)
	}
	for _, option := range s.Votes {
		var c, r int
		if _, err := fmt.Sscanf(option, "%d-%d", &c, &r); err == nil {
			scores[s.Chains[c][r].Author]++
		}
	}
	if err := engine.Save(m, s); err != nil {
		return err
	}
	env.Finish(m, scores, voteSummary(s, env), chainParts(s, env)...)
	return nil
}
