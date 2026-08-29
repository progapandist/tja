package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	leftWidth  = 26
	rightWidth = 46
	minWidth   = 72
	chromeH    = 3 // header line, footer line, and the gap between them
)

var (
	accent    = lipgloss.Color("212")
	secondary = lipgloss.Color("81")
	green     = lipgloss.Color("84")
	subtle    = lipgloss.Color("241")
	muted     = lipgloss.Color("245")
	bright    = lipgloss.Color("255")
	ink       = lipgloss.Color("235")

	box = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(subtle).
		Padding(0, 1)

	brandStyle    = lipgloss.NewStyle().Foreground(bright).Bold(true)
	kickerStyle   = lipgloss.NewStyle().Foreground(secondary)
	metaStyle     = lipgloss.NewStyle().Foreground(muted)
	stemStyle     = lipgloss.NewStyle().Foreground(accent).Bold(true)
	itemStyle     = lipgloss.NewStyle().Foreground(muted)
	selectedStyle = lipgloss.NewStyle().Foreground(ink).Background(accent).Bold(true)
	headingStyle  = lipgloss.NewStyle().Foreground(secondary).Bold(true)
	bodyStyle     = lipgloss.NewStyle().Foreground(bright)
	formStyle     = lipgloss.NewStyle().Foreground(bright).Bold(true)
	sepStyle      = lipgloss.NewStyle().Foreground(green)
	insepStyle    = lipgloss.NewStyle().Foreground(secondary)
	exampleStyle  = lipgloss.NewStyle().Foreground(muted).Italic(true)
	footerStyle   = lipgloss.NewStyle().Foreground(muted)
	keyStyle      = lipgloss.NewStyle().Foreground(secondary).Bold(true)
)

// row is one line of the left pane: either a stem heading or a verb under it.
type row struct {
	stem *Stem // set when this row is a heading
	verb *Verb
}

type model struct {
	stems     []*Stem
	rows      []row
	cursor    int // index into rows; always on a verb row
	top       int // first visible row
	w, h      int
	query     string
	filtering bool
}

func (m model) Init() tea.Cmd { return nil }

func (m *model) build() {
	q := strings.ToLower(m.query)
	m.rows = nil
	for _, s := range m.stems {
		var hits []*Verb
		for i := range s.Verbs {
			v := &s.Verbs[i]
			if q == "" || strings.Contains(strings.ToLower(v.Name+" "+v.Official+" "+v.Colloquial), q) {
				hits = append(hits, v)
			}
		}
		if len(hits) == 0 {
			continue
		}
		m.rows = append(m.rows, row{stem: s})
		for _, v := range hits {
			m.rows = append(m.rows, row{verb: v})
		}
	}
	m.cursor = 0
	m.move(1)
	m.top = 0
}

// move steps the cursor by dir, skipping heading rows.
func (m *model) move(dir int) {
	for i := m.cursor; i >= 0 && i < len(m.rows); i += dir {
		if m.rows[i].verb != nil {
			m.cursor = i
			return
		}
	}
	// Nothing that way: fall back to the nearest verb in the other direction.
	for i := m.cursor; i >= 0 && i < len(m.rows); i -= dir {
		if m.rows[i].verb != nil {
			m.cursor = i
			return
		}
	}
}

func (m *model) listHeight() int { return max(1, m.h-chromeH-box.GetVerticalFrameSize()) }

func (m *model) scroll() {
	h := m.listHeight()
	if m.cursor < m.top {
		m.top = m.cursor
	}
	if m.cursor >= m.top+h {
		m.top = m.cursor - h + 1
	}
	// Keep the stem heading visible when the cursor sits right under it.
	if m.top > 0 && m.cursor == m.top && m.rows[m.top-1].stem != nil {
		m.top--
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.w, m.h = msg.Width, msg.Height
		m.scroll()
	case tea.KeyMsg:
		if m.filtering {
			switch msg.String() {
			case "enter", "esc":
				m.filtering = false
			case "backspace":
				if m.query != "" {
					m.query = m.query[:len(m.query)-1]
					m.build()
				}
			case "ctrl+c":
				return m, tea.Quit
			default:
				// KeyRunes covers umlauts too; len(string) would not.
				if msg.Type == tea.KeyRunes || msg.Type == tea.KeySpace {
					m.query += msg.String()
					m.build()
				}
			}
			m.scroll()
			return m, nil
		}
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		case "j", "down":
			m.cursor++
			m.move(1)
		case "k", "up":
			m.cursor--
			m.move(-1)
		case "g", "home":
			m.cursor = 0
			m.move(1)
		case "G", "end":
			m.cursor = len(m.rows) - 1
			m.move(-1)
		case "ctrl+d", "pgdown":
			m.cursor += m.listHeight()
			m.move(-1)
		case "ctrl+u", "pgup":
			m.cursor -= m.listHeight()
			m.move(1)
		case "n", "tab":
			m.jumpStem(1)
		case "p", "shift+tab":
			m.jumpStem(-1)
		case "/":
			m.filtering = true
		case "x":
			m.query = ""
			m.build()
		}
		m.scroll()
	}
	return m, nil
}

// jumpStem moves to the first verb of the next or previous stem group.
func (m *model) jumpStem(dir int) {
	for i := m.cursor + dir; i >= 0 && i < len(m.rows); i += dir {
		if m.rows[i].stem != nil {
			m.cursor = i
			m.move(1)
			return
		}
	}
}

func (m model) current() *Verb {
	if m.cursor < len(m.rows) {
		return m.rows[m.cursor].verb
	}
	return nil
}

func (m model) View() string {
	if m.h == 0 {
		return ""
	}
	if m.w < minWidth {
		return fmt.Sprintf("tja braucht %d Spalten (hat %d).\n", minWidth, m.w)
	}
	inner := m.listHeight()
	centerWidth := m.w - leftWidth - rightWidth - 3*box.GetHorizontalFrameSize()

	panes := lipgloss.JoinHorizontal(lipgloss.Top,
		box.Width(leftWidth).Height(inner).Render(m.list(inner)),
		box.Width(centerWidth).Height(inner).Render(m.forms(centerWidth)),
		box.Width(rightWidth).Height(inner).Render(m.meanings(rightWidth)),
	)
	return m.header() + "\n" + panes + "\n" + m.footer()
}

func (m model) header() string {
	left := brandStyle.Render("tja") + kickerStyle.Render("  ·  deutsche Verben nach Stamm")
	right := metaStyle.Render(fmt.Sprintf("%d Stämme · %d Verben", len(m.stems), countVerbs(m.stems)))
	if m.query != "" {
		right = keyStyle.Render("/"+m.query) + metaStyle.Render(fmt.Sprintf("  %d Treffer", countRows(m.rows)))
	}
	gap := max(1, m.w-lipgloss.Width(left)-lipgloss.Width(right))
	return left + strings.Repeat(" ", gap) + right
}

func (m model) footer() string {
	if m.filtering {
		return footerStyle.Render("suchen: ") + keyStyle.Render(m.query+"█") + footerStyle.Render("   enter fertig")
	}
	k := func(key, what string) string { return keyStyle.Render(key) + footerStyle.Render(" "+what) }
	return strings.Join([]string{
		k("j/k", "verb"), k("n/p", "stamm"), k("/", "suche"), k("x", "reset"), k("q", "ende"),
	}, footerStyle.Render("  ·  "))
}

func (m model) list(h int) string {
	var b strings.Builder
	for i := m.top; i < len(m.rows) && i < m.top+h; i++ {
		r := m.rows[i]
		switch {
		case r.stem != nil:
			b.WriteString(stemStyle.Render(trunc(r.stem.Name, leftWidth)))
		case i == m.cursor:
			b.WriteString(selectedStyle.Render(pad(" "+r.verb.Name, leftWidth)))
		default:
			b.WriteString(itemStyle.Render(trunc(" "+r.verb.Name, leftWidth)))
		}
		b.WriteString("\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

func (m model) forms(w int) string {
	v := m.current()
	if v == nil {
		return metaStyle.Render("nichts gefunden")
	}
	present, past, perfect := v.Forms()
	kind := insepStyle.Render("untrennbar")
	if v.Sep {
		kind = sepStyle.Render("trennbar")
	}
	if v.Prefix() == "" {
		kind = metaStyle.Render("stamm")
	}

	line := func(label, val string) string {
		return metaStyle.Render(pad(label, 12)) + formStyle.Render(val)
	}
	rows := []string{
		stemStyle.Render(v.Name),
		metaStyle.Render(v.Stem.Gloss),
		"",
		kind + metaStyle.Render("   "+dash(v.Prefix())+" + "+v.Stem.Name),
		"",
		headingStyle.Render("wo der stamm sich ändert"),
		line("er/sie/es", present),
		line("präteritum", past),
		line("perfekt", perfect),
		"",
		headingStyle.Render("stammformen"),
		metaStyle.Render(v.Stem.Name + " · " + v.Stem.Present + " · " + v.Stem.Past + " · " + v.Stem.Aux + " " + v.Stem.PartII),
	}
	return wrapAll(rows, w)
}

func (m model) meanings(w int) string {
	v := m.current()
	if v == nil {
		return ""
	}
	rows := []string{
		headingStyle.Render("offiziell"),
		bodyStyle.Render(v.Official),
		"",
		headingStyle.Render("umgangssprachlich"),
		bodyStyle.Render(v.Colloquial),
		"",
		headingStyle.Render("beispiel"),
		exampleStyle.Render(v.Example),
	}
	return wrapAll(rows, w)
}

func wrapAll(rows []string, w int) string {
	wrap := lipgloss.NewStyle().Width(w - box.GetHorizontalFrameSize())
	for i, r := range rows {
		rows[i] = wrap.Render(r)
	}
	return strings.Join(rows, "\n")
}

func dash(s string) string {
	if s == "" {
		return "—"
	}
	return s
}

func pad(s string, w int) string {
	if n := w - lipgloss.Width(s); n > 0 {
		return s + strings.Repeat(" ", n)
	}
	return trunc(s, w)
}

func trunc(s string, w int) string {
	if lipgloss.Width(s) <= w {
		return s
	}
	r := []rune(s)
	return string(r[:max(0, w-1)]) + "…"
}

func countVerbs(stems []*Stem) int {
	n := 0
	for _, s := range stems {
		n += len(s.Verbs)
	}
	return n
}

func countRows(rows []row) int {
	n := 0
	for _, r := range rows {
		if r.verb != nil {
			n++
		}
	}
	return n
}

func main() {
	m := model{stems: load()}
	m.build()
	if _, err := tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
