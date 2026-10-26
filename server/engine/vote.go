package engine

// Ballot is a pick-one vote: voter id → chosen option.
type Ballot map[string]string

// Tally counts the votes of each option.
func (b Ballot) Tally() map[string]int {
	counts := map[string]int{}
	for _, option := range b {
		counts[option]++
	}
	return counts
}

// Rounds is how many rounds a short game plays in each mode.
func Rounds(mode Mode, live, async int) int {
	if mode == Live {
		return live
	}
	return async
}

// TookPart gives userID an entry in scores, with no points, so the ranking
// counts them as a participant.
func TookPart(scores map[string]int, userID string) {
	if _, ok := scores[userID]; !ok {
		scores[userID] = 0
	}
}
