package guinote

import (
	"slices"

	"github.com/adriadam10/mattermost-plugin-juegos/server/engine"
	"github.com/adriadam10/mattermost-plugin-juegos/server/games/duel"
)

// botTurns plays for the bots until a person is on turn or the coto ends. A
// bot always sings and changes the seven when it may.
func (g Game) botTurns(m *engine.Match, s *state, env *engine.Env) {
	hard := duel.Hard(m, g.Info())
	for !m.Over() && engine.IsBot(s.Seats[s.H.Turn]) {
		seat := s.H.Turn
		name := env.Name(s.Seats[seat])
		if s.mayTalk(m, seat) {
			if s.tuteOf(seat) != "" {
				_ = s.tute(m, seat, name, env)
				continue
			}
			if s.canChange(m, seat) {
				_ = s.change(seat, name)
			}
			before := len(s.History)
			for _, suit := range s.canteSuits(seat) {
				_ = s.cante(m, seat, suit, name, env)
			}
			if m.Over() || len(s.History) != before {
				continue
			}
		}
		_ = s.play(m, seat, botCard(&s.H, seat, hard, env), env)
	}
}

func botCard(h *hand, seat int, hard bool, env *engine.Env) card {
	options := legal(h.Hands[seat], h.Trick, seat, h.Trump, h.arrastre())
	if !hard {
		return options[env.Rand.IntN(len(options))]
	}
	if len(h.Trick) == 0 {
		return lead(h, seat, options)
	}
	return follow(h, seat, options)
}

// cost is how much it hurts to give a card away: trumps, points and cards
// that still make a cante are worth keeping.
func cost(h *hand, seat int, c card) int {
	v := c.points()*10 + (9 - c.rank())
	if c.suit() == h.Trump {
		v += 200
	}
	if !h.arrastre() && (c.number() == king || c.number() == jack) && !slices.Contains(h.Canted, c.suit()) {
		other := makeCard(c.suit(), king+jack-c.number())
		if slices.Contains(h.Hands[seat], other) {
			v += 300
		}
	}
	return v
}

func cheapest(h *hand, seat int, cards []card) card {
	return slices.MinFunc(cards, func(a, b card) int { return cost(h, seat, a) - cost(h, seat, b) })
}

// unseenTrumps are the trumps neither played, nor in seat's hand, nor showing as muestra.
func unseenTrumps(h *hand, seat int) []card {
	var out []card
	for _, n := range strength {
		c := makeCard(h.Trump, n)
		if !slices.Contains(h.Played, c) && !slices.Contains(h.Hands[seat], c) && c != h.Muestra {
			out = append(out, c)
		}
	}
	return out
}

// lead: in the arrastre, draw trumps with the best one left; otherwise lead
// something cheap.
func lead(h *hand, seat int, options []card) card {
	if h.arrastre() {
		trumps := filter(options, func(c card) bool { return c.suit() == h.Trump })
		unseen := unseenTrumps(h, seat)
		if len(trumps) > 0 && len(unseen) > 0 {
			best := slices.MinFunc(trumps, func(a, b card) int { return a.rank() - b.rank() })
			if best.rank() < unseen[0].rank() {
				return best
			}
		}
	}
	return cheapest(h, seat, options)
}

// follow: load points on a trick the partner has safely won, win a trick worth
// it with the cheapest card that does, and otherwise give away little.
func follow(h *hand, seat int, options []card) card {
	best := winning(h.Trick, h.Trump)
	last := len(h.Trick) == seats-1
	lead := h.Trick[0].Card.suit()
	if best.Seat%2 == seat%2 {
		safe := last || best.Card.suit() == h.Trump && best.Card.rank() <= 1 || best.Card.suit() != h.Trump && best.Card.rank() == 0
		if safe {
			loads := filter(options, func(c card) bool { return c.suit() != h.Trump })
			if len(loads) > 0 {
				return slices.MaxFunc(loads, func(a, b card) int {
					return (a.points()*10 - cost(h, seat, a)/100) - (b.points()*10 - cost(h, seat, b)/100)
				})
			}
		}
		return cheapest(h, seat, options)
	}
	winners := filter(options, func(c card) bool { return beats(c, best.Card, lead, h.Trump) })
	worth := trickPoints(h.Trick)
	if len(winners) > 0 {
		inSuit := filter(winners, func(c card) bool { return c.suit() != h.Trump })
		if len(inSuit) > 0 && (last || worth > 0 || inSuit[0].rank() <= 1) {
			return cheapest(h, seat, inSuit)
		}
		if worth >= 10 || last && worth >= 4 || h.arrastre() {
			return cheapest(h, seat, winners)
		}
	}
	return cheapest(h, seat, options)
}
