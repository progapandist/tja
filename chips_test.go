package main

import (
	"strings"
	"testing"
)

func TestChipsSurviveNarrowScreens(t *testing.T) {
	for _, w := range []int{120, 80, 64, 46, 38, 30} {
		m := newModel()
		m.w, m.h = w, 30
		var keys []string
		for _, c := range m.chips() {
			keys = append(keys, c.key)
		}
		got := strings.Join(keys, " ")
		for _, want := range []string{"t", "q", "j", "k"} {
			if !strings.Contains(" "+got+" ", " "+want+" ") {
				t.Errorf("w=%d: %q is missing from %q", w, want, got)
			}
		}
		t.Logf("w=%3d: %s", w, got)
	}
}
