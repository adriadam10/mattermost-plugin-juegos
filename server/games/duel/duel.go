// Package duel holds what the one-against-one games share: a challenged rival
// or the first who joins, the bot when nobody does, turns and their deadlines.
package duel

import (
	"fmt"

	"github.com/adriadam10/mattermost-plugin-juegos/server/engine"
)

const (
	// PhaseTurn is the phase of a turn: a player's move, with its own deadline.
	PhaseTurn = "turno"
	levelKey  = "nivel"
)

// Options lets the creator challenge someone and pick how well the bot plays.
var Options = []engine.Option{
	{Key: engine.OptionRival, Label: "Reto a"},
	Level,
}

// Level is how well the bot plays; the guiñote uses it too.
var Level = engine.Option{
	Key:   levelKey,
	Label: "Nivel del bot",
	Choices: []engine.Choice{
		{Value: "facil", Label: "🙂 Fácil"},
		{Value: "dificil", Label: "😈 Difícil"},
	},
	Default: "facil",
}

// Hard tells whether the bot of m plays its best.
func Hard(m *engine.Match, info engine.Info) bool { return m.Option(info, levelKey) == "dificil" }

// Seats are the two players in a random order (who starts), with the bot if
// nobody took the second seat.
func Seats(m *engine.Match, env *engine.Env) [2]string {
	s := engine.Seats(m, 2)
	if env.Rand.IntN(2) == 1 {
		s[0], s[1] = s[1], s[0]
	}
	return [2]string{s[0], s[1]}
}

// Turn starts the turn of a player; each turn is a new round, so the async
// notice tells them it is their go.
func Turn(m *engine.Match, env *engine.Env) {
	m.Round++
	env.SetPhase(m, PhaseTurn, 1)
}

// Notice is the async notice of a turn, or nothing if the bot is on turn.
func Notice(m *engine.Match, env *engine.Env, seats [2]string, turn int) string {
	if m.Phase != PhaseTurn || engine.IsBot(seats[turn]) {
		return ""
	}
	return "🎯 Te toca, " + env.Name(seats[turn])
}

// Finish ends the match: winner is a seat, or -1 for a draw. Beating the bot
// does not count for the ranking.
func Finish(m *engine.Match, env *engine.Env, seats [2]string, winner int, summary string) {
	scores := map[string]int{}
	for i, id := range seats {
		if engine.IsBot(id) {
			continue
		}
		scores[id] = 0
		if i == winner {
			scores[id] = 1
		}
	}
	env.Finish(m, scores, summary)
	m.Result.Unranked = engine.IsBot(seats[0]) || engine.IsBot(seats[1])
}

// Timeout is a turn that ran out: whoever was on turn loses.
func Timeout(m *engine.Match, env *engine.Env, seats [2]string, turn int, board string) {
	Finish(m, env, seats, 1-turn, fmt.Sprintf("⏰ %s no ha jugado a tiempo: gana %s.\n\n%s", env.Name(seats[turn]), env.Name(seats[1-turn]), board))
}

// SeatOf is the seat of userID, or -1 for someone who watches.
func SeatOf(seats [2]string, userID string) int {
	for i, id := range seats {
		if id == userID {
			return i
		}
	}
	return -1
}

// Names are the names of both seats as the webapp shows them.
func Names(env *engine.Env, seats [2]string) [2]string {
	return [2]string{env.Name(seats[0]), env.Name(seats[1])}
}
