package main

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
)

// avatarPNG draws the bot's picture: a die showing five, so no image file is needed.
func avatarPNG() ([]byte, error) {
	const size = 128
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	face := color.RGBA{R: 0x7e, G: 0x25, B: 0x53, A: 255}
	pip := color.RGBA{R: 0xff, G: 0xf1, B: 0xe8, A: 255}
	pips := [][2]int{{36, 36}, {92, 36}, {64, 64}, {36, 92}, {92, 92}}
	for y := range size {
		for x := range size {
			c := face
			for _, p := range pips {
				if dx, dy := x-p[0], y-p[1]; dx*dx+dy*dy <= 12*12 {
					c = pip
				}
			}
			img.Set(x, y, c)
		}
	}
	var buf bytes.Buffer
	err := png.Encode(&buf, img)
	return buf.Bytes(), err
}
