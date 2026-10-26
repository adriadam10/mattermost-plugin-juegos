package main

import (
	"time"

	"github.com/adriadam10/mattermost-plugin-juegos/server/engine"
)

type playerView struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

type scoreView struct {
	Username string `json:"username"`
	Score    int    `json:"score"`
	Awarded  int    `json:"awarded"`
}

// matchView is a match as the webapp draws it for one user.
type matchView struct {
	ID        string       `json:"id"`
	Info      engine.Info  `json:"info"`
	Mode      engine.Mode  `json:"mode"`
	Phase     string       `json:"phase"`
	Round     int          `json:"round"`
	Deadline  int64        `json:"deadline"`
	Now       int64        `json:"now"`
	Creator   string       `json:"creator"`
	IsCreator bool         `json:"is_creator"`
	Joined    bool         `json:"joined"`
	Players   []playerView `json:"players"`
	Game      any          `json:"game"`
	Scores    []scoreView  `json:"scores,omitempty"`
}

func (p *Plugin) matchView(m *engine.Match, userID string) matchView {
	game, _ := gameByID(m.Game)
	v := matchView{
		ID: m.ID, Mode: m.Mode, Phase: m.Phase, Round: m.Round, Deadline: m.Deadline,
		Now: time.Now().UnixMilli(), Creator: p.username(m.CreatorID), IsCreator: m.CreatorID == userID,
		Joined: m.IsPlayer(userID), Players: make([]playerView, len(m.Players)),
	}
	for i, id := range m.Players {
		v.Players[i] = playerView{ID: id, Username: p.username(id)}
	}
	if game == nil {
		return v
	}
	v.Info = game.Info()
	if m.Phase != engine.PhaseLobby && m.Phase != engine.PhaseCanceled {
		v.Game = game.View(m, userID, p.env(m, time.Now()))
	}
	if m.Result != nil {
		for _, id := range engine.Ranked(m.Result.Scores) {
			v.Scores = append(v.Scores, scoreView{Username: p.username(id), Score: m.Result.Scores[id], Awarded: m.Result.Awarded[id]})
		}
	}
	return v
}
