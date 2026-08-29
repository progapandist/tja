package main

import (
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	prefixWidth = 14
	stemWidth   = 16
	centerWidth = 36
	minWidth    = 96
	chromeH     = 3 // header line, footer line, and the gap between them
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
	boxFocused = box.BorderForeground(accent)

	brandStyle   = lipgloss.NewStyle().Foreground(bright).Bold(true)
	kickerStyle  = lipgloss.NewStyle().Foreground(secondary)
	metaStyle    = lipgloss.NewStyle().Foreground(muted)
	reelStyle    = lipgloss.NewStyle().Foreground(bright)
	deadStyle    = lipgloss.NewStyle().Foreground(subtle)
	pickedStyle  = lipgloss.NewStyle().Foreground(ink).Background(accent).Bold(true)
	restingStyle = lipgloss.NewStyle().Foreground(ink).Background(subtle)
	headingStyle = lipgloss.NewStyle().Foreground(secondary).Bold(true)
	bodyStyle    = lipgloss.NewStyle().Foreground(bright)
	wordStyle    = lipgloss.NewStyle().Foreground(accent).Bold(true)
	formStyle    = lipgloss.NewStyle().Foreground(bright).Bold(true)
	ghostStyle   = lipgloss.NewStyle().Foreground(subtle)
	sepStyle     = lipgloss.NewStyle().Foreground(green)
	insepStyle   = lipgloss.NewStyle().Foreground(secondary)
	exampleStyle = lipgloss.NewStyle().Foreground(muted).Italic(true)
	footerStyle  = lipgloss.NewStyle().Foreground(muted)
	keyStyle     = lipgloss.NewStyle().Foreground(secondary).Bold(true)
)

// The two reels: prefixes on the left, stems beside them. Spinning them
// independently is the point — most combinations are real words, and the ones
// that are not are worth seeing too.
type model struct {
	stems    []*Stem
	prefixes []string        // "" (the bare stem) first, then alphabetical
	sepOf    map[string]bool // separability, taken from the verbs we do have
	pi, si   int             // reel positions
	ptop, st int             // first visible row of each reel
	focus    int             // 0 = prefix reel, 1 = stem reel
	w, h     int

	testing  bool // flash-card mode: one prefix, one stem, meaning hidden
	card     *Verb
	revealed bool
}

func newModel() model {
	stems := load()
	seen := map[string]int{} // prefix -> separable votes minus inseparable
	for _, s := range stems {
		for _, v := range s.Verbs {
			p := v.Prefix()
			if p == "" {
				continue
			}
			if _, ok := seen[p]; !ok {
				seen[p] = 0
			}
			if v.Sep {
				seen[p]++
			} else {
				seen[p]--
			}
		}
	}
	m := model{stems: stems, prefixes: []string{""}, sepOf: map[string]bool{}}
	for p, votes := range seen {
		m.prefixes = append(m.prefixes, p)
		m.sepOf[p] = votes >= 0
	}
	sort.Strings(m.prefixes[1:])
	return m
}

func (m model) Init() tea.Cmd { return nil }

// matches returns the real verbs for the current combination. It is usually
// zero or one; a handful of prefixes (über-, um-) give two words that differ
// only by separability.
func (m model) matches() []*Verb {
	var out []*Verb
	s := m.stems[m.si]
	want := m.prefixes[m.pi] + s.Name
	for i := range s.Verbs {
		if s.Verbs[i].Name == want {
			out = append(out, &s.Verbs[i])
		}
	}
	return out
}

func (m model) exists(pi, si int) bool {
	want := m.prefixes[pi] + m.stems[si].Name
	for _, v := range m.stems[si].Verbs {
		if v.Name == want {
			return true
		}
	}
	return false
}

// ghost is the word the combination would make if it existed, conjugated by
// the same rules as a real one.
func (m model) ghost() Verb {
	p := m.prefixes[m.pi]
	return Verb{Name: p + m.stems[m.si].Name, Sep: m.sepOf[p], Stem: m.stems[m.si]}
}

func (m *model) spin() {
	for {
		pi, si := rand.Intn(len(m.prefixes)), rand.Intn(len(m.stems))
		if m.exists(pi, si) {
			m.pi, m.si = pi, si
			return
		}
	}
}

func (m model) reelHeight() int { return max(1, m.h-chromeH-box.GetVerticalFrameSize()) }

func scroll(cursor, top, h int) int {
	if cursor < top {
		return cursor
	}
	if cursor >= top+h {
		return cursor - h + 1
	}
	return top
}

func (m *model) rescroll() {
	h := m.reelHeight()
	m.ptop = scroll(m.pi, m.ptop, h)
	m.st = scroll(m.si, m.st, h)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.w, m.h = msg.Width, msg.Height
	case tea.KeyMsg:
		if m.testing {
			return m.updateTest(msg)
		}
		cur, n := &m.si, len(m.stems)
		if m.focus == 0 {
			cur, n = &m.pi, len(m.prefixes)
		}
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		case "j", "down":
			*cur = min(*cur+1, n-1)
		case "k", "up":
			*cur = max(*cur-1, 0)
		case "ctrl+d", "pgdown":
			*cur = min(*cur+m.reelHeight(), n-1)
		case "ctrl+u", "pgup":
			*cur = max(*cur-m.reelHeight(), 0)
		case "g", "home":
			*cur = 0
		case "G", "end":
			*cur = n - 1
		case "h", "left", "l", "right", "tab":
			m.focus = 1 - m.focus
		case " ", "s":
			m.spin()
		case "t":
			m.testing = true
			m.deal()
		case "J":
			// Step to the next prefix that actually makes a word here.
			for i := m.pi + 1; i < len(m.prefixes); i++ {
				if m.exists(i, m.si) {
					m.pi = i
					break
				}
			}
		case "K":
			for i := m.pi - 1; i >= 0; i-- {
				if m.exists(i, m.si) {
					m.pi = i
					break
				}
			}
		}
		m.rescroll()
	}
	return m, nil
}

// deal picks the next flash card and moves the reels onto it, so leaving the
// test drops you where the card was.
func (m *model) deal() {
	m.spin()
	m.card = m.matches()[0]
	m.revealed = false
}

func (m model) updateTest(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "esc", "t":
		m.testing = false
	case " ", "enter":
		if m.revealed {
			m.deal()
		} else {
			m.revealed = true
		}
	case "n", "j", "right", "l":
		m.deal()
	case "r":
		m.revealed = true
	}
	return m, nil
}

func (m model) testView() string {
	v := m.card
	reel := func(s string) string {
		return box.Width(18).Align(lipgloss.Center).Render("\n" + wordStyle.Render(s) + "\n")
	}
	prefix := m.prefixLabel(v.Prefix())
	card := []string{
		lipgloss.JoinHorizontal(lipgloss.Top, reel(prefix), reel(v.Stem.Name)),
		"",
	}
	if !m.revealed {
		card = append(card,
			metaStyle.Render("Was bedeutet es? Wie lauten die Formen?"),
			"",
			footerStyle.Render("␣ aufdecken  ·  n nächste  ·  esc zurück"))
	} else {
		present, past, perfect := v.Forms()
		kind := insepStyle.Render("untrennbar")
		if v.Sep {
			kind = sepStyle.Render("trennbar")
		}
		if v.Prefix() == "" {
			kind = metaStyle.Render("stamm")
		}
		w := min(64, m.w-4)
		wrap := lipgloss.NewStyle().Width(w)
		card = append(card,
			wordStyle.Render(v.Name)+"   "+kind,
			formStyle.Render(present+"  ·  "+past+"  ·  "+perfect),
			"",
			wrap.Render(headingStyle.Render("offiziell   ")+bodyStyle.Render(v.Official)),
			wrap.Render(headingStyle.Render("umgangssprachlich   ")+bodyStyle.Render(v.Colloquial)),
			"",
			wrap.Render(exampleStyle.Render(v.Example)),
			"",
			footerStyle.Render("␣/n nächste  ·  esc zurück"))
	}
	body := lipgloss.JoinVertical(lipgloss.Center, card...)
	return lipgloss.Place(m.w, m.h, lipgloss.Center, lipgloss.Center, body)
}

func (m model) View() string {
	if m.h == 0 {
		return ""
	}
	if m.testing {
		return m.testView()
	}
	if m.w < minWidth {
		return fmt.Sprintf("tja braucht %d Spalten (hat %d).\n", minWidth, m.w)
	}
	inner := m.reelHeight()
	rightWidth := m.w - prefixWidth - stemWidth - centerWidth - 4*box.GetHorizontalFrameSize()

	pane := func(focused bool, w int, s string) string {
		b := box
		if focused {
			b = boxFocused
		}
		return b.Width(w).Height(inner).Render(s)
	}
	panes := lipgloss.JoinHorizontal(lipgloss.Top,
		pane(m.focus == 0, prefixWidth, m.prefixReel(inner)),
		pane(m.focus == 1, stemWidth, m.stemReel(inner)),
		pane(false, centerWidth, m.forms(centerWidth)),
		pane(false, rightWidth, m.meanings(rightWidth)),
	)
	return m.header() + "\n" + panes + "\n" + m.footer()
}

func (m model) header() string {
	left := brandStyle.Render("tja") + kickerStyle.Render("  ·  Präfix + Stamm = ?")
	word := m.prefixes[m.pi] + m.stems[m.si].Name
	right := wordStyle.Render(word)
	if !m.exists(m.pi, m.si) {
		right = ghostStyle.Render(word + "  (kein Wort)")
	}
	gap := max(1, m.w-lipgloss.Width(left)-lipgloss.Width(right))
	return left + strings.Repeat(" ", gap) + right
}

func (m model) footer() string {
	k := func(key, what string) string { return keyStyle.Render(key) + footerStyle.Render(" "+what) }
	return strings.Join([]string{
		k("j/k", "drehen"), k("h/l", "walze"), k("J/K", "nur echte"), k("␣", "zufall"),
		k("t", "test"), k("q", "ende"),
	}, footerStyle.Render("  ·  "))
}

// prefixLabel marks separable prefixes with the hyphen dictionaries use.
func (m model) prefixLabel(p string) string {
	switch {
	case p == "":
		return "—"
	case m.sepOf[p]:
		return p + "-"
	}
	return p
}

func (m model) prefixReel(h int) string {
	var b strings.Builder
	for i := m.ptop; i < len(m.prefixes) && i < m.ptop+h; i++ {
		label := m.prefixLabel(m.prefixes[i])
		style := reelStyle
		if !m.exists(i, m.si) {
			style = deadStyle // no such word with the stem currently showing
		}
		if i == m.pi {
			style = restingStyle
			if m.focus == 0 {
				style = pickedStyle
			}
		}
		b.WriteString(style.Render(pad(" "+label, prefixWidth)) + "\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

func (m model) stemReel(h int) string {
	var b strings.Builder
	for i := m.st; i < len(m.stems) && i < m.st+h; i++ {
		style := reelStyle
		if !m.exists(m.pi, i) {
			style = deadStyle
		}
		if i == m.si {
			style = restingStyle
			if m.focus == 1 {
				style = pickedStyle
			}
		}
		b.WriteString(style.Render(pad(" "+m.stems[i].Name, stemWidth)) + "\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

func (m model) forms(w int) string {
	s := m.stems[m.si]
	hits := m.matches()
	var rows []string

	if len(hits) == 0 {
		g := m.ghost()
		present, past, perfect := g.Forms()
		rows = []string{
			ghostStyle.Render(g.Name),
			metaStyle.Render("kein belegtes Wort — so ginge es aber"),
			"",
			ghostStyle.Render("er/sie/es   " + present),
			ghostStyle.Render("präteritum  " + past),
			ghostStyle.Render("perfekt     " + perfect),
		}
	}
	for _, v := range hits {
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
		rows = append(rows,
			wordStyle.Render(v.Name),
			kind+metaStyle.Render("   "+dash(v.Prefix())+" + "+s.Name),
			"",
			headingStyle.Render("wo der stamm sich ändert"),
			line("er/sie/es", present),
			line("präteritum", past),
			line("perfekt", perfect),
			"",
		)
	}
	rows = append(rows,
		"",
		headingStyle.Render("stammformen"),
		metaStyle.Render(s.Name+" · "+s.Present+" · "+s.Past+" · "+s.Aux+" "+s.PartII),
		metaStyle.Render(s.Gloss),
	)
	return wrapAll(rows, w)
}

func (m model) meanings(w int) string {
	hits := m.matches()
	if len(hits) == 0 {
		var real []string
		for i, p := range m.prefixes {
			if m.exists(i, m.si) {
				real = append(real, dash(p))
			}
		}
		return wrapAll([]string{
			headingStyle.Render("es gibt stattdessen"),
			bodyStyle.Render(strings.Join(real, ", ") + " + " + m.stems[m.si].Name),
			"",
			metaStyle.Render("J/K springt zur nächsten echten Vorsilbe."),
		}, w)
	}
	var rows []string
	for _, v := range hits {
		if len(hits) > 1 {
			label := "untrennbar"
			if v.Sep {
				label = "trennbar"
			}
			rows = append(rows, wordStyle.Render(v.Name+" ("+label+")"))
		}
		rows = append(rows,
			headingStyle.Render("offiziell"),
			bodyStyle.Render(v.Official),
			"",
			headingStyle.Render("umgangssprachlich"),
			bodyStyle.Render(v.Colloquial),
			"",
			headingStyle.Render("beispiel"),
			exampleStyle.Render(v.Example),
			"",
		)
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

func main() {
	m := newModel()
	if _, err := tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
