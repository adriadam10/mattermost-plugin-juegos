package guinote

import (
	"slices"

	"github.com/adriadam10/mattermost-plugin-juegos/server/engine"
)

type seatView struct {
	Name  string `json:"name"`
	Bot   bool   `json:"bot"`
	Cards int    `json:"cards"`
}

type view struct {
	Seats   [seats]seatView `json:"seats"`
	Mine    int             `json:"mine"`
	Turn    int             `json:"turn"`
	Dealer  int             `json:"dealer"`
	Hand    []card          `json:"hand"`
	Legal   []card          `json:"legal"`
	Trick   []play          `json:"trick"`
	Last    []play          `json:"last"`
	Muestra card            `json:"muestra"`
	Trump   string          `json:"trump"`
	Deck    int             `json:"deck"`
	Games   [2]int          `json:"games"`
	Vueltas bool            `json:"vueltas"`
	// Points of the first deal, known in the vueltas.
	Carry   [2]int   `json:"carry"`
	Cantes  [2]int   `json:"cantes"`
	Log     []string `json:"log"`
	History []string `json:"history"`
	// What the player may do besides playing a card.
	CanteSuits []string `json:"cante_suits"`
	Tute       bool     `json:"tute"`
	Change     bool     `json:"change"`
}

func (Game) View(m *engine.Match, userID string, env *engine.Env) any {
	s, err := engine.Load[state](m)
	if err != nil {
		return nil
	}
	h := &s.H
	v := view{
		Mine: slices.Index(s.Seats[:], userID), Turn: h.Turn, Dealer: s.Dealer, Trick: h.Trick, Last: h.Last,
		Muestra: h.Muestra, Trump: suitNames[h.Trump], Deck: len(h.Deck), Games: s.Games, Vueltas: s.Vueltas,
		Carry: s.Carry, Cantes: h.Cantes, Log: s.Log, History: s.History,
	}
	for i, id := range s.Seats {
		v.Seats[i] = seatView{Name: env.Name(id), Bot: engine.IsBot(id), Cards: len(h.Hands[i])}
	}
	if v.Mine < 0 || m.Over() {
		return v
	}
	v.Hand = slices.Clone(h.Hands[v.Mine])
	sortHand(v.Hand, h.Trump)
	if h.Turn == v.Mine {
		v.Legal = legal(h.Hands[v.Mine], h.Trick, v.Mine, h.Trump, h.arrastre())
	}
	if s.mayTalk(m, v.Mine) {
		for _, suit := range s.canteSuits(v.Mine) {
			v.CanteSuits = append(v.CanteSuits, string(suit))
		}
		v.Tute = s.tuteOf(v.Mine) != ""
		v.Change = s.canChange(m, v.Mine)
	}
	return v
}
