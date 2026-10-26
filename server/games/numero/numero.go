// Package numero is «Número único más bajo»: everybody picks a number in secret
// and the lowest one that nobody else picked wins the round.
package numero

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/adriadam10/mattermost-plugin-juegos/server/engine"
)

const (
	maxNumber = 20
	// Fewer pickers make it a coin toss: the round does not count.
	minPickers = 3
	phasePick  = "elegir"
)

type Game struct{}

type round struct {
	Picks map[string]int `json:"picks"`
	// Empty when nobody had a unique number or too few picked.
	Winner string `json:"winner,omitempty"`
	Number int    `json:"number,omitempty"`
}

type state struct {
	Rounds int            `json:"rounds"`
	Done   []round        `json:"done"`
	Picks  map[string]int `json:"picks"`
}

func (Game) Info() engine.Info {
	return engine.Info{
		ID:          "numero",
		Name:        "Número único más bajo",
		Emoji:       "🔢",
		Description: fmt.Sprintf("Elige en secreto un número del 1 al %d. Gana el más bajo que nadie más haya elegido.", maxNumber),
		MinPlayers:  minPickers,
	}
}

func (Game) Start(m *engine.Match, env *engine.Env) error {
	m.Round = 1
	env.SetPhase(m, phasePick, 0.5)
	return engine.Save(m, state{Rounds: engine.Rounds(m.Mode, 3, 1), Picks: map[string]int{}})
}

func (Game) Act(m *engine.Match, userID string, move json.RawMessage, env *engine.Env) error {
	if m.Phase != phasePick {
		return engine.Refusal("Ahora no se puede elegir")
	}
	mv, err := engine.Decode[struct {
		Number int `json:"number"`
	}](move)
	if err != nil {
		return err
	}
	if mv.Number < 1 || mv.Number > maxNumber {
		return engine.Refusal(fmt.Sprintf("Tiene que ser del 1 al %d", maxNumber))
	}
	s, err := engine.Load[state](m)
	if err != nil {
		return err
	}
	if s.Picks == nil {
		s.Picks = map[string]int{}
	}
	s.Picks[userID] = mv.Number
	m.AddPlayer(userID)
	return engine.Save(m, s)
}

func (Game) Timeout(m *engine.Match, env *engine.Env) error {
	s, err := engine.Load[state](m)
	if err != nil {
		return err
	}
	s.Done = append(s.Done, settle(s.Picks))
	s.Picks = map[string]int{}
	if len(s.Done) < s.Rounds {
		m.Round++
		env.SetPhase(m, phasePick, 0.5)
		return engine.Save(m, s)
	}
	scores := map[string]int{}
	for _, r := range s.Done {
		for id := range r.Picks {
			engine.TookPart(scores, id)
		}
		if r.Winner != "" {
			scores[r.Winner]++
		}
	}
	if err := engine.Save(m, s); err != nil {
		return err
	}
	env.Finish(m, scores, summary(s.Done, env))
	return nil
}

// settle finds the lowest number picked by exactly one player.
func settle(picks map[string]int) round {
	r := round{Picks: picks}
	if len(picks) < minPickers {
		return r
	}
	count := map[int]int{}
	for _, n := range picks {
		count[n]++
	}
	for id, n := range picks {
		if count[n] == 1 && (r.Winner == "" || n < r.Number) {
			r.Winner, r.Number = id, n
		}
	}
	return r
}

func (Game) View(m *engine.Match, userID string, env *engine.Env) any {
	s, err := engine.Load[state](m)
	if err != nil {
		return nil
	}
	mine, picked := s.Picks[userID]
	done := make([]map[string]any, len(s.Done))
	for i, r := range s.Done {
		picks := map[string]int{}
		for id, n := range r.Picks {
			picks[env.Username(id)] = n
		}
		done[i] = map[string]any{"picks": picks, "winner": env.Username(r.Winner), "number": r.Number}
	}
	return map[string]any{
		"max":     maxNumber,
		"rounds":  s.Rounds,
		"pickers": len(s.Picks),
		"mine":    mine,
		"picked":  picked,
		"done":    done,
		"min":     minPickers,
	}
}

func summary(rounds []round, env *engine.Env) string {
	var b strings.Builder
	for i, r := range rounds {
		fmt.Fprintf(&b, "**Ronda %d**: ", i+1)
		switch {
		case len(r.Picks) < minPickers:
			fmt.Fprintf(&b, "no cuenta, eligieron %d (hacen falta %d).\n", len(r.Picks), minPickers)
			continue
		case r.Winner == "":
			b.WriteString("nadie tuvo un número único. ")
		default:
			fmt.Fprintf(&b, "gana @%s con el **%d**. ", env.Username(r.Winner), r.Number)
		}
		b.WriteString(histogram(r.Picks, env))
		b.WriteString("\n")
	}
	return b.String()
}

// histogram lists each number with who picked it, lowest first.
func histogram(picks map[string]int, env *engine.Env) string {
	byNumber := map[int][]string{}
	for id, n := range picks {
		byNumber[n] = append(byNumber[n], "@"+env.Username(id))
	}
	numbers := make([]int, 0, len(byNumber))
	for n := range byNumber {
		numbers = append(numbers, n)
	}
	sort.Ints(numbers)
	parts := make([]string, len(numbers))
	for i, n := range numbers {
		sort.Strings(byNumber[n])
		parts[i] = fmt.Sprintf("%d: %s", n, strings.Join(byNumber[n], " "))
	}
	return "(" + strings.Join(parts, " · ") + ")"
}

func (Game) Notice(m *engine.Match, _ *engine.Env) string {
	if m.Phase != phasePick {
		return ""
	}
	return fmt.Sprintf("🔢 Ronda %d: elegid vuestro número.", m.Round)
}
