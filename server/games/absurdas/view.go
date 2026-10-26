package absurdas

import (
	"strconv"

	"github.com/adriadam10/mattermost-plugin-juegos/server/engine"
)

type taskView struct {
	Matchup int    `json:"matchup"`
	Prompt  string `json:"prompt"`
	Answer  string `json:"answer"`
}

// pairView hides who wrote what until the round ends.
type pairView struct {
	Matchup int       `json:"matchup"`
	Prompt  string    `json:"prompt"`
	Answers [2]string `json:"answers"`
	Mine    bool      `json:"mine"`
	// -1 when this user has not voted it.
	Vote int `json:"vote"`
}

type revealView struct {
	Prompt  string    `json:"prompt"`
	Answers [2]string `json:"answers"`
	Authors [2]string `json:"authors"`
	Votes   [2]int    `json:"votes"`
	Points  [2]int    `json:"points"`
}

func (Game) View(m *engine.Match, userID string, env *engine.Env) any {
	s, err := engine.Load[state](m)
	if err != nil {
		return nil
	}
	v := map[string]any{"rounds": s.Rounds, "scores": s.Scores}
	switch m.Phase {
	case phaseAnswer:
		tasks := []taskView{}
		answered, total := 0, 0
		for k, mu := range s.Matchups {
			for side, author := range mu.Authors {
				total++
				if mu.Answers[side] != "" {
					answered++
				}
				if author == userID {
					tasks = append(tasks, taskView{Matchup: k, Prompt: mu.Prompt, Answer: mu.Answers[side]})
				}
			}
		}
		v["tasks"], v["answered"], v["total"] = tasks, answered, total
	case phaseVote:
		pairs := []pairView{}
		for k, mu := range s.Matchups {
			if mu.Answers[0] == "" || mu.Answers[1] == "" {
				continue
			}
			vote := -1
			if side, ok := mu.Votes[userID]; ok {
				vote, _ = strconv.Atoi(side)
			}
			pairs = append(pairs, pairView{
				Matchup: k, Prompt: mu.Prompt, Answers: mu.Answers,
				Mine: mu.Authors[0] == userID || mu.Authors[1] == userID, Vote: vote,
			})
		}
		v["pairs"] = pairs
	}
	reveals := []revealView{}
	for _, round := range s.Done {
		for _, mu := range round {
			tally := mu.Votes.Tally()
			pts := points(mu)
			reveals = append(reveals, revealView{
				Prompt:  mu.Prompt,
				Answers: mu.Answers,
				Authors: [2]string{env.Username(mu.Authors[0]), env.Username(mu.Authors[1])},
				Votes:   [2]int{tally["0"], tally["1"]},
				Points:  [2]int{pts[mu.Authors[0]], pts[mu.Authors[1]]},
			})
		}
	}
	v["reveals"] = reveals
	return v
}

func (Game) Notice(m *engine.Match, env *engine.Env) string {
	switch m.Phase {
	case phaseAnswer:
		return "✍️ Toca responder: " + env.Mentions(m.Players)
	case phaseVote:
		return "🗳️ Ya se puede votar. Puede votar todo el canal, no solo los jugadores."
	}
	return ""
}
