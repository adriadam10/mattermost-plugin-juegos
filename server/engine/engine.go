// Package engine holds what every game shares: the match, its phases and
// deadlines, the lobby and the final scores. A game only says what happens in
// its own phases; see Game.
package engine

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
)

// Mode is how fast a match is played.
type Mode string

const (
	// Live matches last a few minutes, with phases of about a minute.
	Live Mode = "directo"
	// Async matches are played between tasks, with phases of hours.
	Async Mode = "async"
)

// Phases every match goes through; a game adds its own between them.
const (
	PhaseLobby    = "lobby"
	PhaseFinished = "fin"
	PhaseCanceled = "cancelada"
)

// Info describes a game in the menu.
type Info struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Emoji       string `json:"emoji"`
	Description string `json:"description"`
	MinPlayers  int    `json:"min_players"`
	MaxPlayers  int    `json:"max_players"`
	// Lobby games need to know the players before they start (turns, pairings);
	// the others start at once and whoever acts plays.
	Lobby bool `json:"lobby"`
	// Turns games are played in turns: the creator cannot skip a phase, since
	// that would make whoever is on turn lose.
	Turns bool `json:"turns,omitempty"`
	// Bots games fill the empty seats with bots when they start.
	Bots bool `json:"bots,omitempty"`
	// Options the creator picks in the lobby.
	Options []Option `json:"options,omitempty"`
}

// Option is a choice the creator makes before the match starts.
type Option struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	// Empty Choices means the value is a username: the only one who may join.
	Choices []Choice `json:"choices,omitempty"`
	Default string   `json:"default,omitempty"`
}

// Choice is one value of an Option.
type Choice struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// OptionRival is the username of the only player a creator wants to face.
const OptionRival = "rival"

// Game is one game. Every method gets the match to read or change; the engine
// saves it afterwards, so a method that returns an error changes nothing.
type Game interface {
	Info() Info
	// Start sets the first phase. With a lobby, m.Players holds who joined.
	Start(m *Match, env *Env) error
	// Act applies a move by userID in the current phase. A refused move
	// returns a Refusal, which the player sees.
	Act(m *Match, userID string, move json.RawMessage, env *Env) error
	// Timeout runs when the phase deadline passes, or when the creator skips ahead.
	Timeout(m *Match, env *Env) error
	// View is the game part of the match as userID may see it: never the
	// secrets of other players.
	View(m *Match, userID string, env *Env) any
}

// Noticer is optional: in async mode the engine posts Notice in the match
// thread when a phase starts, so people know they have something to do.
type Noticer interface {
	Notice(m *Match, env *Env) string
}

// Match is one match of any game, as stored.
type Match struct {
	ID        string `json:"id"`
	Game      string `json:"game"`
	ChannelID string `json:"channel_id"`
	PostID    string `json:"post_id"`
	CreatorID string `json:"creator_id"`
	Mode      Mode   `json:"mode"`
	Phase     string `json:"phase"`
	Round     int    `json:"round"`
	// When the phase ends, in Unix milliseconds; 0 means no deadline.
	Deadline  int64    `json:"deadline"`
	CreatedAt int64    `json:"created_at"`
	Players   []string `json:"players"`
	// Options picked by the creator; see Info.Options.
	Options map[string]string `json:"options,omitempty"`
	// Game state, only read and written by the game through Load and Save.
	Data json.RawMessage `json:"data,omitempty"`
	// Keys of stored blobs (drawings...) to delete with the match.
	Blobs  []string `json:"blobs,omitempty"`
	Result *Result  `json:"result,omitempty"`
}

// Result is how a match ended.
type Result struct {
	// Game points per player; the ranking only looks at the order.
	Scores map[string]int `json:"scores"`
	// Ranking points given to each player for their place.
	Awarded map[string]int `json:"awarded,omitempty"`
	// Unranked results do not go to the ranking (beating only bots).
	Unranked bool `json:"unranked,omitempty"`
	// Markdown posted in the match thread.
	Summary string `json:"summary"`
	// More thread replies after the summary, for results with many files.
	Parts []Part `json:"parts,omitempty"`
}

// Part is a thread reply with the results; Mattermost takes up to 10 files per post.
type Part struct {
	Message string `json:"message"`
	Files   []File `json:"files,omitempty"`
}

// File is a blob posted as an attachment when the match ends.
type File struct {
	Name string `json:"name"`
	Blob string `json:"blob"`
}

// Refusal is a move the game does not accept; its text is shown to the player.
type Refusal string

func (r Refusal) Error() string { return string(r) }

// IsRefusal tells a refused move from a real failure.
func IsRefusal(err error) bool {
	var r Refusal
	return errors.As(err, &r)
}

// Load decodes the game state of m into a new T.
func Load[T any](m *Match) (*T, error) {
	v := new(T)
	if len(m.Data) == 0 {
		return v, nil
	}
	if err := json.Unmarshal(m.Data, v); err != nil {
		return nil, fmt.Errorf("decode %s state: %w", m.Game, err)
	}
	return v, nil
}

// Save stores the game state v in m.
func Save(m *Match, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("encode %s state: %w", m.Game, err)
	}
	m.Data = data
	return nil
}

// Decode reads a move, refusing a malformed one.
func Decode[T any](move json.RawMessage) (*T, error) {
	v := new(T)
	if err := json.Unmarshal(move, v); err != nil {
		return nil, Refusal("Jugada no válida")
	}
	return v, nil
}

// IsPlayer tells whether userID is in m.Players.
func (m *Match) IsPlayer(userID string) bool {
	for _, p := range m.Players {
		if p == userID {
			return true
		}
	}
	return false
}

// AddPlayer adds userID if missing, for games without a lobby.
func (m *Match) AddPlayer(userID string) {
	if !m.IsPlayer(userID) {
		m.Players = append(m.Players, userID)
	}
}

// Option is the value of an option picked by the creator, or its default.
func (m *Match) Option(info Info, key string) string {
	if v, ok := m.Options[key]; ok {
		return v
	}
	for _, o := range info.Options {
		if o.Key == key {
			return o.Default
		}
	}
	return ""
}

// BotPrefix starts the id of a seat taken by a bot.
const BotPrefix = "bot:"

// IsBot tells a bot seat from a user.
func IsBot(id string) bool { return strings.HasPrefix(id, BotPrefix) }

// Seats are m.Players followed by bots up to n, for games that fill empty seats.
func Seats(m *Match, n int) []string {
	seats := slices.Clone(m.Players)
	for i := len(seats); i < n; i++ {
		seats = append(seats, fmt.Sprintf("%s%d", BotPrefix, i))
	}
	return seats
}

// Over tells whether the match has ended, finished or canceled.
func (m *Match) Over() bool {
	return m.Phase == PhaseFinished || m.Phase == PhaseCanceled
}
