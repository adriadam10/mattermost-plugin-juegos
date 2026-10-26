package main

import (
	"encoding/json"
	"errors"
	"math/rand/v2"
	"slices"
	"time"

	"github.com/mattermost/mattermost/server/public/model"

	"github.com/adriadam10/mattermost-plugin-juegos/server/engine"
)

const (
	activeKey  = "active"
	archiveKey = "archive"
	// Finished matches (and their drawings) are deleted after this long; the
	// results stay in the channel thread.
	keepMatches = 30 * 24 * time.Hour
	// The lobby lasts this many phases of the mode.
	lobbyWeight = 2
)

var errUnknownGame = errors.New("unknown game")

func matchKey(id string) string { return "match-" + id }

type archived struct {
	ID    string `json:"id"`
	Ended int64  `json:"ended"`
}

// env is what a game call gets for match m.
func (p *Plugin) env(m *engine.Match, now time.Time) *engine.Env {
	s := p.settings.Load()
	base := s.asyncPhase
	if m.Mode == engine.Live {
		base = s.livePhase
	}
	return &engine.Env{Services: services{p}, Now: now, Rand: rand.New(rand.NewPCG(uint64(now.UnixNano()), rand.Uint64())), Base: base}
}

// createMatch starts a match of gameID in channelID and posts it there.
func (p *Plugin) createMatch(userID, channelID, gameID string, mode engine.Mode) (*engine.Match, error) {
	game, ok := gameByID(gameID)
	if !ok {
		return nil, errUnknownGame
	}
	if mode != engine.Live && mode != engine.Async {
		return nil, engine.Refusal("Modo desconocido")
	}
	if !p.API.HasPermissionToChannel(userID, channelID, model.PermissionCreatePost) {
		return nil, engine.Refusal("No puedes escribir en este canal")
	}
	now := time.Now()
	m := &engine.Match{ID: model.NewId(), Game: gameID, ChannelID: channelID, CreatorID: userID, Mode: mode, CreatedAt: now.UnixMilli()}
	env := p.env(m, now)
	if game.Info().Lobby {
		m.Players = []string{userID}
		env.SetPhase(m, engine.PhaseLobby, lobbyWeight)
	} else if err := game.Start(m, env); err != nil {
		return nil, err
	}
	post, err := p.postMatch(m, game.Info())
	if err != nil {
		return nil, err
	}
	m.PostID = post.Id
	data, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	if appErr := p.API.KVSet(matchKey(m.ID), data); appErr != nil {
		return nil, appErr
	}
	if _, _, err := updateJSON(p.API, activeKey, false, func(ids *[]string) error {
		*ids = append(*ids, m.ID)
		return nil
	}); err != nil {
		return nil, err
	}
	p.notice(m, game, env)
	return m, nil
}

func (p *Plugin) loadMatch(id string) (*engine.Match, error) {
	m, raw, err := loadJSON[engine.Match](p.API, matchKey(id))
	if err == nil && raw == nil {
		return nil, errNotFound
	}
	return m, err
}

// updateMatch applies change and then does what follows from the new phase:
// notices, results and ranking, and telling the channel.
func (p *Plugin) updateMatch(id string, change func(*engine.Match, engine.Game, *engine.Env) error) (*engine.Match, error) {
	var game engine.Game
	var env *engine.Env
	before, after, err := updateJSON(p.API, matchKey(id), true, func(m *engine.Match) error {
		g, ok := gameByID(m.Game)
		if !ok {
			return errUnknownGame
		}
		game, env = g, p.env(m, time.Now())
		if err := change(m, g, env); err != nil {
			return err
		}
		if m.Phase == engine.PhaseFinished && m.Result != nil && m.Result.Awarded == nil {
			m.Result.Awarded = map[string]int{}
			if len(m.Result.Scores) >= g.Info().MinPlayers {
				m.Result.Awarded = engine.Award(m.Result.Scores)
			}
		}
		return nil
	})
	if err != nil {
		return &after, err
	}
	if before.Phase != after.Phase || before.Round != after.Round {
		p.notice(&after, game, env)
	}
	if !before.Over() && after.Over() {
		p.ended(&after, game)
	}
	p.broadcastMatch(&after)
	return &after, nil
}

// act is a move by userID: the engine's own (join, start...) or the game's.
func (p *Plugin) act(id, userID, action string, move json.RawMessage) (*engine.Match, error) {
	return p.updateMatch(id, func(m *engine.Match, g engine.Game, env *engine.Env) error {
		if m.Over() {
			return engine.Refusal("La partida ya ha terminado")
		}
		info := g.Info()
		isCreator := userID == m.CreatorID
		switch action {
		case "join":
			if m.Phase != engine.PhaseLobby {
				return engine.Refusal("La partida ya ha empezado")
			}
			if info.MaxPlayers > 0 && len(m.Players) >= info.MaxPlayers && !m.IsPlayer(userID) {
				return engine.Refusal("La partida está llena")
			}
			m.AddPlayer(userID)
		case "leave":
			if m.Phase != engine.PhaseLobby || isCreator {
				return engine.Refusal("Ahora no puedes salir")
			}
			m.Players = slices.DeleteFunc(m.Players, func(id string) bool { return id == userID })
		case "start":
			if m.Phase != engine.PhaseLobby || !isCreator {
				return engine.Refusal("Solo quien la ha creado puede empezarla")
			}
			if len(m.Players) < info.MinPlayers {
				return engine.Refusal("Faltan jugadores")
			}
			return g.Start(m, env)
		case "skip":
			if m.Phase == engine.PhaseLobby || !isCreator {
				return engine.Refusal("Solo quien la ha creado puede pasar de fase")
			}
			return g.Timeout(m, env)
		case "cancel":
			if !isCreator && !p.API.HasPermissionTo(userID, model.PermissionManageSystem) {
				return engine.Refusal("Solo quien la ha creado puede cancelarla")
			}
			m.Phase, m.Deadline = engine.PhaseCanceled, 0
		case "move":
			if m.Phase == engine.PhaseLobby {
				return engine.Refusal("La partida aún no ha empezado")
			}
			// Each game decides who may do what: votes are often open to the whole channel.
			return g.Act(m, userID, move, env)
		default:
			return engine.Refusal("Acción desconocida")
		}
		return nil
	})
}

// expireDeadlines moves on every active match whose phase is over.
func (p *Plugin) expireDeadlines(now time.Time) {
	ids, _, err := loadJSON[[]string](p.API, activeKey)
	if err != nil {
		p.API.LogError("Failed to load the active matches", "err", err.Error())
		return
	}
	for _, id := range *ids {
		m, err := p.loadMatch(id)
		if errors.Is(err, errNotFound) || err == nil && m.Over() {
			p.unlistActive(id)
			continue
		}
		if err != nil || m.Deadline == 0 || m.Deadline > now.UnixMilli() {
			continue
		}
		if _, err := p.updateMatch(id, expire); err != nil && !engine.IsRefusal(err) {
			p.API.LogError("Failed to move a match on", "match", id, "err", err.Error())
		}
	}
}

// expire ends a phase whose time is up. A lobby with enough players starts;
// otherwise the match is canceled.
func expire(m *engine.Match, g engine.Game, env *engine.Env) error {
	if m.Over() || m.Deadline == 0 || m.Deadline > env.Now.UnixMilli() {
		return engine.Refusal("not due")
	}
	if m.Phase != engine.PhaseLobby {
		return g.Timeout(m, env)
	}
	if len(m.Players) >= g.Info().MinPlayers {
		return g.Start(m, env)
	}
	m.Phase, m.Deadline = engine.PhaseCanceled, 0
	return nil
}

func (p *Plugin) unlistActive(id string) {
	if _, _, err := updateJSON(p.API, activeKey, false, func(ids *[]string) error {
		*ids = slices.DeleteFunc(*ids, func(x string) bool { return x == id })
		return nil
	}); err != nil {
		p.API.LogError("Failed to unlist a match", "match", id, "err", err.Error())
	}
}

// ended posts the results, adds them to the ranking and moves the match to the archive.
func (p *Plugin) ended(m *engine.Match, game engine.Game) {
	p.unlistActive(m.ID)
	if _, _, err := updateJSON(p.API, archiveKey, false, func(list *[]archived) error {
		*list = append(*list, archived{ID: m.ID, Ended: time.Now().UnixMilli()})
		return nil
	}); err != nil {
		p.API.LogError("Failed to archive a match", "match", m.ID, "err", err.Error())
	}
	if m.Phase == engine.PhaseCanceled {
		p.reply(m, "❌ Partida cancelada.", nil)
		return
	}
	if err := p.addToRanking(m, game); err != nil {
		p.API.LogError("Failed to update the ranking", "match", m.ID, "err", err.Error())
	}
	p.postResults(m, game)
}

// purgeOldMatches deletes matches that ended long ago, with their drawings.
func (p *Plugin) purgeOldMatches(now time.Time) {
	cutoff := now.Add(-keepMatches).UnixMilli()
	var old []archived
	if _, _, err := updateJSON(p.API, archiveKey, false, func(list *[]archived) error {
		old = nil
		*list = slices.DeleteFunc(*list, func(a archived) bool {
			if a.Ended < cutoff {
				old = append(old, a)
				return true
			}
			return false
		})
		return nil
	}); err != nil {
		p.API.LogError("Failed to purge old matches", "err", err.Error())
		return
	}
	for _, a := range old {
		if m, err := p.loadMatch(a.ID); err == nil {
			for _, blob := range m.Blobs {
				_ = p.API.KVDelete(blob)
			}
		}
		_ = p.API.KVDelete(matchKey(a.ID))
	}
}
