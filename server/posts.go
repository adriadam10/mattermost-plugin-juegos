package main

import (
	"fmt"
	"strings"

	"github.com/mattermost/mattermost/server/public/model"

	"github.com/adriadam10/mattermost-plugin-juegos/server/engine"
)

const (
	// Post type the webapp draws as the match board.
	matchPostType = "custom_juegos_match"
	matchEvent    = "match"
)

// author is who posts in channel: the bot where it can join, or the creator in
// direct and group messages, where bots cannot be added.
func (p *Plugin) author(channelID, fallbackUserID string) string {
	channel, appErr := p.API.GetChannel(channelID)
	if appErr != nil || channel.Type == model.ChannelTypeDirect || channel.Type == model.ChannelTypeGroup {
		return fallbackUserID
	}
	if _, appErr := p.API.GetChannelMember(channelID, p.botID); appErr != nil {
		if _, appErr := p.API.AddChannelMember(channelID, p.botID); appErr != nil {
			p.API.LogWarn("Failed to add the bot to a channel", "channel_id", channelID, "err", appErr.Error())
			return fallbackUserID
		}
	}
	return p.botID
}

func (p *Plugin) postMatch(m *engine.Match, info engine.Info) (*model.Post, error) {
	mode := "en directo"
	if m.Mode == engine.Async {
		mode = "asíncrona"
	}
	post := &model.Post{
		UserId:    p.author(m.ChannelID, m.CreatorID),
		ChannelId: m.ChannelID,
		Type:      matchPostType,
		// What the mobile apps show, since they do not draw plugin posts.
		Message: fmt.Sprintf("%s **%s**, partida %s de @%s. Para jugar, abre Mattermost en el navegador o en el escritorio.",
			info.Emoji, info.Name, mode, p.username(m.CreatorID)),
	}
	post.AddProp("match_id", m.ID)
	created, appErr := p.API.CreatePost(post)
	if appErr != nil {
		return nil, appErr
	}
	return created, nil
}

// reply posts in the match thread.
func (p *Plugin) reply(m *engine.Match, message string, fileIDs []string) {
	post := &model.Post{UserId: p.author(m.ChannelID, m.CreatorID), ChannelId: m.ChannelID, RootId: m.PostID, Message: message, FileIds: fileIDs}
	if _, appErr := p.API.CreatePost(post); appErr != nil {
		p.API.LogError("Failed to reply in a match thread", "match", m.ID, "err", appErr.Error())
	}
}

// notice tells the thread that a new phase started, only in async matches:
// live ones are watched on screen.
func (p *Plugin) notice(m *engine.Match, game engine.Game, env *engine.Env) {
	n, ok := game.(engine.Noticer)
	if !ok || m.Mode != engine.Async || m.Over() || m.Phase == engine.PhaseLobby {
		return
	}
	if text := n.Notice(m, env); text != "" {
		p.reply(m, fmt.Sprintf("%s (hasta las %s)", text, p.clock(m.Deadline)), nil)
	}
}

func (p *Plugin) clock(ms int64) string {
	return model.GetTimeForMillis(ms).In(p.settings.Load().loc).Format("15:04 del 02/01")
}

func (p *Plugin) postResults(m *engine.Match, game engine.Game) {
	info := game.Info()
	var b strings.Builder
	fmt.Fprintf(&b, "🏁 **%s**: resultados\n\n", info.Name)
	if len(m.Result.Awarded) == 0 {
		fmt.Fprintf(&b, "_Partida amistosa: hacen falta %d participantes para puntuar en el ranking._\n\n", info.MinPlayers)
	}
	b.WriteString("| | Jugador | Puntos | Ranking |\n|---|---|---|---|\n")
	for i, id := range engine.Ranked(m.Result.Scores) {
		fmt.Fprintf(&b, "| %s | @%s | %d | +%d |\n", medal(i), p.username(id), m.Result.Scores[id], m.Result.Awarded[id])
	}
	b.WriteString("\n" + m.Result.Summary)
	p.reply(m, b.String(), nil)
	for _, part := range m.Result.Parts {
		p.reply(m, part.Message, p.uploadFiles(m.ChannelID, part.Files))
	}
}

func medal(i int) string {
	if i < 3 {
		return []string{"🥇", "🥈", "🥉"}[i]
	}
	return fmt.Sprint(i + 1)
}

// uploadFiles turns blobs into channel files; a missing one is skipped.
func (p *Plugin) uploadFiles(channelID string, files []engine.File) []string {
	ids := make([]string, 0, len(files))
	for _, f := range files {
		data, appErr := p.API.KVGet(f.Blob)
		if appErr != nil || data == nil {
			continue
		}
		info, appErr := p.API.UploadFile(data, channelID, f.Name)
		if appErr != nil {
			p.API.LogWarn("Failed to upload a file", "name", f.Name, "err", appErr.Error())
			continue
		}
		ids = append(ids, info.Id)
	}
	return ids
}

// broadcastMatch tells the channel a match changed; each client asks for its own view.
func (p *Plugin) broadcastMatch(m *engine.Match) {
	p.API.PublishWebSocketEvent(matchEvent, map[string]any{"match_id": m.ID}, &model.WebsocketBroadcast{ChannelId: m.ChannelID})
}
