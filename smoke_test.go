package main

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func TestSmoke(t *testing.T) {
	for _, sz := range [][2]int{{120, 26}, {80, 20}, {46, 24}} {
		var mm tea.Model = newModel()
		mm, _ = mm.Update(tea.WindowSizeMsg{Width: sz[0], Height: sz[1]})
		key := func(s string) { mm, _ = mm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}) }
		key("l")
		for i := 0; i < 5; i++ {
			key("j")
		}
		key("h")
		key("j")
		key("j")
		v := mm.View()
		fmt.Printf("=== %dx%d ===\n%s\n", sz[0], sz[1], v)
		for i, line := range strings.Split(v, "\n") {
			if w := lipgloss.Width(line); w > sz[0] {
				t.Errorf("%dx%d line %d is %d wide", sz[0], sz[1], i, w)
			}
		}
		if n := strings.Count(v, "\n") + 1; n > sz[1] {
			t.Errorf("%dx%d rendered %d lines", sz[0], sz[1], n)
		}
	}
}
