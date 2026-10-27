// Package guinote is the Aragonese guiñote: four players in two pairs, sitting
// crosswise, with the official rules of the Asociación Aragonesa de Guiñote
// (cantes, change of the seven, arrastre, ten for the last trick, 101 or
// vueltas) plus the tute. A match is a coto: best of three games. Bots take the
// empty seats.
package guinote

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/adriadam10/mattermost-plugin-juegos/server/engine"
	"github.com/adriadam10/mattermost-plugin-juegos/server/games/duel"
)

const (
	seats     = 4
	handSize  = 6
	toWin     = 101
	minPoints = 30
	lastTrick = 10
	// Games a pair must win to take the coto.
	gamesToWin = 2
	logSize    = 6
)

type Game struct{}

// hand is one deal, from the shuffle to the last trick.
type hand struct {
	Deck    []card        `json:"deck"`
	Muestra card          `json:"muestra"`
	Trump   byte          `json:"trump"`
	Hands   [seats][]card `json:"hands"`
	Trick   []play        `json:"trick"`
	Turn    int           `json:"turn"`
	// Last trick and who won it; -1 before the first.
	Last       []play `json:"last"`
	LastWinner int    `json:"last_winner"`
	Played     []card `json:"played"`
	// Card points of each pair, with the ten of the last trick, and their cantes.
	Won    [2]int `json:"won"`
	Cantes [2]int `json:"cantes"`
	Canted []byte `json:"canted"`
}

type state struct {
	Seats  [seats]string `json:"seats"`
	Dealer int           `json:"dealer"`
	Games  [2]int        `json:"games"`
	// In the vueltas, the points of the first deal count too.
	Vueltas bool     `json:"vueltas"`
	Carry   [2]int   `json:"carry"`
	H       hand     `json:"hand"`
	Log     []string `json:"log"`
	History []string `json:"history"`
}

func (Game) Info() engine.Info {
	return engine.Info{
		ID:          "guinote",
		Name:        "Guiñote",
		Emoji:       "🃏",
		Description: "Por parejas, con cantes, arrastre y vueltas; gana el coto quien se lleve dos juegos. Los huecos los juega el bot. Parejas por orden de llegada: 1.º y 3.º contra 2.º y 4.º.",
		MinPlayers:  1,
		MaxPlayers:  seats,
		Lobby:       true,
		Turns:       true,
		Bots:        true,
		Options:     []engine.Option{duel.Level},
	}
}

func (g Game) Start(m *engine.Match, env *engine.Env) error {
	s := &state{Dealer: env.Rand.IntN(seats)}
	copy(s.Seats[:], engine.Seats(m, seats))
	s.deal(env)
	duel.Turn(m, env)
	g.botTurns(m, s, env)
	return engine.Save(m, s)
}

func (s *state) deal(env *engine.Env) {
	deck := newDeck(env.Rand)
	h := hand{LastWinner: -1, Turn: (s.Dealer + 1) % seats}
	for i := range seats {
		seat := (s.Dealer + 1 + i) % seats
		h.Hands[seat] = deck[i*handSize : (i+1)*handSize]
	}
	h.Muestra = deck[seats*handSize]
	h.Trump = h.Muestra.suit()
	h.Deck = deck[seats*handSize+1:]
	s.H = h
	what := "Reparto"
	if s.Vueltas {
		what = "Reparto de vueltas"
	}
	s.log(fmt.Sprintf("%s: pinta %s (muestra %s)", what, suitNames[h.Trump], h.Muestra))
}

func (s *state) log(line string) {
	s.Log = append(s.Log, line)
	if len(s.Log) > logSize {
		s.Log = s.Log[len(s.Log)-logSize:]
	}
}

func (h *hand) arrastre() bool { return len(h.Deck) == 0 && h.Muestra == "" }

func (g Game) Act(m *engine.Match, userID string, move json.RawMessage, env *engine.Env) error {
	s, err := engine.Load[state](m)
	if err != nil {
		return err
	}
	seat := slices.Index(s.Seats[:], userID)
	if seat < 0 || m.Phase != duel.PhaseTurn {
		return engine.Refusal("No juegas en esta partida")
	}
	mv, err := engine.Decode[struct {
		Card    card   `json:"card"`
		Cante   string `json:"cante"`
		Tute    bool   `json:"tute"`
		Cambiar bool   `json:"cambiar"`
	}](move)
	if err != nil {
		return err
	}
	name := env.Name(userID)
	switch {
	case mv.Tute:
		err = s.tute(m, seat, name, env)
	case mv.Cambiar:
		err = s.change(seat, name)
	case mv.Cante != "":
		err = s.cante(m, seat, mv.Cante[0], name, env)
	default:
		err = s.play(m, seat, mv.Card, env)
	}
	if err != nil {
		return err
	}
	g.botTurns(m, s, env)
	return engine.Save(m, s)
}

// mayTalk tells whether seat may cante or change the seven now: their pair won
// the last trick and no opponent has played to the next one (rules 7.12, 7.13).
func (s *state) mayTalk(m *engine.Match, seat int) bool {
	h := &s.H
	if m.Over() || h.LastWinner < 0 || h.LastWinner%2 != seat%2 {
		return false
	}
	for _, p := range h.Trick {
		if p.Seat%2 != seat%2 {
			return false
		}
	}
	return true
}

// canteSuits are the suits seat holds king and jack of, not sung yet.
func (s *state) canteSuits(seat int) []byte {
	var out []byte
	for _, suit := range suits {
		hand := s.H.Hands[seat]
		if !slices.Contains(s.H.Canted, suit) && slices.Contains(hand, makeCard(suit, king)) && slices.Contains(hand, makeCard(suit, jack)) {
			out = append(out, suit)
		}
	}
	// The forty first.
	slices.SortStableFunc(out, func(a, b byte) int {
		if a == s.H.Trump {
			return -1
		}
		if b == s.H.Trump {
			return 1
		}
		return 0
	})
	return out
}

func (s *state) cante(m *engine.Match, seat int, suit byte, name string, env *engine.Env) error {
	if !s.mayTalk(m, seat) {
		return engine.Refusal("Solo se canta después de que tu pareja gane una baza, antes de que tire el contrario")
	}
	if !slices.Contains(s.canteSuits(seat), suit) {
		return engine.Refusal("Para cantar hace falta el rey y la sota de ese palo")
	}
	value, what := 20, "las 20 en "+suitNames[suit]
	if suit == s.H.Trump {
		value, what = 40, "las 40"
	}
	s.H.Canted = append(s.H.Canted, suit)
	s.H.Cantes[seat%2] += value
	s.log(fmt.Sprintf("%s canta %s", name, what))
	s.checkVueltas(m, env)
	return nil
}

// tute is four kings or four knights in hand: the pair wins the game outright.
func (s *state) tute(m *engine.Match, seat int, name string, env *engine.Env) error {
	if !s.mayTalk(m, seat) {
		return engine.Refusal("Solo se canta después de que tu pareja gane una baza, antes de que tire el contrario")
	}
	what := s.tuteOf(seat)
	if what == "" {
		return engine.Refusal("Para el tute hacen falta los cuatro reyes o los cuatro caballos")
	}
	s.log(fmt.Sprintf("%s canta tute de %s", name, what))
	s.winGame(m, seat%2, fmt.Sprintf("tute de %s de %s", what, name), env)
	return nil
}

func (s *state) tuteOf(seat int) string {
	for _, n := range []int{king, knight} {
		count := 0
		for _, c := range s.H.Hands[seat] {
			if c.number() == n {
				count++
			}
		}
		if count == len(suits) {
			return map[int]string{king: "reyes", knight: "caballos"}[n]
		}
	}
	return ""
}

func (s *state) canChange(m *engine.Match, seat int) bool {
	h := &s.H
	return h.Muestra != "" && len(h.Deck) > 0 && s.mayTalk(m, seat) && slices.Contains(h.Hands[seat], makeCard(h.Trump, seven)) && h.Muestra.number() != seven
}

// change swaps the seven of trumps for the muestra (rule 7.13).
func (s *state) change(seat int, name string) error {
	h := &s.H
	seven := makeCard(h.Trump, seven)
	if h.Muestra == "" || len(h.Deck) == 0 || !slices.Contains(h.Hands[seat], seven) {
		return engine.Refusal("Para cambiar hace falta el 7 de pinta y que quede muestra")
	}
	if h.LastWinner < 0 || h.LastWinner%2 != seat%2 {
		return engine.Refusal("Solo se cambia después de que tu pareja gane una baza")
	}
	i := slices.Index(h.Hands[seat], seven)
	h.Hands[seat][i], h.Muestra = h.Muestra, seven
	s.log(fmt.Sprintf("%s cambia el 7 por %s", name, h.Hands[seat][i]))
	return nil
}

func (s *state) play(m *engine.Match, seat int, c card, env *engine.Env) error {
	h := &s.H
	if seat != h.Turn {
		return engine.Refusal("No es tu turno")
	}
	if !slices.Contains(h.Hands[seat], c) {
		return engine.Refusal("No tienes esa carta")
	}
	if !slices.Contains(legal(h.Hands[seat], h.Trick, seat, h.Trump, h.arrastre()), c) {
		return engine.Refusal("En el arrastre hay que asistir al palo y montar (o fallar) si se puede")
	}
	h.Hands[seat] = slices.DeleteFunc(slices.Clone(h.Hands[seat]), func(x card) bool { return x == c })
	h.Trick = append(h.Trick, play{Seat: seat, Card: c})
	h.Played = append(h.Played, c)
	if len(h.Trick) < seats {
		h.Turn = (seat + 1) % seats
		duel.Turn(m, env)
		return nil
	}
	s.endTrick(m, env)
	return nil
}

func (s *state) endTrick(m *engine.Match, env *engine.Env) {
	h := &s.H
	w := winning(h.Trick, h.Trump)
	pair := w.Seat % 2
	h.Won[pair] += trickPoints(h.Trick)
	h.Last, h.Trick, h.LastWinner, h.Turn = h.Trick, nil, w.Seat, w.Seat
	if len(h.Hands[0]) == 0 {
		h.Won[pair] += lastTrick
		s.log(fmt.Sprintf("Las diez de últimas, para %s", env.Name(s.Seats[w.Seat])))
		if !s.checkVueltas(m, env) {
			s.endHand(m, env)
		}
		return
	}
	for i := range seats {
		seat := (w.Seat + i) % seats
		switch {
		case len(h.Deck) > 0:
			h.Hands[seat] = append(h.Hands[seat], h.Deck[0])
			h.Deck = h.Deck[1:]
		case h.Muestra != "":
			h.Hands[seat] = append(h.Hands[seat], h.Muestra)
			h.Muestra = ""
		}
	}
	if !s.checkVueltas(m, env) {
		duel.Turn(m, env)
	}
}

func (s *state) total(pair int) int { return s.Carry[pair] + s.H.Won[pair] + s.H.Cantes[pair] }

// checkVueltas ends the vueltas as soon as the pair that won the last trick
// passes 100 (rule 4.3). It tells whether the game ended.
func (s *state) checkVueltas(m *engine.Match, env *engine.Env) bool {
	if !s.Vueltas || s.H.LastWinner < 0 {
		return false
	}
	pair := s.H.LastWinner % 2
	if s.total(pair) < toWin {
		return false
	}
	s.winGame(m, pair, fmt.Sprintf("de vueltas, %d a %d", s.total(pair), s.total(1-pair)), env)
	return true
}

// endHand counts the points when the last trick is played.
func (s *state) endHand(m *engine.Match, env *engine.Env) {
	h := &s.H
	last := h.LastWinner % 2
	score := fmt.Sprintf("%d a %d", s.total(0), s.total(1))
	if s.Vueltas {
		// Only reached when the last trick did not take its pair past 100.
		best := 0
		if s.total(1) > s.total(0) {
			best = 1
		}
		s.winGame(m, best, "de vueltas, "+score, env)
		return
	}
	// A pair wins with 101 counting the cantes, but needs 30 without them;
	// with both past 100, the one with the last trick (rule 4.3).
	ok := [2]bool{}
	for p := range ok {
		ok[p] = s.total(p) >= toWin && h.Won[p] >= minPoints
	}
	switch {
	case ok[last]:
		s.winGame(m, last, "de buenas, "+score, env)
	case ok[1-last]:
		s.winGame(m, 1-last, "de buenas, "+score, env)
	default:
		s.log(fmt.Sprintf("Nadie llega a %d (%s): vueltas", toWin, score))
		s.Carry = [2]int{s.total(0), s.total(1)}
		s.Vueltas = true
		s.Dealer = h.LastWinner
		s.deal(env)
		duel.Turn(m, env)
	}
}

func (s *state) pairName(env *engine.Env, pair int) string {
	return env.Name(s.Seats[pair]) + " y " + env.Name(s.Seats[pair+2])
}

func (s *state) winGame(m *engine.Match, pair int, how string, env *engine.Env) {
	s.Games[pair]++
	line := fmt.Sprintf("Juego %d para %s (%s)", len(s.History)+1, s.pairName(env, pair), how)
	s.History = append(s.History, line)
	s.log(line)
	if s.Games[pair] >= gamesToWin {
		s.finish(m, pair, fmt.Sprintf("🏆 Coto para **%s**, %d a %d.", s.pairName(env, pair), s.Games[pair], s.Games[1-pair]), env)
		return
	}
	s.Vueltas, s.Carry = false, [2]int{}
	if s.H.LastWinner >= 0 {
		// Deals whoever made the last trick.
		s.Dealer = s.H.LastWinner
	} else {
		s.Dealer = (s.Dealer + 1) % seats
	}
	s.deal(env)
	duel.Turn(m, env)
}

// finish ends the coto. It only counts for the ranking with people on both sides.
func (s *state) finish(m *engine.Match, winner int, headline string, env *engine.Env) {
	scores := map[string]int{}
	humans := [2]bool{}
	for seat, id := range s.Seats {
		if engine.IsBot(id) {
			continue
		}
		humans[seat%2] = true
		scores[id] = 0
		if seat%2 == winner {
			scores[id] = 1
		}
	}
	env.Finish(m, scores, headline+"\n\n"+strings.Join(s.History, "\n"))
	m.Result.Unranked = !humans[0] || !humans[1]
}

func (g Game) Timeout(m *engine.Match, env *engine.Env) error {
	s, err := engine.Load[state](m)
	if err != nil {
		return err
	}
	seat := s.H.Turn
	if engine.IsBot(s.Seats[seat]) {
		g.botTurns(m, s, env)
	} else {
		s.finish(m, 1-seat%2, fmt.Sprintf("⏰ %s no ha jugado a tiempo: el coto es para %s.", env.Name(s.Seats[seat]), s.pairName(env, 1-seat%2)), env)
	}
	return engine.Save(m, s)
}

func (Game) Notice(m *engine.Match, env *engine.Env) string {
	s, err := engine.Load[state](m)
	if err != nil || m.Phase != duel.PhaseTurn || engine.IsBot(s.Seats[s.H.Turn]) {
		return ""
	}
	return "🃏 Te toca, " + env.Name(s.Seats[s.H.Turn])
}
