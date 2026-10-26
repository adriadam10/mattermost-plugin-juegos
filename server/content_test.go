package main

import (
	"strings"
	"testing"
)

func TestContentPacks(t *testing.T) {
	content, err := builtinContent()
	if err != nil {
		t.Fatal(err)
	}
	for kind := range contentKinds {
		if len(content[kind]) < 20 {
			t.Errorf("%s has only %d texts", kind, len(content[kind]))
		}
	}
	for _, text := range content["absurdas"] {
		if !strings.Contains(text, "___") && !strings.HasSuffix(text, "?") {
			t.Errorf("absurdas without a blank or a question: %q", text)
		}
	}
}

func TestCanvasCooldown(t *testing.T) {
	var c canvas
	c.init(4)
	now := mustTime(t)
	if err := c.paint("a", 1, 2, 3, now, 60e9); err != nil {
		t.Fatal(err)
	}
	if err := c.paint("a", 0, 0, 3, now, 60e9); err == nil {
		t.Fatal("painted again within the cooldown")
	}
	if c.Pixels[2*4+1] != 3 || c.Users[c.Owners[2*4+1]-1] != "a" {
		t.Fatalf("pixel not saved: %v %v", c.Pixels, c.Owners)
	}
}
