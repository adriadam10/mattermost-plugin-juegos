package main

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/mattermost/mattermost/server/public/model"

	"github.com/adriadam10/mattermost-plugin-juegos/server/engine"
)

// A season is a calendar month in the configured time zone, "2006-01".
const seasonKey = "season"

type standing struct {
	Points int `json:"points"`
	Played int `json:"played"`
	Wins   int `json:"wins"`
}

func rankingKey(season string) string { return "ranking-" + season }

func (p *Plugin) season(t time.Time) string {
	return t.In(p.settings.Load().loc).Format("2006-01")
}

func (p *Plugin) addToRanking(m *engine.Match, game engine.Game) error {
	if len(m.Result.Awarded) == 0 {
		return nil
	}
	ranked := engine.Ranked(m.Result.Scores)
	best := m.Result.Scores[ranked[0]]
	_, _, err := updateJSON(p.API, rankingKey(p.season(time.Now())), false, func(r *map[string]*standing) error {
		if *r == nil {
			*r = map[string]*standing{}
		}
		for id, points := range m.Result.Awarded {
			st := (*r)[id]
			if st == nil {
				st = &standing{}
				(*r)[id] = st
			}
			st.Points += points
			st.Played++
			if best > 0 && m.Result.Scores[id] == best {
				st.Wins++
			}
		}
		return nil
	})
	return err
}

type rankRow struct {
	UserID   string `json:"-"`
	Username string `json:"username"`
	standing
}

func (p *Plugin) ranking(season string) ([]rankRow, error) {
	r, _, err := loadJSON[map[string]*standing](p.API, rankingKey(season))
	if err != nil {
		return nil, err
	}
	rows := make([]rankRow, 0, len(*r))
	for id, st := range *r {
		rows = append(rows, rankRow{UserID: id, Username: p.username(id), standing: *st})
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Points != rows[j].Points {
			return rows[i].Points > rows[j].Points
		}
		if rows[i].Wins != rows[j].Wins {
			return rows[i].Wins > rows[j].Wins
		}
		return rows[i].Username < rows[j].Username
	})
	return rows, nil
}

func (p *Plugin) rankingText(season string, limit int) (string, error) {
	rows, err := p.ranking(season)
	if err != nil {
		return "", err
	}
	if len(rows) == 0 {
		return fmt.Sprintf("Nadie ha puntuado aún en la temporada %s.", season), nil
	}
	var b strings.Builder
	fmt.Fprintf(&b, "🏆 **Ranking de la temporada %s**\n\n| | Jugador | Puntos | Partidas | Victorias |\n|---|---|---|---|---|\n", season)
	for i, row := range rows[:min(limit, len(rows))] {
		fmt.Fprintf(&b, "| %s | @%s | %d | %d | %d |\n", medal(i), row.Username, row.Points, row.Played, row.Wins)
	}
	return b.String(), nil
}

// checkSeason closes the last season when the month changes: podium and the
// season canvas go to the announcement channel.
func (p *Plugin) checkSeason(now time.Time) {
	current := p.season(now)
	previous, _, err := updateJSON(p.API, seasonKey, false, func(s *string) error {
		if *s == current {
			return errUnchanged
		}
		*s = current
		return nil
	})
	if errors.Is(err, errUnchanged) {
		return
	}
	if err != nil {
		p.API.LogError("Failed to check the season", "err", err.Error())
		return
	}
	if previous == "" {
		return
	}
	p.closeSeason(previous)
}

var errUnchanged = errors.New("unchanged")

func (p *Plugin) closeSeason(season string) {
	channelID, _ := p.announceChannelID.Load().(string)
	if channelID == "" {
		return
	}
	text, err := p.rankingText(season, 10)
	if err != nil {
		p.API.LogError("Failed to load the closing ranking", "season", season, "err", err.Error())
		return
	}
	post := &model.Post{UserId: p.botID, ChannelId: channelID, Message: "🎉 **Se cierra la temporada " + season + "**\n\n" + text}
	if png, err := p.canvasPNG(season); err == nil && png != nil {
		if info, appErr := p.API.UploadFile(png, channelID, "lienzo-"+season+".png"); appErr == nil {
			post.FileIds = []string{info.Id}
			post.Message += "\nY así quedó el lienzo de la temporada 🖼️"
		}
	}
	if _, appErr := p.API.CreatePost(post); appErr != nil {
		p.API.LogError("Failed to announce the end of the season", "err", appErr.Error())
	}
}

// canvasPNG renders the canvas of season, or nil if nobody painted.
func (p *Plugin) canvasPNG(season string) ([]byte, error) {
	c, raw, err := loadJSON[canvas](p.API, canvasKey(season))
	if err != nil || raw == nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := c.encodePNG(&buf, canvasExportScale); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
