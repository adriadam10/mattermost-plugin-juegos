package main

import (
	"testing"
	"time"
)

func mustTime(t *testing.T) time.Time {
	t.Helper()
	return time.Unix(1_800_000_000, 0)
}

func TestDefaults(t *testing.T) {
	s, err := configuration{}.settings()
	if err != nil {
		t.Fatal(err)
	}
	if s.trigger != "juegos" || s.livePhase != time.Minute || s.asyncPhase != 4*time.Hour || s.canvasSize != 48 {
		t.Fatalf("defaults %+v", s)
	}
}

func TestNumbersAsStrings(t *testing.T) {
	s, err := configuration{LivePhaseSeconds: "90", AnnouncementChannel: "equipo/juegos"}.settings()
	if err != nil {
		t.Fatal(err)
	}
	if s.livePhase != 90*time.Second || s.announceChannel != "juegos" {
		t.Fatalf("%+v", s)
	}
	if _, err := (configuration{AnnouncementChannel: "juegos"}).settings(); err == nil {
		t.Fatal("accepted a channel without a team")
	}
}
