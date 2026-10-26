package telefono

import (
	"fmt"
	"strings"

	"github.com/adriadam10/mattermost-plugin-juegos/server/engine"
)

type stepView struct {
	User    string `json:"user"`
	Text    string `json:"text,omitempty"`
	Drawing string `json:"drawing,omitempty"`
	Option  string `json:"option"`
	Mine    bool   `json:"mine"`
	Votes   int    `json:"votes"`
}

func (Game) View(m *engine.Match, userID string, env *engine.Env) any {
	s, err := engine.Load[state](m)
	if err != nil {
		return nil
	}
	v := map[string]any{"steps": len(s.Order)}
	switch m.Phase {
	case phaseWrite, phaseDraw, phaseDescribe:
		r := m.Round - 1
		done := 0
		for c := range s.Chains {
			if s.Chains[c][r].Done {
				done++
			}
		}
		v["done"] = done
		if c := s.task(userID, r); c >= 0 {
			st := s.Chains[c][r]
			v["mine"] = map[string]any{"text": st.Text, "drawing": st.Drawing, "done": st.Done}
			v["prev"] = previous(s.Chains[c], r)
			v["suggestion"] = s.Suggest[userID]
		}
	case phaseVote, engine.PhaseFinished:
		v["chains"] = chains(s, userID, m.Phase == engine.PhaseFinished, env)
		v["vote"] = s.Votes[userID]
		v["voted"] = len(s.Votes)
	}
	return v
}

// previous is the last step before r with something in it, so a blank step
// does not leave the next player with nothing.
func previous(chain []step, r int) map[string]string {
	for i := r - 1; i >= 0; i-- {
		if chain[i].Text != "" || chain[i].Drawing != "" {
			return map[string]string{"text": chain[i].Text, "drawing": chain[i].Drawing}
		}
	}
	return nil
}

// chains shows every chain; who voted what only once the match is over.
func chains(s *state, userID string, showVotes bool, env *engine.Env) [][]stepView {
	tally := s.Votes.Tally()
	out := make([][]stepView, len(s.Chains))
	for c, chain := range s.Chains {
		out[c] = []stepView{}
		for r, st := range chain {
			if st.Text == "" && st.Drawing == "" {
				continue
			}
			option := fmt.Sprintf("%d-%d", c, r)
			sv := stepView{User: env.Username(st.Author), Text: st.Text, Drawing: st.Drawing, Option: option, Mine: st.Author == userID}
			if showVotes {
				sv.Votes = tally[option]
			}
			out[c] = append(out[c], sv)
		}
	}
	return out
}

func voteSummary(s *state, env *engine.Env) string {
	tally := s.Votes.Tally()
	best, most := "", 0
	for option, n := range tally {
		if n > most || n == most && option < best {
			best, most = option, n
		}
	}
	if best == "" {
		return "Nadie votó. Las cadenas, aquí abajo 👇"
	}
	var c, r int
	_, _ = fmt.Sscanf(best, "%d-%d", &c, &r)
	st := s.Chains[c][r]
	what := "el dibujo"
	if st.Drawing == "" {
		what = fmt.Sprintf("«%s»", st.Text)
	}
	return fmt.Sprintf("Momento más votado (%d votos): %s de @%s, en la cadena %d. Las cadenas, aquí abajo 👇", most, what, env.Username(st.Author), c+1)
}

// chainParts posts each chain as its own reply, with its drawings attached in order.
func chainParts(s *state, env *engine.Env) []engine.Part {
	parts := make([]engine.Part, 0, len(s.Chains))
	for c, chain := range s.Chains {
		var b strings.Builder
		fmt.Fprintf(&b, "**Cadena %d**\n", c+1)
		var files []engine.File
		for r, st := range chain {
			switch {
			case st.Text != "":
				fmt.Fprintf(&b, "%d. «%s» — @%s\n", r+1, st.Text, env.Username(st.Author))
			case st.Drawing != "":
				files = append(files, engine.File{Name: fmt.Sprintf("cadena%d-paso%d.png", c+1, r+1), Blob: st.Drawing})
				fmt.Fprintf(&b, "%d. 🖌️ dibujo %d — @%s\n", r+1, len(files), env.Username(st.Author))
			default:
				fmt.Fprintf(&b, "%d. _(en blanco)_ — @%s\n", r+1, env.Username(st.Author))
			}
		}
		parts = append(parts, engine.Part{Message: b.String(), Files: files})
	}
	return parts
}

func (Game) Notice(m *engine.Match, env *engine.Env) string {
	switch m.Phase {
	case phaseWrite:
		return "✍️ Escribid vuestra frase de arranque: " + env.Mentions(m.Players)
	case phaseDraw:
		return fmt.Sprintf("🖌️ Paso %d: toca dibujar.", m.Round)
	case phaseDescribe:
		return fmt.Sprintf("🔎 Paso %d: toca describir el dibujo que os ha llegado.", m.Round)
	case phaseVote:
		return "🗳️ Cadenas reveladas: votad vuestro momento favorito."
	}
	return ""
}
