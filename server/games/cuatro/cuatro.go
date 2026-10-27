// Package cuatro is «4 en raya»: two players drop discs in turns and the first
// to line up four wins. Against a challenged rival, the first who joins or the bot.
package cuatro

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/adriadam10/mattermost-plugin-juegos/server/engine"
	"github.com/adriadam10/mattermost-plugin-juegos/server/games/duel"
)

type Game struct{}

type state struct {
	Seats [2]string `json:"seats"`
	Turn  int       `json:"turn"`
	Board board     `json:"board"`
	// Columns played, in order.
	Moves []int `json:"moves"`
	// Winning cells, once someone lined up four.
	Line [][2]int `json:"line,omitempty"`
}

func (Game) Info() engine.Info {
	return engine.Info{
		ID:          "4enraya",
		Name:        "4 en raya",
		Emoji:       "🔴",
		Description: "Por turnos, deja caer fichas en el tablero; gana quien junte cuatro en línea.",
		MinPlayers:  1,
		MaxPlayers:  2,
		Lobby:       true,
		Turns:       true,
		Bots:        true,
		Options:     duel.Options,
	}
}

func (g Game) Start(m *engine.Match, env *engine.Env) error {
	s := &state{Seats: duel.Seats(m, env)}
	duel.Turn(m, env)
	g.botTurn(m, s, env)
	return engine.Save(m, s)
}

func (g Game) Act(m *engine.Match, userID string, move json.RawMessage, env *engine.Env) error {
	s, err := engine.Load[state](m)
	if err != nil {
		return err
	}
	if m.Phase != duel.PhaseTurn || s.Seats[s.Turn] != userID {
		return engine.Refusal("No es tu turno")
	}
	mv, err := engine.Decode[struct {
		Column int `json:"column"`
	}](move)
	if err != nil {
		return err
	}
	if mv.Column < 0 || mv.Column >= cols || !s.Board.open(mv.Column) {
		return engine.Refusal("Esa columna está llena")
	}
	g.play(m, s, mv.Column, env)
	g.botTurn(m, s, env)
	return engine.Save(m, s)
}

// play drops a disc of the player on turn and ends the match or passes the turn.
func (Game) play(m *engine.Match, s *state, column int, env *engine.Env) {
	r := s.Board.drop(column, s.Turn+1)
	s.Moves = append(s.Moves, column)
	if line := s.Board.line(r, column); line != nil {
		s.Line = line
		duel.Finish(m, env, s.Seats, s.Turn, fmt.Sprintf("🏆 Gana %s en %d jugadas.\n\n%s", env.Name(s.Seats[s.Turn]), len(s.Moves), s.text()))
		return
	}
	if s.Board.full() {
		duel.Finish(m, env, s.Seats, -1, "🤝 Empate: tablero lleno.\n\n"+s.text())
		return
	}
	s.Turn = 1 - s.Turn
	duel.Turn(m, env)
}

func (g Game) botTurn(m *engine.Match, s *state, env *engine.Env) {
	if m.Over() || !engine.IsBot(s.Seats[s.Turn]) {
		return
	}
	g.play(m, s, botMove(s.Board, s.Turn+1, duel.Hard(m, g.Info()), env.Rand), env)
}

func (Game) Timeout(m *engine.Match, env *engine.Env) error {
	s, err := engine.Load[state](m)
	if err != nil {
		return err
	}
	duel.Timeout(m, env, s.Seats, s.Turn, s.text())
	return engine.Save(m, s)
}

func (Game) Notice(m *engine.Match, env *engine.Env) string {
	s, err := engine.Load[state](m)
	if err != nil {
		return ""
	}
	return duel.Notice(m, env, s.Seats, s.Turn)
}

type view struct {
	Players [2]string `json:"players"`
	Turn    int       `json:"turn"`
	Mine    int       `json:"mine"`
	Cells   board     `json:"cells"`
	Rows    int       `json:"rows"`
	Cols    int       `json:"cols"`
	Last    int       `json:"last"`
	Line    [][2]int  `json:"line,omitempty"`
	Winner  int       `json:"winner"`
}

func (Game) View(m *engine.Match, userID string, env *engine.Env) any {
	s, err := engine.Load[state](m)
	if err != nil {
		return nil
	}
	v := view{Players: duel.Names(env, s.Seats), Turn: s.Turn, Mine: duel.SeatOf(s.Seats, userID), Cells: s.Board, Rows: rows, Cols: cols, Last: -1, Line: s.Line, Winner: -1}
	if len(s.Moves) > 0 {
		v.Last = s.Moves[len(s.Moves)-1]
	}
	if s.Line != nil {
		v.Winner = s.Board.at(s.Line[0][0], s.Line[0][1]) - 1
	}
	return v
}

// text draws the board for the thread, which the mobile apps can show.
func (s *state) text() string {
	discs := []string{"⚪", "🔴", "🟡"}
	var b strings.Builder
	for r := range rows {
		for c := range cols {
			b.WriteString(discs[s.Board.at(r, c)])
		}
		b.WriteString("\n")
	}
	return b.String()
}
