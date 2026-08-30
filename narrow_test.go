package main

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// A phone gives the web terminal about 40 columns, which is also the narrowest
// the two reel panes can be drawn in. The answer card has to fit inside that
// without its values being eaten: truncating a styled string by raw runes cuts
// through the escape sequences and leaves nothing but the ellipsis behind.
func TestAnswerCardFitsANarrowTerminal(t *testing.T) {
	for _, w := range []int{40, 46, 64, 80, 120} {
		var mm tea.Model = newModel()
		mm, _ = mm.Update(tea.WindowSizeMsg{Width: w, Height: 60})
		mm, _ = mm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("t")})
		mm, _ = mm.Update(tea.KeyMsg{Type: tea.KeyEnter})

		view := mm.View()
		for i, line := range strings.Split(view, "\n") {
			if got := lipgloss.Width(line); got > w {
				t.Errorf("width %d: line %d overflows at %d columns: %q", w, i, got, line)
			}
		}
		for _, field := range []string{"rektion", "nebensatz", "präteritum"} {
			for _, line := range strings.Split(view, "\n") {
				if !strings.Contains(line, field) {
					continue
				}
				rest := strings.TrimSpace(strings.SplitN(line, field, 2)[1])
				if rest == "" || rest == "…" {
					t.Errorf("width %d: %s lost its value: %q", w, field, line)
				}
			}
		}
	}
}

// The card also has to fit vertically, which is what actually broke on a
// phone: the reels plus a revealed answer came to more rows than the terminal
// had, and centring it meant the top was the part that got cut. Measured
// against every verb, since the longest card is the one that decides.
func TestAnswerCardFitsAPhoneVertically(t *testing.T) {
	for _, h := range []int{28, 32, 40, 60} {
		worst, name := 0, ""
		for _, stem := range load() {
			for i := range stem.Verbs {
				m := newModel()
				m.w, m.h = 40, h
				m.testing, m.revealed, m.card = true, true, &stem.Verbs[i]
				if n := len(strings.Split(m.View(), "\n")); n > worst {
					worst, name = n, stem.Verbs[i].Name
				}
			}
		}
		if worst > h {
			t.Errorf("40x%d: %q needs %d rows, terminal has %d", h, name, worst, h)
		}
	}
}
