package engine

import (
	"math/rand/v2"
	"sort"
	"strings"
	"time"
)

// Services is what the plugin lends the games: content, GIFs, blobs and names.
type Services interface {
	// Pick returns up to n different texts of a content kind (frases de absurdas...).
	Pick(kind string, n int) []string
	// Gif returns the URL of a random GIF for tag.
	Gif(tag string) (string, error)
	// PutBlob stores data under a new unguessable key and returns the key.
	PutBlob(data []byte) (string, error)
	Username(userID string) string
}

// Env is passed to every game call.
type Env struct {
	Services
	Now  time.Time
	Rand *rand.Rand
	// Phase length for a weight of 1 in this match's mode.
	Base time.Duration
}

// SetPhase starts a phase that lasts weight times the base length of the mode.
func (e *Env) SetPhase(m *Match, phase string, weight float64) {
	m.Phase = phase
	m.Deadline = e.Now.Add(time.Duration(weight * float64(e.Base))).UnixMilli()
}

// PutBlob stores data and remembers it in m, so it goes with the match.
func (e *Env) PutBlob(m *Match, data []byte) (string, error) {
	key, err := e.Services.PutBlob(data)
	if err != nil {
		return "", err
	}
	m.Blobs = append(m.Blobs, key)
	return key, nil
}

// Finish ends the match with the game points of each participant.
func (e *Env) Finish(m *Match, scores map[string]int, summary string, parts ...Part) {
	m.Phase = PhaseFinished
	m.Deadline = 0
	m.Result = &Result{Scores: scores, Summary: summary, Parts: parts}
}

// Mentions turns user ids into "@a @b"; bots are named, not mentioned.
func (e *Env) Mentions(ids []string) string {
	parts := make([]string, len(ids))
	for i, id := range ids {
		parts[i] = e.Name(id)
	}
	return strings.Join(parts, " ")
}

// Name is "@user" for a user and "🤖 bot" for a bot seat.
func (e *Env) Name(id string) string {
	if IsBot(id) {
		return "🤖 bot"
	}
	return "@" + e.Username(id)
}

// Places gives ranking points by place: ties share it. Only players who
// scored or took part (present in scores) get anything.
var Places = []int{5, 3, 2}

// ParticipationPoints is what everyone else who played gets.
const ParticipationPoints = 1

// Award turns game scores into ranking points.
func Award(scores map[string]int) map[string]int {
	ids := make([]string, 0, len(scores))
	for id := range scores {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return scores[ids[i]] > scores[ids[j]] })
	awarded := make(map[string]int, len(ids))
	place := 0
	for i, id := range ids {
		if i > 0 && scores[id] < scores[ids[i-1]] {
			place = i
		}
		points := ParticipationPoints
		if place < len(Places) && scores[id] > 0 {
			points = Places[place]
		}
		awarded[id] = points
	}
	return awarded
}

// Ranked returns the ids of scores from best to worst, ties by id for a stable order.
func Ranked(scores map[string]int) []string {
	ids := make([]string, 0, len(scores))
	for id := range scores {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool {
		if scores[ids[i]] != scores[ids[j]] {
			return scores[ids[i]] > scores[ids[j]]
		}
		return ids[i] < ids[j]
	})
	return ids
}
