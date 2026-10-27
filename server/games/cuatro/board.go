package cuatro

import "math/rand/v2"

const (
	rows  = 6
	cols  = 7
	empty = 0
)

// board holds 0 for an empty cell and seat+1 for a disc; row 0 is the top.
type board [rows * cols]int

func (b *board) at(r, c int) int { return b[r*cols+c] }

// drop puts a disc of who in column c and returns its row, or -1 if the column is full.
func (b *board) drop(c, who int) int {
	for r := rows - 1; r >= 0; r-- {
		if b[r*cols+c] == empty {
			b[r*cols+c] = who
			return r
		}
	}
	return -1
}

func (b *board) undo(c int) {
	for r := range rows {
		if b[r*cols+c] != empty {
			b[r*cols+c] = empty
			return
		}
	}
}

func (b *board) open(c int) bool { return b[c] == empty }

func (b *board) full() bool {
	for c := range cols {
		if b.open(c) {
			return false
		}
	}
	return true
}

var directions = [][2]int{{0, 1}, {1, 0}, {1, 1}, {1, -1}}

// line returns the cells of a four in a row through (r, c), or nil.
func (b *board) line(r, c int) [][2]int {
	who := b.at(r, c)
	for _, d := range directions {
		cells := [][2]int{{r, c}}
		for _, sign := range []int{1, -1} {
			for i := 1; ; i++ {
				rr, cc := r+sign*i*d[0], c+sign*i*d[1]
				if rr < 0 || rr >= rows || cc < 0 || cc >= cols || b.at(rr, cc) != who {
					break
				}
				cells = append(cells, [2]int{rr, cc})
			}
		}
		if len(cells) >= 4 {
			return cells
		}
	}
	return nil
}

// Columns from the centre out: better moves first prunes the search more.
var order = []int{3, 2, 4, 1, 5, 0, 6}

// botMove picks a column for who. The easy bot only takes a win it sees; the
// hard one searches hardDepth moves ahead.
func botMove(b board, who int, hard bool, rnd *rand.Rand) int {
	for _, c := range order {
		if !b.open(c) {
			continue
		}
		r := b.drop(c, who)
		won := b.line(r, c) != nil
		b.undo(c)
		if won {
			return c
		}
	}
	if !hard {
		var open []int
		for c := range cols {
			if b.open(c) {
				open = append(open, c)
			}
		}
		return open[rnd.IntN(len(open))]
	}
	best, bestScore := -1, -winScore*2
	for _, c := range order {
		if !b.open(c) {
			continue
		}
		r := b.drop(c, who)
		score := -negamax(&b, r, c, 3-who, hardDepth-1, -winScore*2, -bestScore)
		b.undo(c)
		if score > bestScore {
			best, bestScore = c, score
		}
	}
	return best
}

const (
	hardDepth = 7
	winScore  = 1_000_000
)

// negamax scores the board for who, to move, after the other player dropped in (r, c).
func negamax(b *board, r, c, who, depth, alpha, beta int) int {
	if b.line(r, c) != nil {
		// The last move won: sooner losses are worse.
		return -winScore - depth
	}
	if b.full() {
		return 0
	}
	if depth == 0 {
		return evaluate(b, who)
	}
	for _, cc := range order {
		if !b.open(cc) {
			continue
		}
		rr := b.drop(cc, who)
		score := -negamax(b, rr, cc, 3-who, depth-1, -beta, -alpha)
		b.undo(cc)
		if score > alpha {
			alpha = score
		}
		if alpha >= beta {
			break
		}
	}
	return alpha
}

// evaluate counts the open windows of four each player could still fill.
func evaluate(b *board, who int) int {
	score := 0
	for r := range rows {
		for c := range cols {
			for _, d := range directions {
				er, ec := r+3*d[0], c+3*d[1]
				if er < 0 || er >= rows || ec < 0 || ec >= cols {
					continue
				}
				mine, theirs := 0, 0
				for i := range 4 {
					switch b.at(r+i*d[0], c+i*d[1]) {
					case who:
						mine++
					case empty:
					default:
						theirs++
					}
				}
				score += windowScore(mine, theirs) - windowScore(theirs, mine)
			}
		}
	}
	return score
}

func windowScore(mine, theirs int) int {
	if theirs > 0 {
		return 0
	}
	return [...]int{0, 1, 10, 100, 0}[mine]
}
