// Package absurdas is «Respuestas absurdas»: everybody completes two prompts in
// secret, each prompt goes to two players, and the channel votes blind for the
// funnier answer of each pair.
package absurdas

import (
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/adriadam10/mattermost-plugin-juegos/server/engine"
)

const (
	phaseAnswer = "responder"
	phaseVote   = "votar"
	maxAnswer   = 140
	// Points per vote, and the bonus for taking every vote of a pair.
	votePoints  = 100
	sweepBonus  = 100
	contentKind = "absurdas"
)

type Game struct{}

type matchup struct {
	Prompt  string        `json:"prompt"`
	Authors [2]string     `json:"authors"`
	Answers [2]string     `json:"answers"`
	Votes   engine.Ballot `json:"votes"`
}

type state struct {
	Rounds   int            `json:"rounds"`
	Matchups []matchup      `json:"matchups"`
	Done     [][]matchup    `json:"done"`
	Scores   map[string]int `json:"scores"`
}

func (Game) Info() engine.Info {
	return engine.Info{
		ID:          "absurdas",
		Name:        "Respuestas absurdas",
		Emoji:       "🤪",
		Description: "Completa dos frases en secreto. Cada frase la responden dos personas y el canal vota a ciegas la más graciosa.",
		MinPlayers:  3,
		MaxPlayers:  10,
		Lobby:       true,
	}
}

func (Game) Start(m *engine.Match, env *engine.Env) error {
	s := &state{Rounds: engine.Rounds(m.Mode, 2, 1), Scores: map[string]int{}}
	for _, id := range m.Players {
		s.Scores[id] = 0
	}
	if err := deal(m, s, env); err != nil {
		return err
	}
	return engine.Save(m, s)
}

// deal gives every player two prompts: prompt k goes to players k and k+1, so
// nobody faces the same rival twice in a round (with three or more players).
func deal(m *engine.Match, s *state, env *engine.Env) error {
	players := append([]string(nil), m.Players...)
	env.Rand.Shuffle(len(players), func(i, j int) { players[i], players[j] = players[j], players[i] })
	prompts := env.Pick(contentKind, len(players))
	if len(prompts) < len(players) {
		return fmt.Errorf("only %d prompts for %d players", len(prompts), len(players))
	}
	s.Matchups = make([]matchup, len(players))
	for k := range players {
		s.Matchups[k] = matchup{
			Prompt:  prompts[k],
			Authors: [2]string{players[k], players[(k+1)%len(players)]},
			Votes:   engine.Ballot{},
		}
	}
	m.Round++
	env.SetPhase(m, phaseAnswer, 2)
	return nil
}

type move struct {
	Matchup int    `json:"matchup"`
	Text    string `json:"text"`
	Side    int    `json:"side"`
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
	if mv.Matchup < 0 || mv.Matchup >= len(s.Matchups) {
		return engine.Refusal("Esa frase no existe")
	}
	mu := &s.Matchups[mv.Matchup]
	switch m.Phase {
	case phaseAnswer:
		if err := answer(mu, userID, mv.Text); err != nil {
			return err
		}
		if allAnswered(s.Matchups) {
			env.SetPhase(m, phaseVote, 1)
		}
	case phaseVote:
		if err := vote(mu, userID, mv.Side); err != nil {
			return err
		}
		if allVoted(m.Players, s.Matchups) {
			return finishRound(m, s, env)
		}
	default:
		return engine.Refusal("Ahora no toca")
	}
	return engine.Save(m, s)
}

func answer(mu *matchup, userID, text string) error {
	text = strings.Join(strings.Fields(text), " ")
	if text == "" {
		return engine.Refusal("La respuesta está vacía")
	}
	if utf8.RuneCountInString(text) > maxAnswer {
		return engine.Refusal(fmt.Sprintf("Máximo %d caracteres", maxAnswer))
	}
	for i, author := range mu.Authors {
		if author == userID {
			mu.Answers[i] = text
			return nil
		}
	}
	return engine.Refusal("Esa frase no te ha tocado")
}

func vote(mu *matchup, userID string, side int) error {
	if mu.Authors[0] == userID || mu.Authors[1] == userID {
		return engine.Refusal("No puedes votar tu propia frase")
	}
	if side != 0 && side != 1 || mu.Answers[side] == "" {
		return engine.Refusal("Esa respuesta no existe")
	}
	mu.Votes[userID] = fmt.Sprint(side)
	return nil
}

func allAnswered(ms []matchup) bool {
	for _, mu := range ms {
		if mu.Answers[0] == "" || mu.Answers[1] == "" {
			return false
		}
	}
	return true
}

// allVoted: every player has voted every pair they are not in (the audience can
// vote too, but nobody waits for it).
func allVoted(players []string, ms []matchup) bool {
	for _, mu := range ms {
		if mu.Answers[0] == "" || mu.Answers[1] == "" {
			continue
		}
		for _, id := range players {
			_, voted := mu.Votes[id]
			if !voted && id != mu.Authors[0] && id != mu.Authors[1] {
				return false
			}
		}
	}
	return true
}

func (Game) Timeout(m *engine.Match, env *engine.Env) error {
	s, err := engine.Load[state](m)
	if err != nil {
		return err
	}
	if m.Phase == phaseAnswer {
		env.SetPhase(m, phaseVote, 1)
		return engine.Save(m, s)
	}
	return finishRound(m, s, env)
}

func finishRound(m *engine.Match, s *state, env *engine.Env) error {
	for _, mu := range s.Matchups {
		for id, pts := range points(mu) {
			s.Scores[id] += pts
		}
	}
	s.Done = append(s.Done, s.Matchups)
	s.Matchups = nil
	if len(s.Done) < s.Rounds {
		if err := deal(m, s, env); err != nil {
			return err
		}
		return engine.Save(m, s)
	}
	if err := engine.Save(m, s); err != nil {
		return err
	}
	env.Finish(m, s.Scores, summary(s, env))
	return nil
}

// points of one pair: 100 per vote and a bonus for a sweep. An answer left
// blank gives the other one a walkover.
func points(mu matchup) map[string]int {
	pts := map[string]int{}
	switch {
	case mu.Answers[0] == "" && mu.Answers[1] == "":
		return pts
	case mu.Answers[0] == "":
		pts[mu.Authors[1]] = votePoints
		return pts
	case mu.Answers[1] == "":
		pts[mu.Authors[0]] = votePoints
		return pts
	}
	tally := mu.Votes.Tally()
	for side, author := range mu.Authors {
		got := tally[fmt.Sprint(side)]
		pts[author] = got * votePoints
		if got >= 2 && got == len(mu.Votes) {
			pts[author] += sweepBonus
		}
	}
	return pts
}

func summary(s *state, env *engine.Env) string {
	var b strings.Builder
	for _, round := range s.Done {
		for _, mu := range round {
			tally := mu.Votes.Tally()
			fmt.Fprintf(&b, "**%s**\n", mu.Prompt)
			for side, author := range mu.Authors {
				text := mu.Answers[side]
				if text == "" {
					text = "_(en blanco)_"
				}
				fmt.Fprintf(&b, "- %s — @%s (%d votos)\n", text, env.Username(author), tally[fmt.Sprint(side)])
			}
		}
	}
	return b.String()
}
