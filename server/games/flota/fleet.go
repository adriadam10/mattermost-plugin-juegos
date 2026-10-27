package flota

import (
	"math/rand/v2"
	"slices"
)

const size = 10

// fleet is the Spanish fleet: a carrier of four, two of three, three of two and four of one.
var fleet = []int{4, 3, 3, 2, 2, 2, 1, 1, 1, 1}

// What a cell of a board looks like to the player who shoots at it.
const (
	unknown = iota
	water
	hit
	sunk
)

type cell [2]int

func inside(r, c int) bool { return r >= 0 && r < size && c >= 0 && c < size }

// validFleet tells whether ships are the fleet: straight, inside, and not
// touching each other, not even at a corner.
func validFleet(ships [][]cell) bool {
	sizes := make([]int, len(ships))
	for i, s := range ships {
		sizes[i] = len(s)
	}
	slices.Sort(sizes)
	want := slices.Clone(fleet)
	slices.Sort(want)
	if !slices.Equal(sizes, want) {
		return false
	}
	return validPartial(ships)
}

// straight tells whether s is a row or column of adjacent cells inside the board.
func straight(s []cell) bool {
	s = slices.Clone(s)
	slices.SortFunc(s, func(a, b cell) int { return (a[0]*size + a[1]) - (b[0]*size + b[1]) })
	for i, p := range s {
		if !inside(p[0], p[1]) {
			return false
		}
		if i == 0 {
			continue
		}
		dr, dc := p[0]-s[0][0], p[1]-s[0][1]
		if dr != 0 && dc != 0 || dr+dc != i {
			return false
		}
	}
	return true
}

// randomFleet places the fleet at random, biggest ships first.
func randomFleet(rnd *rand.Rand) [][]cell {
	for {
		var ships [][]cell
		for _, n := range fleet {
			placed := false
			for try := 0; try < 200 && !placed; try++ {
				ship := line(rnd.IntN(size), rnd.IntN(size), n, rnd.IntN(2) == 0)
				if validPartial(append(ships, ship)) {
					ships = append(ships, ship)
					placed = true
				}
			}
			if !placed {
				break
			}
		}
		if len(ships) == len(fleet) {
			return ships
		}
	}
}

func line(r, c, n int, across bool) []cell {
	s := make([]cell, n)
	for i := range s {
		if across {
			s[i] = cell{r, c + i}
		} else {
			s[i] = cell{r + i, c}
		}
	}
	return s
}

// validPartial checks part of a fleet: straight ships that do not touch.
func validPartial(ships [][]cell) bool {
	var taken [size][size]int
	for i, s := range ships {
		if !straight(s) {
			return false
		}
		for _, p := range s {
			for dr := -1; dr <= 1; dr++ {
				for dc := -1; dc <= 1; dc++ {
					r, c := p[0]+dr, p[1]+dc
					if inside(r, c) && taken[r][c] != 0 && taken[r][c] != i+1 {
						return false
					}
				}
			}
			taken[p[0]][p[1]] = i + 1
		}
	}
	return true
}

// botShot picks where the bot shoots on a board it sees as shots. The easy bot
// shoots anywhere new; the hard one counts, for every cell, how many ways the
// ships still afloat (left) fit through it, and leans hard on cells next to hits.
func botShot(shots *[size * size]int, left []int, hard bool, rnd *rand.Rand) cell {
	var open []cell
	for i, v := range shots {
		if v == unknown {
			open = append(open, cell{i / size, i % size})
		}
	}
	if !hard {
		return open[rnd.IntN(len(open))]
	}
	var weight [size * size]int
	for _, n := range left {
		for r := range size {
			for c := range size {
				for _, across := range []bool{true, false} {
					s := line(r, c, n, across)
					fits, hits := true, 0
					for _, p := range s {
						if !inside(p[0], p[1]) {
							fits = false
							break
						}
						switch shots[p[0]*size+p[1]] {
						case water, sunk:
							fits = false
						case hit:
							hits++
						}
					}
					if !fits {
						continue
					}
					for _, p := range s {
						weight[p[0]*size+p[1]] += 1 + hits*50
					}
					if n == 1 {
						break
					}
				}
			}
		}
	}
	best := []cell{}
	bestWeight := -1
	for _, p := range open {
		w := weight[p[0]*size+p[1]]
		switch {
		case w > bestWeight:
			best, bestWeight = []cell{p}, w
		case w == bestWeight:
			best = append(best, p)
		}
	}
	return best[rnd.IntN(len(best))]
}
