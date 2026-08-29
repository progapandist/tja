package main

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func at(m model, x, y int, b tea.MouseButton) model {
	mm, _ := m.click(tea.MouseMsg{X: x, Y: y, Button: b, Action: tea.MouseActionPress})
	return mm.(model)
}

// A tap has to land on the row it looks like it lands on, and the footer chips
// have to do what their key does.
func TestClick(t *testing.T) {
	m := newModel()
	m.w, m.h = 120, 30
	g := m.geometry()

	m = at(m, 2, g.reelY+3, tea.MouseButtonLeft) // fourth prefix
	if m.focus != 0 || m.pfx != m.prefixList()[3] {
		t.Errorf("prefix tap: focus %d, prefix %q", m.focus, m.pfx)
	}
	m = at(m, g.sX0+2, g.reelY+1, tea.MouseButtonLeft) // second stem
	if m.focus != 1 || m.stem != m.stemList()[1] {
		t.Errorf("stem tap: focus %d, stem %s", m.focus, m.stem.Name)
	}

	before := m.stem
	m = at(m, g.sX0+2, g.reelY, tea.MouseButtonWheelUp)
	if m.stem == before {
		t.Error("wheel over the stem reel did not scroll it")
	}

	for _, c := range m.chips() {
		if c.key == "f" {
			was := m.filtered
			m = at(m, c.x0+1, g.footerY, tea.MouseButtonLeft)
			if m.filtered == was {
				t.Error("tapping the filter chip did not toggle it")
			}
		}
	}
	if m2 := at(m, 5, g.footerY-1, tea.MouseButtonLeft); m2.testing {
		t.Error("a tap below the reels should not start the test")
	}
}
