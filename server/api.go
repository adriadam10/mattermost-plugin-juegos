package main

import (
	"embed"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"time"

	"github.com/mattermost/mattermost/server/public/model"

	"github.com/adriadam10/mattermost-plugin-juegos/server/engine"
)

var blobKeyPattern = regexp.MustCompile(`^blob-[a-z0-9]{26}$`)

// The Spanish deck of the guiñote; see the README credits.
//
//go:embed cards/*.webp
var cardsFS embed.FS

var cardFiles = cacheForever(http.FileServerFS(cardsFS))

func cacheForever(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		next.ServeHTTP(w, r)
	})
}

func (p *Plugin) newRouter() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/games", p.handleGames)
	mux.HandleFunc("POST /api/v1/matches", p.handleCreate)
	mux.HandleFunc("GET /api/v1/matches/{id}", p.handleMatch)
	mux.HandleFunc("POST /api/v1/matches/{id}/act", p.handleAct)
	mux.HandleFunc("GET /api/v1/blobs/{key}", p.handleBlob)
	mux.Handle("GET /api/v1/cards/", http.StripPrefix("/api/v1/", cardFiles))
	mux.HandleFunc("GET /api/v1/canvas", p.handleCanvas)
	mux.HandleFunc("POST /api/v1/canvas", p.handlePaint)
	mux.HandleFunc("GET /api/v1/ranking", p.handleRanking)
	return requireUser(mux)
}

func requireUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Mattermost-User-Id") == "" {
			http.Error(w, "not authorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func userOf(r *http.Request) string { return r.Header.Get("Mattermost-User-Id") }

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func decodeBody(w http.ResponseWriter, r *http.Request, v any) bool {
	// Drawings travel as base64 PNGs, hence the room.
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(v); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return false
	}
	return true
}

// fail answers a refusal with its text for the player, and anything else as a server error.
func (p *Plugin) fail(w http.ResponseWriter, what string, err error) {
	switch {
	case engine.IsRefusal(err):
		writeJSON(w, http.StatusConflict, map[string]string{"refused": err.Error()})
	case errors.Is(err, errNotFound), errors.Is(err, errUnknownGame):
		http.Error(w, err.Error(), http.StatusNotFound)
	default:
		p.API.LogError("Request failed", "what", what, "err", err.Error())
		http.Error(w, what+" failed", http.StatusInternalServerError)
	}
}

func (p *Plugin) handleGames(w http.ResponseWriter, _ *http.Request) {
	s := p.settings.Load()
	infos := make([]engine.Info, len(games))
	for i, g := range games {
		infos[i] = g.Info()
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"games":         infos,
		"live_seconds":  int(s.livePhase.Seconds()),
		"async_hours":   int(s.asyncPhase.Hours()),
		"trigger":       s.trigger,
		"giphy_enabled": p.giphyKey() != "",
		"doom_url":      s.doomURL,
	})
}

func (p *Plugin) handleCreate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Game      string            `json:"game"`
		Mode      engine.Mode       `json:"mode"`
		ChannelID string            `json:"channel_id"`
		Options   map[string]string `json:"options"`
	}
	if !decodeBody(w, r, &req) {
		return
	}
	m, err := p.createMatch(userOf(r), req.ChannelID, req.Game, req.Mode, req.Options)
	if err != nil {
		p.fail(w, "create match", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"id": m.ID, "post_id": m.PostID})
}

// canSee: matches are as private as their channel.
func (p *Plugin) canSee(userID string, m *engine.Match) bool {
	return p.API.HasPermissionToChannel(userID, m.ChannelID, model.PermissionReadChannel)
}

func (p *Plugin) handleMatch(w http.ResponseWriter, r *http.Request) {
	m, err := p.loadMatch(r.PathValue("id"))
	if err != nil {
		p.fail(w, "load match", err)
		return
	}
	if !p.canSee(userOf(r), m) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	writeJSON(w, http.StatusOK, p.matchView(m, userOf(r)))
}

func (p *Plugin) handleAct(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Action string          `json:"action"`
		Move   json.RawMessage `json:"move"`
	}
	if !decodeBody(w, r, &req) {
		return
	}
	id, userID := r.PathValue("id"), userOf(r)
	m, err := p.loadMatch(id)
	if err != nil {
		p.fail(w, "load match", err)
		return
	}
	if !p.canSee(userID, m) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	m, err = p.act(id, userID, req.Action, req.Move)
	if err != nil {
		p.fail(w, "act", err)
		return
	}
	writeJSON(w, http.StatusOK, p.matchView(m, userID))
}

// handleBlob serves a drawing. Keys are random and only views that may show
// a drawing carry its key.
func (p *Plugin) handleBlob(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	if !blobKeyPattern.MatchString(key) {
		http.NotFound(w, r)
		return
	}
	data, appErr := p.API.KVGet(key)
	if appErr != nil || data == nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "private, max-age=31536000, immutable")
	_, _ = w.Write(data)
}

func (p *Plugin) handleCanvas(w http.ResponseWriter, r *http.Request) {
	view, err := p.loadCanvas(userOf(r))
	if err != nil {
		p.fail(w, "load canvas", err)
		return
	}
	writeJSON(w, http.StatusOK, view)
}

func (p *Plugin) handlePaint(w http.ResponseWriter, r *http.Request) {
	var req struct {
		X     int `json:"x"`
		Y     int `json:"y"`
		Color int `json:"color"`
	}
	if !decodeBody(w, r, &req) {
		return
	}
	view, err := p.paint(userOf(r), req.X, req.Y, req.Color)
	if err != nil && !engine.IsRefusal(err) {
		p.fail(w, "paint", err)
		return
	}
	resp := map[string]any{"canvas": view}
	if err != nil {
		resp["refused"] = err.Error()
	}
	writeJSON(w, http.StatusOK, resp)
}

func (p *Plugin) handleRanking(w http.ResponseWriter, r *http.Request) {
	season := r.URL.Query().Get("season")
	if _, err := time.Parse("2006-01", season); err != nil {
		season = p.season(time.Now())
	}
	rows, err := p.ranking(season)
	if err != nil {
		p.fail(w, "ranking", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"season": season, "rows": rows})
}
