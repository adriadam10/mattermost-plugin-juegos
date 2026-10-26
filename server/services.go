package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/mattermost/mattermost/server/public/model"
)

// The GIF commands plugin already has a Giphy key; it is used when ours is empty.
const gifPluginID = "com.github.moussetc.mattermost.plugin.giphy"

var httpClient = &http.Client{Timeout: 10 * time.Second}

// services is what the plugin lends the games (engine.Services).
type services struct{ p *Plugin }

func (s services) Pick(kind string, n int) []string { return s.p.pick(kind, n) }
func (s services) Username(userID string) string    { return s.p.username(userID) }

func blobKey() string { return "blob-" + model.NewId() }

func (s services) PutBlob(data []byte) (string, error) {
	key := blobKey()
	if appErr := s.p.API.KVSet(key, data); appErr != nil {
		return "", appErr
	}
	return key, nil
}

func (s services) Gif(tag string) (string, error) {
	key := s.p.giphyKey()
	if key == "" {
		return "", errors.New("no Giphy API key: set one in the plugin settings")
	}
	q := url.Values{"api_key": {key}, "tag": {tag}, "rating": {"pg-13"}}
	resp, err := httpClient.Get("https://api.giphy.com/v1/gifs/random?" + q.Encode())
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("giphy: %s", resp.Status)
	}
	var body struct {
		Data struct {
			Images map[string]struct {
				URL string `json:"url"`
			} `json:"images"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", fmt.Errorf("giphy: %w", err)
	}
	for _, size := range []string{"downsized_medium", "fixed_height", "original"} {
		if img, ok := body.Data.Images[size]; ok && img.URL != "" {
			return img.URL, nil
		}
	}
	return "", fmt.Errorf("giphy: no GIF for %q", tag)
}

func (p *Plugin) giphyKey() string {
	if key := p.settings.Load().giphyKey; key != "" {
		return key
	}
	if cfg, ok := p.API.GetUnsanitizedConfig().PluginSettings.Plugins[gifPluginID]; ok {
		if key, ok := cfg["apikey"].(string); ok {
			return key
		}
	}
	return ""
}
