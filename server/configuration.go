package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// configuration mirrors settings_schema in plugin.json. The System Console can
// store a number setting as a JSON number or as a string, so those are decoded loosely.
type configuration struct {
	CommandTrigger        string
	AnnouncementChannel   string
	TimeZone              string
	LivePhaseSeconds      any
	AsyncPhaseHours       any
	GiphyAPIKey           string
	CanvasSize            any
	CanvasCooldownMinutes any
}

// settings is the validated configuration.
type settings struct {
	trigger string
	// Empty when announcements are off.
	announceTeam, announceChannel string
	loc                           *time.Location
	livePhase, asyncPhase         time.Duration
	giphyKey                      string
	canvasSize                    int
	canvasCooldown                time.Duration
}

func number(name string, value any, def, lo, hi int) (int, error) {
	var n int
	switch v := value.(type) {
	case nil:
		return def, nil
	case float64:
		n = int(v)
	case string:
		if v == "" {
			return def, nil
		}
		parsed, err := strconv.Atoi(v)
		if err != nil {
			return 0, fmt.Errorf("%s must be a whole number, got %q", name, v)
		}
		n = parsed
	default:
		return 0, fmt.Errorf("%s has an unexpected type %T", name, v)
	}
	if n < lo || n > hi {
		return 0, fmt.Errorf("%s must be between %d and %d, got %d", name, lo, hi, n)
	}
	return n, nil
}

func orDefault(value, def string) string {
	if v := strings.TrimSpace(value); v != "" {
		return v
	}
	return def
}

func (c configuration) settings() (*settings, error) {
	s := &settings{
		trigger:  strings.TrimPrefix(orDefault(c.CommandTrigger, "juegos"), "/"),
		giphyKey: strings.TrimSpace(c.GiphyAPIKey),
	}
	if strings.ContainsAny(s.trigger, " /") {
		return nil, fmt.Errorf("CommandTrigger must be a single word, got %q", s.trigger)
	}
	var err error
	if s.loc, err = time.LoadLocation(orDefault(c.TimeZone, "Europe/Madrid")); err != nil {
		return nil, fmt.Errorf("TimeZone: %w", err)
	}
	seconds, err := number("LivePhaseSeconds", c.LivePhaseSeconds, 60, 20, 600)
	if err != nil {
		return nil, err
	}
	hours, err := number("AsyncPhaseHours", c.AsyncPhaseHours, 4, 1, 72)
	if err != nil {
		return nil, err
	}
	s.livePhase, s.asyncPhase = time.Duration(seconds)*time.Second, time.Duration(hours)*time.Hour
	if s.canvasSize, err = number("CanvasSize", c.CanvasSize, 48, 16, 128); err != nil {
		return nil, err
	}
	minutes, err := number("CanvasCooldownMinutes", c.CanvasCooldownMinutes, 3, 0, 1440)
	if err != nil {
		return nil, err
	}
	s.canvasCooldown = time.Duration(minutes) * time.Minute
	if channel := strings.TrimSpace(c.AnnouncementChannel); channel != "" {
		team, name, ok := strings.Cut(channel, "/")
		if !ok || team == "" || name == "" {
			return nil, fmt.Errorf("AnnouncementChannel must look like team-name/channel-name, got %q", channel)
		}
		s.announceTeam, s.announceChannel = team, name
	}
	return s, nil
}

func (p *Plugin) OnConfigurationChange() error {
	var c configuration
	if err := p.API.LoadPluginConfiguration(&c); err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}
	s, err := c.settings()
	if err != nil {
		return err
	}
	old := p.settings.Swap(s)
	// On the first load the plugin is not active yet; OnActivate does the rest.
	if old == nil || p.botID == "" {
		return nil
	}
	return p.applySettings(old, s)
}
