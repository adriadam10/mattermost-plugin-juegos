package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"path"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/mattermost/mattermost/server/public/model"
)

// Content packs: content/<kind>*.json, each {"source": "...", "items": [...]}.
// More texts for a kind are a new file or more lines in one; see the README.
//
//go:embed content/*.json
var contentFS embed.FS

// Kinds people can add to with /juegos frase.
var contentKinds = map[string]string{
	"absurdas": "frase con un hueco ___ (o una pregunta) para Respuestas absurdas",
	"telefono": "frase de arranque para Teléfono dibujado",
	"pie":      "tema de GIF en inglés para Pie de foto (p. ej. «confused cat»)",
}

const (
	maxUserContent = 200
	// Picks avoid the texts used lately, up to half of what there is.
	recentShare = 2
)

type pack struct {
	Source string   `json:"source"`
	Items  []string `json:"items"`
}

func builtinContent() (map[string][]string, error) {
	files, err := contentFS.ReadDir("content")
	if err != nil {
		return nil, err
	}
	out := map[string][]string{}
	for _, f := range files {
		raw, err := contentFS.ReadFile(path.Join("content", f.Name()))
		if err != nil {
			return nil, err
		}
		var pk pack
		if err := json.Unmarshal(raw, &pk); err != nil {
			return nil, fmt.Errorf("content/%s: %w", f.Name(), err)
		}
		kind, _, _ := strings.Cut(strings.TrimSuffix(f.Name(), ".json"), "-")
		out[kind] = append(out[kind], pk.Items...)
	}
	return out, nil
}

// userContent is what people added for a kind.
type userContent struct {
	Text   string `json:"text"`
	Author string `json:"author"`
}

func userContentKey(kind string) string { return "content-" + kind }
func recentKey(kind string) string      { return "recent-" + kind }

func (p *Plugin) allContent(kind string) []string {
	builtin, err := builtinContent()
	if err != nil {
		p.API.LogError("Broken content pack", "err", err.Error())
	}
	items := builtin[kind]
	added, _, err := loadJSON[[]userContent](p.API, userContentKey(kind))
	if err == nil {
		for _, c := range *added {
			items = append(items, c.Text)
		}
	}
	return items
}

// pick returns up to n random texts of kind, avoiding the ones used lately so
// the same prompt does not come back every other match.
func (p *Plugin) pick(kind string, n int) []string {
	items := p.allContent(kind)
	recent, _, err := loadJSON[[]string](p.API, recentKey(kind))
	if err != nil {
		recent = &[]string{}
	}
	fresh := slices.DeleteFunc(slices.Clone(items), func(s string) bool { return slices.Contains(*recent, s) })
	if len(fresh) < n {
		fresh = slices.Clone(items)
	}
	rand.Shuffle(len(fresh), func(i, j int) { fresh[i], fresh[j] = fresh[j], fresh[i] })
	picked := fresh[:min(n, len(fresh))]
	// Best effort: a lost write only makes a repeat a bit more likely.
	updated := append(slices.Clone(*recent), picked...)
	if limit := len(items) / recentShare; len(updated) > limit {
		updated = updated[len(updated)-limit:]
	}
	if data, err := json.Marshal(updated); err == nil {
		_ = p.API.KVSet(recentKey(kind), data)
	}
	return picked
}

// addContent saves a text a user proposes for kind.
func (p *Plugin) addContent(kind, text, userID string) error {
	if _, ok := contentKinds[kind]; !ok {
		return fmt.Errorf("tipo desconocido %q", kind)
	}
	text = strings.Join(strings.Fields(text), " ")
	if utf8.RuneCountInString(text) < 3 || utf8.RuneCountInString(text) > 200 {
		return fmt.Errorf("la frase tiene que tener entre 3 y 200 caracteres")
	}
	if kind == "absurdas" && !strings.Contains(text, "___") && !strings.HasSuffix(text, "?") {
		return fmt.Errorf("las frases de absurdas llevan un hueco ___ o son una pregunta")
	}
	if slices.Contains(p.allContent(kind), text) {
		return fmt.Errorf("esa ya está")
	}
	_, _, err := updateJSON(p.API, userContentKey(kind), false, func(list *[]userContent) error {
		if len(*list) >= maxUserContent {
			return fmt.Errorf("ya hay %d frases añadidas de %s", maxUserContent, kind)
		}
		*list = append(*list, userContent{Text: text, Author: userID})
		return nil
	})
	return err
}

// removeContent deletes a text people added; only system admins may.
func (p *Plugin) removeContent(kind, text, userID string) error {
	if !p.API.HasPermissionTo(userID, model.PermissionManageSystem) {
		return fmt.Errorf("solo un administrador puede quitar frases")
	}
	text = strings.Join(strings.Fields(text), " ")
	found := false
	_, _, err := updateJSON(p.API, userContentKey(kind), false, func(list *[]userContent) error {
		before := len(*list)
		*list = slices.DeleteFunc(*list, func(c userContent) bool { return c.Text == text })
		found = len(*list) < before
		return nil
	})
	if err == nil && !found {
		return fmt.Errorf("no está entre las añadidas (las del plugin se quitan en el repo)")
	}
	return err
}
