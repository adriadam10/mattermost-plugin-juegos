package main

import (
	"errors"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
	// The distroless Mattermost image has no zoneinfo.
	_ "time/tzdata"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
)

const (
	botUsername = "juegos"
	// Live phases last about a minute: deadlines are checked this often.
	tickInterval = 5 * time.Second
)

type Plugin struct {
	plugin.MattermostPlugin
	router   http.Handler
	settings atomic.Pointer[settings]
	botID    string
	// Resolved from the AnnouncementChannel setting; empty when announcements are off.
	announceChannelID atomic.Value
	stop              chan struct{}
	ticker            sync.WaitGroup
}

func (p *Plugin) OnActivate() error {
	s := p.settings.Load()
	if s == nil {
		return errors.New("no valid configuration")
	}
	botID, err := p.API.EnsureBotUser(&model.Bot{Username: botUsername, DisplayName: "Juegos", Description: "Minijuegos para el equipo"})
	if err != nil {
		return fmt.Errorf("ensure bot: %w", err)
	}
	p.botID = botID
	if avatar, err := avatarPNG(); err != nil {
		p.API.LogWarn("Failed to draw the bot avatar", "err", err.Error())
	} else if appErr := p.API.SetProfileImage(botID, avatar); appErr != nil {
		p.API.LogWarn("Failed to set the bot avatar", "err", appErr.Error())
	}
	if err := p.applySettings(nil, s); err != nil {
		return err
	}
	p.router = p.newRouter()
	p.stop = make(chan struct{})
	p.ticker.Add(1)
	go p.tick()
	return nil
}

func (p *Plugin) OnDeactivate() error {
	close(p.stop)
	p.ticker.Wait()
	return nil
}

// applySettings registers the slash command and resolves the announcement
// channel; old is nil on activation.
func (p *Plugin) applySettings(old, s *settings) error {
	if old != nil && old.trigger != s.trigger {
		if err := p.API.UnregisterCommand("", old.trigger); err != nil {
			p.API.LogWarn("Failed to unregister the old slash command", "trigger", old.trigger, "err", err.Error())
		}
	}
	if err := p.API.RegisterCommand(command(s.trigger)); err != nil {
		return fmt.Errorf("register /%s: %w", s.trigger, err)
	}
	p.announceChannelID.Store(p.resolveAnnounceChannel(s))
	return nil
}

// resolveAnnounceChannel finds the channel and puts the bot in it. A wrong
// channel only turns announcements off: the games keep working.
func (p *Plugin) resolveAnnounceChannel(s *settings) string {
	if s.announceTeam == "" {
		return ""
	}
	team, appErr := p.API.GetTeamByName(s.announceTeam)
	if appErr != nil {
		p.API.LogError("Announcement team not found", "team", s.announceTeam, "err", appErr.Error())
		return ""
	}
	channel, appErr := p.API.GetChannelByName(team.Id, s.announceChannel, false)
	if appErr != nil {
		p.API.LogError("Announcement channel not found", "channel", s.announceChannel, "err", appErr.Error())
		return ""
	}
	if _, appErr := p.API.CreateTeamMember(team.Id, p.botID); appErr != nil {
		p.API.LogWarn("Failed to add the bot to the announcement team", "err", appErr.Error())
	}
	if _, appErr := p.API.AddChannelMember(channel.Id, p.botID); appErr != nil {
		p.API.LogWarn("Failed to add the bot to the announcement channel", "err", appErr.Error())
	}
	return channel.Id
}

func (p *Plugin) ServeHTTP(_ *plugin.Context, w http.ResponseWriter, r *http.Request) {
	p.router.ServeHTTP(w, r)
}

func (p *Plugin) tick() {
	defer p.ticker.Done()
	t := time.NewTicker(tickInterval)
	defer t.Stop()
	lastPurge := time.Time{}
	for {
		select {
		case <-p.stop:
			return
		case now := <-t.C:
			p.expireDeadlines(now)
			p.checkSeason(now)
			if now.Sub(lastPurge) > 24*time.Hour {
				p.purgeOldMatches(now)
				lastPurge = now
			}
		}
	}
}

func (p *Plugin) username(userID string) string {
	if userID == "" {
		return ""
	}
	user, appErr := p.API.GetUser(userID)
	if appErr != nil {
		return "?"
	}
	return user.Username
}

func main() {
	plugin.ClientMain(&Plugin{})
}
