package main

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// The button row is the only way through the app on a touchscreen, so a tap
// has to land on the row it is actually drawn on. It used to be drawn one row
// above where the hit test looked, and every tap fell through to "flip the
// card" instead — which made "esc back" behave like "space".
func TestTappingTheGutterInTestMode(t *testing.T) {
	for _, h := range []int{28, 32, 40, 60} {
		m := newModel()
		m.w, m.h = 40, h
		m.deal()
		m.testing, m.revealed = true, true

		lines := strings.Split(m.View(), "\n")
		drawn := -1
		for i, l := range lines {
			if strings.Contains(l, "back") {
				drawn = i
			}
		}
		if drawn != m.gutterY() {
			t.Fatalf("h=%d: gutter drawn on row %d, hit test reads row %d", h, drawn, m.gutterY())
		}

		var back *chip
		for i, c := range m.chips() {
			if c.key == "esc" {
				back = &m.chips()[i]
			}
		}
		if back == nil {
			t.Fatalf("h=%d: no way back", h)
		}
		next, _ := m.click(tea.MouseMsg{
			X: back.x0, Y: m.gutterY(),
			Button: tea.MouseButtonLeft, Action: tea.MouseActionPress,
		})
		if next.(model).testing {
			t.Errorf("h=%d: tapping %q did not leave test mode", h, strings.TrimSpace(back.label))
		}
	}
}

// Whatever survives the squeeze has to fit whole: a chip clipped by the right
// edge is still a tap target, on a label you can only half read.
func TestGutterFitsEveryWidth(t *testing.T) {
	for _, w := range []int{30, 36, 40, 46, 64, 120} {
		for _, testing_ := range []bool{false, true} {
			m := newModel()
			m.w, m.h = w, 30
			m.deal()
			m.testing, m.revealed = testing_, true
			for _, c := range m.chips() {
				if c.x1 >= w {
					t.Errorf("w=%d testing=%v: chip %q runs to column %d", w, testing_, strings.TrimSpace(c.label), c.x1)
				}
			}
			if got := lipgloss.Width(m.buttonRow(m.chips())); got > w {
				t.Errorf("w=%d testing=%v: button row is %d columns", w, testing_, got)
			}
		}
	}
}
