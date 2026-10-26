package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"time"

	"github.com/mattermost/mattermost/server/public/model"

	"github.com/adriadam10/mattermost-plugin-juegos/server/engine"
)

const (
	pixelEvent = "pixel"
	// The season canvas posted at the end is scaled up so it is not a thumbnail.
	canvasExportScale = 8
)

// palette is PICO-8's, with white first as the background.
var palette = []string{
	"#fff1e8", "#000000", "#1d2b53", "#7e2553", "#008751", "#ab5236", "#5f574f", "#c2c3c7",
	"#ff004d", "#ffa300", "#ffec27", "#00e436", "#29adff", "#83769c", "#ff77a8", "#ffccaa",
}

// canvas is the season's shared pixel canvas.
type canvas struct {
	Size   int    `json:"size"`
	Pixels []byte `json:"pixels"`
	// Who painted each pixel: index+1 into Users, 0 for nobody.
	Owners []uint16 `json:"owners"`
	Users  []string `json:"users"`
	// When each user last painted, Unix milliseconds.
	Last map[string]int64 `json:"last"`
}

func canvasKey(season string) string { return "canvas-" + season }

func (c *canvas) init(size int) {
	if c.Size != 0 {
		return
	}
	c.Size = size
	c.Pixels = make([]byte, size*size)
	c.Owners = make([]uint16, size*size)
	c.Last = map[string]int64{}
}

func (c *canvas) userIndex(userID string) uint16 {
	for i, id := range c.Users {
		if id == userID {
			return uint16(i + 1)
		}
	}
	c.Users = append(c.Users, userID)
	return uint16(len(c.Users))
}

// nextAt is when userID may paint again.
func (c *canvas) nextAt(userID string, cooldown time.Duration) int64 {
	last, ok := c.Last[userID]
	if !ok {
		return 0
	}
	return last + cooldown.Milliseconds()
}

func (c *canvas) paint(userID string, x, y, colour int, now time.Time, cooldown time.Duration) error {
	if x < 0 || y < 0 || x >= c.Size || y >= c.Size {
		return engine.Refusal("Fuera del lienzo")
	}
	if colour < 0 || colour >= len(palette) {
		return engine.Refusal("Color desconocido")
	}
	if now.UnixMilli() < c.nextAt(userID, cooldown) {
		return engine.Refusal("Aún no puedes pintar otro píxel")
	}
	i := y*c.Size + x
	c.Pixels[i] = byte(colour)
	c.Owners[i] = c.userIndex(userID)
	c.Last[userID] = now.UnixMilli()
	return nil
}

func (c *canvas) encodePNG(w io.Writer, scale int) error {
	pal := make(color.Palette, len(palette))
	for i, hex := range palette {
		var r, g, b uint8
		_, _ = fmt.Sscanf(hex, "#%02x%02x%02x", &r, &g, &b)
		pal[i] = color.RGBA{R: r, G: g, B: b, A: 255}
	}
	img := image.NewPaletted(image.Rect(0, 0, c.Size*scale, c.Size*scale), pal)
	for y := range c.Size * scale {
		for x := range c.Size * scale {
			img.Pix[y*img.Stride+x] = c.Pixels[(y/scale)*c.Size+x/scale]
		}
	}
	return png.Encode(w, img)
}

type canvasView struct {
	Season  string   `json:"season"`
	Size    int      `json:"size"`
	Palette []string `json:"palette"`
	Pixels  []byte   `json:"pixels"`
	Owners  []uint16 `json:"owners"`
	Users   []string `json:"users"`
	NextAt  int64    `json:"next_at"`
	// Cooldown in milliseconds.
	Cooldown int64 `json:"cooldown"`
	Now      int64 `json:"now"`
}

func (p *Plugin) canvasView(c *canvas, season, userID string) canvasView {
	s := p.settings.Load()
	c.init(s.canvasSize)
	users := make([]string, len(c.Users))
	for i, id := range c.Users {
		users[i] = p.username(id)
	}
	return canvasView{
		Season: season, Size: c.Size, Palette: palette, Pixels: c.Pixels, Owners: c.Owners, Users: users,
		NextAt: c.nextAt(userID, s.canvasCooldown), Cooldown: s.canvasCooldown.Milliseconds(), Now: time.Now().UnixMilli(),
	}
}

func (p *Plugin) loadCanvas(userID string) (canvasView, error) {
	season := p.season(time.Now())
	c, _, err := loadJSON[canvas](p.API, canvasKey(season))
	if err != nil {
		return canvasView{}, err
	}
	return p.canvasView(c, season, userID), nil
}

func (p *Plugin) paint(userID string, x, y, colour int) (canvasView, error) {
	s := p.settings.Load()
	now := time.Now()
	season := p.season(now)
	_, c, err := updateJSON(p.API, canvasKey(season), false, func(c *canvas) error {
		c.init(s.canvasSize)
		return c.paint(userID, x, y, colour, now, s.canvasCooldown)
	})
	if err != nil {
		return p.canvasView(&c, season, userID), err
	}
	p.API.PublishWebSocketEvent(pixelEvent, map[string]any{
		"season": season, "x": x, "y": y, "color": colour, "username": p.username(userID),
	}, &model.WebsocketBroadcast{})
	return p.canvasView(&c, season, userID), nil
}
