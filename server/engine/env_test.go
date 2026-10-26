package engine

import (
	"reflect"
	"testing"
)

func TestAwardSharesPlacesOnTies(t *testing.T) {
	got := Award(map[string]int{"a": 300, "b": 300, "c": 100, "d": 50, "e": 0})
	want := map[string]int{"a": 5, "b": 5, "c": 2, "d": 1, "e": 1}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Award = %v, want %v", got, want)
	}
}

func TestAwardGivesNoPodiumForZero(t *testing.T) {
	got := Award(map[string]int{"a": 0, "b": 0})
	if got["a"] != ParticipationPoints || got["b"] != ParticipationPoints {
		t.Fatalf("Award = %v", got)
	}
}
