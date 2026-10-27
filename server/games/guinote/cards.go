package guinote

import (
	"math/rand/v2"
	"slices"
	"strconv"
)

// A card is its suit letter and number, "o1" for the ace of oros, as the
// pictures in server/cards are named.
type card string

var (
	suits = []byte{'o', 'c', 'e', 'b'}
	// Numbers from strongest to weakest.
	strength = []int{1, 3, 12, 10, 11, 7, 6, 5, 4, 2}
	points   = map[int]int{1: 11, 3: 10, 12: 4, 10: 3, 11: 2}
)

const (
	king   = 12
	knight = 11
	jack   = 10
	seven  = 7
)

var suitNames = map[byte]string{'o': "oros", 'c': "copas", 'e': "espadas", 'b': "bastos"}

func (c card) suit() byte { return c[0] }

func (c card) number() int {
	n, _ := strconv.Atoi(string(c[1:]))
	return n
}

func (c card) points() int { return points[c.number()] }

// rank is 0 for the strongest number and 9 for the weakest.
func (c card) rank() int { return slices.Index(strength, c.number()) }

func makeCard(suit byte, n int) card { return card(string(suit) + strconv.Itoa(n)) }

func newDeck(rnd *rand.Rand) []card {
	deck := make([]card, 0, 40)
	for _, s := range suits {
		for _, n := range strength {
			deck = append(deck, makeCard(s, n))
		}
	}
	rnd.Shuffle(len(deck), func(i, j int) { deck[i], deck[j] = deck[j], deck[i] })
	return deck
}

// beats tells whether a beats b, with lead the suit led and trump the trump suit.
func beats(a, b card, lead, trump byte) bool {
	switch {
	case a.suit() == b.suit():
		return a.rank() < b.rank()
	case a.suit() == trump:
		return true
	case b.suit() == trump:
		return false
	default:
		return a.suit() == lead && b.suit() != lead
	}
}

type play struct {
	Seat int  `json:"seat"`
	Card card `json:"card"`
}

// winning is the play that wins the trick so far.
func winning(trick []play, trump byte) play {
	best := trick[0]
	for _, p := range trick[1:] {
		if beats(p.Card, best.Card, trick[0].Card.suit(), trump) {
			best = p
		}
	}
	return best
}

// legal are the cards of hand that seat may play on trick. Before the deck runs
// out anything goes; in the arrastre (official rules of the Asociación
// Aragonesa, 7.11) a player must follow suit and beat the best card of an
// opponent if they can, or else trump it; with the partner winning, any card of
// the suit, or any card at all, will do.
func legal(hand []card, trick []play, seat int, trump byte, arrastre bool) []card {
	if !arrastre || len(trick) == 0 {
		return hand
	}
	lead := trick[0].Card.suit()
	best := winning(trick, trump)
	partnerWins := best.Seat%2 == seat%2
	followers := filter(hand, func(c card) bool { return c.suit() == lead })
	if len(followers) > 0 {
		if partnerWins || best.Card.suit() != lead {
			return followers
		}
		if higher := filter(followers, func(c card) bool { return beats(c, best.Card, lead, trump) }); len(higher) > 0 {
			return higher
		}
		return followers
	}
	if partnerWins {
		return hand
	}
	if higher := filter(hand, func(c card) bool { return c.suit() == trump && beats(c, best.Card, lead, trump) }); len(higher) > 0 {
		return higher
	}
	return hand
}

func filter(cards []card, keep func(card) bool) []card {
	var out []card
	for _, c := range cards {
		if keep(c) {
			out = append(out, c)
		}
	}
	return out
}

func trickPoints(trick []play) int {
	total := 0
	for _, p := range trick {
		total += p.Card.points()
	}
	return total
}

// sortHand orders a hand by suit, trumps first, and then by strength.
func sortHand(hand []card, trump byte) {
	order := func(c card) int {
		s := slices.Index(suits, c.suit()) + 1
		if c.suit() == trump {
			s = 0
		}
		return s*10 + c.rank()
	}
	slices.SortFunc(hand, func(a, b card) int { return order(a) - order(b) })
}
