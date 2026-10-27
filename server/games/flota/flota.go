// Package flota is «Hundir la flota»: each player hides the Spanish fleet on a
// 10×10 board and they shoot in turns; a hit shoots again. Against a challenged
// rival, the first who joins or the bot.
package flota

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/adriadam10/mattermost-plugin-juegos/server/engine"
	"github.com/adriadam10/mattermost-plugin-juegos/server/games/duel"
)

const phasePlace = "colocar"

type Game struct{}

// side is a player's fleet and the shots it took.
type side struct {
	Ships  [][]cell `json:"ships"`
	Placed bool     `json:"placed"`
	// Shots taken, as the other player sees them.
	Shots [size * size]int `json:"shots"`
}

type state struct {
	Seats [2]string `json:"seats"`
	Turn  int       `json:"turn"`
	Sides [2]side   `json:"sides"`
	// Last shot of each seat, -1 before any.
	Last  [2]int `json:"last"`
	Fired [2]int `json:"fired"`
}

func (Game) Info() engine.Info {
	return engine.Info{
		ID:          "flota",
		Name:        "Hundir la flota",
		Emoji:       "🚢",
		Description: "Esconde tu flota y dispara por turnos a la del otro; si tocas, repites. Gana quien hunda todos los barcos.",
		MinPlayers:  1,
		MaxPlayers:  2,
		Lobby:       true,
		Turns:       true,
		Bots:        true,
		Options:     duel.Options,
	}
}

func (Game) Start(m *engine.Match, env *engine.Env) error {
	s := &state{Seats: duel.Seats(m, env), Last: [2]int{-1, -1}}
	for i, id := range s.Seats {
		if engine.IsBot(id) {
			s.Sides[i] = side{Ships: randomFleet(env.Rand), Placed: true}
		}
	}
	m.Round = 1
	env.SetPhase(m, phasePlace, 2)
	return engine.Save(m, s)
}

func (g Game) Act(m *engine.Match, userID string, move json.RawMessage, env *engine.Env) error {
	s, err := engine.Load[state](m)
	if err != nil {
		return err
	}
	seat := duel.SeatOf(s.Seats, userID)
	if seat < 0 {
		return engine.Refusal("No juegas en esta partida")
	}
	switch m.Phase {
	case phasePlace:
		err = g.place(m, s, seat, move, env)
	case duel.PhaseTurn:
		err = g.shoot(m, s, seat, move, env)
	default:
		err = engine.Refusal("Ahora no se puede jugar")
	}
	if err != nil {
		return err
	}
	return engine.Save(m, s)
}

func (g Game) place(m *engine.Match, s *state, seat int, move json.RawMessage, env *engine.Env) error {
	if s.Sides[seat].Placed {
		return engine.Refusal("Ya has colocado tu flota")
	}
	mv, err := engine.Decode[struct {
		Ships [][]cell `json:"ships"`
	}](move)
	if err != nil {
		return err
	}
	if !validFleet(mv.Ships) {
		return engine.Refusal("Esa flota no vale: barcos rectos, sin tocarse y dentro del tablero")
	}
	s.Sides[seat] = side{Ships: mv.Ships, Placed: true}
	if s.Sides[1-seat].Placed {
		g.begin(m, s, env)
	}
	return nil
}

// begin starts the shooting once both fleets are hidden.
func (g Game) begin(m *engine.Match, s *state, env *engine.Env) {
	duel.Turn(m, env)
	g.botTurns(m, s, env)
}

func (g Game) shoot(m *engine.Match, s *state, seat int, move json.RawMessage, env *engine.Env) error {
	if seat != s.Turn {
		return engine.Refusal("No es tu turno")
	}
	mv, err := engine.Decode[struct {
		Row    int `json:"row"`
		Column int `json:"column"`
	}](move)
	if err != nil {
		return err
	}
	if !inside(mv.Row, mv.Column) || s.Sides[1-seat].Shots[mv.Row*size+mv.Column] != unknown {
		return engine.Refusal("Ahí ya has disparado")
	}
	g.fire(m, s, cell{mv.Row, mv.Column}, env)
	g.botTurns(m, s, env)
	return nil
}

// fire is a shot of the player on turn: a hit shoots again, water passes the turn.
func (Game) fire(m *engine.Match, s *state, p cell, env *engine.Env) {
	target := &s.Sides[1-s.Turn]
	i := p[0]*size + p[1]
	s.Last[s.Turn] = i
	s.Fired[s.Turn]++
	target.Shots[i] = water
	for _, ship := range target.Ships {
		if !slices.Contains(ship, p) {
			continue
		}
		target.Shots[i] = hit
		if target.afloat(ship) {
			break
		}
		// Sunk: ships never touch, so the water around it is known too.
		for _, q := range ship {
			target.Shots[q[0]*size+q[1]] = sunk
			for dr := -1; dr <= 1; dr++ {
				for dc := -1; dc <= 1; dc++ {
					r, c := q[0]+dr, q[1]+dc
					if inside(r, c) && target.Shots[r*size+c] == unknown {
						target.Shots[r*size+c] = water
					}
				}
			}
		}
	}
	if len(target.sunkSizes()) == len(fleet) {
		duel.Finish(m, env, s.Seats, s.Turn, fmt.Sprintf("🏆 %s hunde toda la flota de %s en %d disparos.\n\n%s",
			env.Name(s.Seats[s.Turn]), env.Name(s.Seats[1-s.Turn]), s.Fired[s.Turn], s.text(env)))
		return
	}
	if target.Shots[i] == water {
		s.Turn = 1 - s.Turn
	}
	duel.Turn(m, env)
}

func (g Game) botTurns(m *engine.Match, s *state, env *engine.Env) {
	for !m.Over() && m.Phase == duel.PhaseTurn && engine.IsBot(s.Seats[s.Turn]) {
		target := &s.Sides[1-s.Turn]
		g.fire(m, s, botShot(&target.Shots, target.leftSizes(), duel.Hard(m, g.Info()), env.Rand), env)
	}
}

func (sd *side) afloat(ship []cell) bool {
	for _, p := range ship {
		if sd.Shots[p[0]*size+p[1]] == unknown {
			return true
		}
	}
	return false
}

func (sd *side) sunkSizes() []int {
	var sizes []int
	for _, ship := range sd.Ships {
		if sd.Shots[ship[0][0]*size+ship[0][1]] == sunk {
			sizes = append(sizes, len(ship))
		}
	}
	return sizes
}

// leftSizes are the sizes of the ships still afloat, which the shooter knows.
func (sd *side) leftSizes() []int {
	left := slices.Clone(fleet)
	for _, n := range sd.sunkSizes() {
		left = slices.Delete(left, slices.Index(left, n), slices.Index(left, n)+1)
	}
	return left
}

func (g Game) Timeout(m *engine.Match, env *engine.Env) error {
	s, err := engine.Load[state](m)
	if err != nil {
		return err
	}
	if m.Phase == phasePlace {
		// Whoever did not place in time gets a random fleet: nobody loses before the first shot.
		for i := range s.Sides {
			if !s.Sides[i].Placed {
				s.Sides[i] = side{Ships: randomFleet(env.Rand), Placed: true}
			}
		}
		g.begin(m, s, env)
	} else {
		duel.Timeout(m, env, s.Seats, s.Turn, s.text(env))
	}
	return engine.Save(m, s)
}

func (Game) Notice(m *engine.Match, env *engine.Env) string {
	s, err := engine.Load[state](m)
	if err != nil {
		return ""
	}
	if m.Phase == phasePlace {
		return "⚓ Colocad la flota: " + env.Mentions(m.Players)
	}
	return duel.Notice(m, env, s.Seats, s.Turn)
}

type boardView struct {
	Shots [size * size]int `json:"shots"`
	// Ships are only sent to their owner, or to everybody once the match is over.
	Ships [][]cell `json:"ships,omitempty"`
	Left  []int    `json:"left"`
	Last  int      `json:"last"`
}

type view struct {
	Players [2]string    `json:"players"`
	Turn    int          `json:"turn"`
	Mine    int          `json:"mine"`
	Placed  [2]bool      `json:"placed"`
	Boards  [2]boardView `json:"boards"`
	Fleet   []int        `json:"fleet"`
	Size    int          `json:"size"`
}

func (Game) View(m *engine.Match, userID string, env *engine.Env) any {
	s, err := engine.Load[state](m)
	if err != nil {
		return nil
	}
	v := view{Players: duel.Names(env, s.Seats), Turn: s.Turn, Mine: duel.SeatOf(s.Seats, userID), Fleet: fleet, Size: size}
	for i := range s.Sides {
		sd := &s.Sides[i]
		v.Placed[i] = sd.Placed
		v.Boards[i] = boardView{Shots: sd.Shots, Left: sd.leftSizes(), Last: s.Last[1-i]}
		if i == v.Mine || m.Over() {
			v.Boards[i].Ships = sd.Ships
		}
	}
	return v
}

// text draws both boards for the thread, which the mobile apps can show.
func (s *state) text(env *engine.Env) string {
	var b strings.Builder
	for i := range s.Sides {
		fmt.Fprintf(&b, "Flota de %s:\n", env.Name(s.Seats[i]))
		var ships [size * size]bool
		for _, ship := range s.Sides[i].Ships {
			for _, p := range ship {
				ships[p[0]*size+p[1]] = true
			}
		}
		for r := range size {
			for c := range size {
				j := r*size + c
				switch {
				case s.Sides[i].Shots[j] == sunk:
					b.WriteString("🟥")
				case s.Sides[i].Shots[j] == hit:
					b.WriteString("🟧")
				case ships[j]:
					b.WriteString("⬛")
				case s.Sides[i].Shots[j] == water:
					b.WriteString("🔹")
				default:
					b.WriteString("🟦")
				}
			}
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}
	return b.String()
}
