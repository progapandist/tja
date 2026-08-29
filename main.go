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
	prefixWidth = 16
	stemWidth   = 16
	centerWidth = 38
	wideWidth   = 100 // four panes fit
	mediumWidth = 64  // reels plus one combined detail pane
	chromeH     = 3   // the header line, the key hints, and a spare row
	repoURL     = "https://github.com/progapandist/tja"
	// lipgloss Width() counts padding but not the border, so a pane of width w
	// occupies w+border columns and has w-padding columns of content.
	border  = 2
	padding = 2
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

	kickerStyle  = lipgloss.NewStyle().Foreground(secondary)
	metaStyle    = lipgloss.NewStyle().Foreground(muted)
	reelStyle    = lipgloss.NewStyle().Foreground(bright)
	deadStyle    = lipgloss.NewStyle().Foreground(subtle)
	ghostStyle   = lipgloss.NewStyle().Foreground(subtle)
	pickedStyle  = lipgloss.NewStyle().Foreground(ink).Background(accent).Bold(true)
	restingStyle = lipgloss.NewStyle().Foreground(ink).Background(subtle)
	headingStyle = lipgloss.NewStyle().Foreground(secondary).Bold(true)
	bodyStyle    = lipgloss.NewStyle().Foreground(bright)
	wordStyle    = lipgloss.NewStyle().Foreground(accent).Bold(true)
	formStyle    = lipgloss.NewStyle().Foreground(bright).Bold(true)
	useStyle     = lipgloss.NewStyle().Foreground(green)
	sepStyle     = lipgloss.NewStyle().Foreground(green)
	insepStyle   = lipgloss.NewStyle().Foreground(secondary)
	exampleStyle = lipgloss.NewStyle().Foreground(muted).Italic(true)
	footerStyle  = lipgloss.NewStyle().Foreground(muted)
	keyStyle     = lipgloss.NewStyle().Foreground(secondary).Bold(true)
	chipStyle    = lipgloss.NewStyle().Foreground(bright).Background(lipgloss.Color("236")).Padding(0, 1)
	badgeStyle   = lipgloss.NewStyle().Foreground(ink).Background(green).Bold(true)
	linkStyle    = lipgloss.NewStyle().Foreground(green).Underline(true)
	thumbStyle   = lipgloss.NewStyle().Foreground(accent)
	trackStyle   = lipgloss.NewStyle().Foreground(subtle)
)

// Two reels that filter each other: the prefix reel shows only prefixes that
// make a word with the stem showing, and the stem reel only stems that take
// the prefix showing. Every combination you can land on is a real verb, so the
// reels are the vocabulary list.
type model struct {
	stems    []*Stem
	all      []string        // every prefix in the file, for the unfiltered reel
	filtered bool            // reels show only combinations that are real words
	count    int             // verbs in the whole file, for the header
	sepOf    map[string]bool // separability, inferred from the verbs we have
	pfx      string          // where the prefix reel is resting ("" = bare stem)
	stem     *Stem           // where the stem reel is resting
	ptop, st int             // first visible row of each reel
	focus    int             // 0 = prefix reel, 1 = stem reel
	w, h     int

	searching bool // "/" search over prefixes, stems and meanings at once
	query     string
	hits      []*Verb
	hit       int
	saved     savedPos // where the reels were, to restore on a cancelled search

	testing  bool // flash-card mode: one prefix, one stem, meaning hidden
	card     *Verb
	revealed bool
}

func newModel() model {
	stems := load()
	votes := map[string]int{}
	count := 0
	for _, s := range stems {
		count += len(s.Verbs)
		for _, v := range s.Verbs {
			if p := v.Prefix(); p != "" {
				if v.Sep {
					votes[p]++
				} else {
					votes[p]--
				}
			}
		}
	}
	// Alphabetical by the umlaut-folded name, the order a German dictionary uses.
	sort.SliceStable(stems, func(i, j int) bool {
		return folds.Replace(stems[i].Name) < folds.Replace(stems[j].Name)
	})
	m := model{stems: stems, count: count, sepOf: map[string]bool{}, stem: stems[0], filtered: true}
	for p, n := range votes {
		m.sepOf[p] = n >= 0
		m.all = append(m.all, p)
	}
	m.all = append(m.all, "")
	sortFolded(m.all)
	return m
}

type savedPos struct {
	pfx  string
	stem *Stem
}

func (m model) Init() tea.Cmd { return nil }

// prefixList is the prefix reel: the prefixes that make a word with the stem
// currently showing. stemList is the mirror image.
func (m model) prefixList() []string {
	if !m.filtered {
		return m.all
	}
	seen := map[string]bool{}
	var out []string
	for _, v := range m.stem.Verbs {
		if p := v.Prefix(); !seen[p] {
			seen[p] = true
			out = append(out, p)
		}
	}
	sortFolded(out)
	return out
}

func sortFolded(xs []string) {
	sort.SliceStable(xs, func(i, j int) bool { return folds.Replace(xs[i]) < folds.Replace(xs[j]) })
}

func (m model) stemList() []*Stem {
	if !m.filtered {
		return m.stems
	}
	var out []*Stem
	for _, s := range m.stems {
		for _, v := range s.Verbs {
			if v.Prefix() == m.pfx {
				out = append(out, s)
				break
			}
		}
	}
	return out
}

func indexOf[T comparable](xs []T, x T) int {
	for i, y := range xs {
		if y == x {
			return i
		}
	}
	return 0
}

func (m model) pi() int { return indexOf(m.prefixList(), m.pfx) }
func (m model) si() int { return indexOf(m.stemList(), m.stem) }

// Moving one reel can only land on a combination the other reel already
// allows, so neither selection ever has to be repaired.
func (m *model) setPI(i int) {
	l := m.prefixList()
	m.pfx = l[min(max(i, 0), len(l)-1)]
}

func (m *model) setSI(i int) {
	l := m.stemList()
	m.stem = l[min(max(i, 0), len(l)-1)]
}

// exists reports whether the prefix and stem make a word.
func (m model) exists(p string, s *Stem) bool {
	for _, v := range s.Verbs {
		if v.Name == p+s.Name {
			return true
		}
	}
	return false
}

// current is the verb the reels are resting on. In unfiltered mode the reels
// can land between words, and then it is a ghost: the word the rules would
// build, marked as not attested.
func (m model) current() (Verb, bool) {
	if hits := m.matches(); len(hits) > 0 {
		return *hits[0], true
	}
	return Verb{Name: m.pfx + m.stem.Name, Sep: m.sepOf[m.pfx], Stem: m.stem}, false
}

// snap pulls the prefix reel back onto a real word, for when filtering is
// switched on while the reels sit on a ghost.
func (m *model) snap() {
	if m.exists(m.pfx, m.stem) {
		return
	}
	for _, v := range m.stem.Verbs {
		m.pfx = v.Prefix()
		return
	}
}

// matches returns the verbs for the current combination — usually one, but a
// few prefixes (über-, um-) give two words that differ only by separability.
func (m model) matches() []*Verb {
	var out []*Verb
	want := m.pfx + m.stem.Name
	for i := range m.stem.Verbs {
		if m.stem.Verbs[i].Name == want {
			out = append(out, &m.stem.Verbs[i])
		}
	}
	return out
}

// fold normalises umlauts so a query typed without them still matches. Trying
// both forms lets "uber" and "ueber" find übernehmen.
var folds = strings.NewReplacer("ä", "a", "ö", "o", "ü", "u", "ß", "s")
var expands = strings.NewReplacer("ä", "ae", "ö", "oe", "ü", "ue", "ß", "ss")

// fuzzy reports whether every rune of q appears in s in order, and how tightly
// they sit together — "anneh" and "annh" both find annehmen.
func fuzzy(s, q string) (int, bool) {
	i, first, last := 0, -1, 0
	rs := []rune(s)
	for _, r := range q {
		for i < len(rs) && rs[i] != r {
			i++
		}
		if i == len(rs) {
			return 0, false
		}
		if first < 0 {
			first = i
		}
		last, i = i, i+1
	}
	return (last - first) + first, true // tighter and earlier scores lower
}

// search ranks every verb against the query. The whole word is matched, so a
// query spanning prefix and stem ("mitneh") works, and meanings count too.
func (m *model) search() {
	q := strings.ToLower(strings.TrimSpace(m.query))
	m.hits, m.hit = nil, 0
	if q == "" {
		return
	}
	type scored struct {
		v *Verb
		n int
	}
	var all []scored
	for _, s := range m.stems {
		for i := range s.Verbs {
			v := &s.Verbs[i]
			name := strings.ToLower(v.Name)
			switch {
			case strings.HasPrefix(name, q) || strings.HasPrefix(folds.Replace(name), folds.Replace(q)):
				all = append(all, scored{v, -1000 + len(name)})
			default:
				n, ok := fuzzy(folds.Replace(name), folds.Replace(q))
				if !ok {
					n, ok = fuzzy(expands.Replace(name), expands.Replace(q))
				}
				if ok {
					all = append(all, scored{v, n})
				} else if text := strings.ToLower(v.Official + " " + v.Colloquial + " " + v.Use); strings.Contains(folds.Replace(text), folds.Replace(q)) {
					all = append(all, scored{v, 1000})
				}
			}
		}
	}
	sort.SliceStable(all, func(i, j int) bool { return all[i].n < all[j].n })
	for _, s := range all {
		m.hits = append(m.hits, s.v)
	}
	m.show()
}

// show moves both reels onto the currently selected search hit.
func (m *model) show() {
	if len(m.hits) == 0 {
		return
	}
	v := m.hits[m.hit]
	m.stem, m.pfx = v.Stem, v.Prefix()
	m.rescroll()
}

func (m *model) spin() {
	m.stem = m.stems[rand.Intn(len(m.stems))]
	l := m.prefixList()
	m.pfx = l[rand.Intn(len(l))]
}

func (m model) reelHeight() int {
	h := m.h - chromeH - box.GetVerticalFrameSize()
	if m.w < mediumWidth {
		// Reels on top, detail underneath: they only get half the screen.
		h = (m.h-chromeH)/2 - box.GetVerticalFrameSize()
	}
	return max(1, h)
}

func scrollTop(cursor, top, h int) int {
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
	m.ptop = scrollTop(m.pi(), min(m.ptop, max(0, len(m.prefixList())-h)), h)
	m.st = scrollTop(m.si(), min(m.st, max(0, len(m.stemList())-h)), h)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.w, m.h = msg.Width, msg.Height
	case tea.KeyMsg:
		if m.testing {
			return m.updateTest(msg)
		}
		if m.searching {
			return m.updateSearch(msg)
		}
		return m.press(msg.String())
	case tea.MouseMsg:
		return m.click(msg)
	}
	return m, nil
}

// press handles one key, or one tap on the button that stands for it.
func (m model) press(key string) (tea.Model, tea.Cmd) {
	{
		cur, n, set := m.si(), len(m.stemList()), (*model).setSI
		if m.focus == 0 {
			cur, n, set = m.pi(), len(m.prefixList()), (*model).setPI
		}
		switch key {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		case "j", "down":
			cur = min(cur+1, n-1)
		case "k", "up":
			cur = max(cur-1, 0)
		case "ctrl+d":
			cur = min(cur+max(1, m.reelHeight()/2), n-1)
		case "ctrl+u":
			cur = max(cur-max(1, m.reelHeight()/2), 0)
		case "ctrl+f", "pgdown":
			cur = min(cur+m.reelHeight(), n-1)
		case "ctrl+b", "pgup":
			cur = max(cur-m.reelHeight(), 0)
		case "g", "home":
			cur = 0
		case "G", "end":
			cur = n - 1
		case "h", "left", "l", "right", "tab":
			m.focus = 1 - m.focus
		case " ", "s":
			m.spin()
		case "f":
			m.filtered = !m.filtered
			m.snap()
		case "t":
			m.testing = true
			m.deal()
		case "/":
			m.searching, m.query = true, ""
			m.saved = savedPos{m.pfx, m.stem}
			m.hits, m.hit = nil, 0
		}
		set(&m, cur)
		m.rescroll()
	}
	return m, nil
}

func (m model) updateSearch(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "enter":
		m.searching = false
	case "esc":
		m.searching = false
		m.pfx, m.stem = m.saved.pfx, m.saved.stem
		m.rescroll()
	case "down", "ctrl+n", "tab":
		if len(m.hits) > 0 {
			m.hit = (m.hit + 1) % len(m.hits)
			m.show()
		}
	case "up", "ctrl+p", "shift+tab":
		if len(m.hits) > 0 {
			m.hit = (m.hit + len(m.hits) - 1) % len(m.hits)
			m.show()
		}
	case "backspace":
		if m.query != "" {
			m.query = string([]rune(m.query)[:len([]rune(m.query))-1])
			m.search()
		}
	default:
		if msg.Type == tea.KeyRunes || msg.Type == tea.KeySpace {
			m.query += msg.String()
			m.search()
		}
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

// keyOf builds the KeyMsg a button stands for, so taps and keys take the same
// path through updateTest.
func keyOf(key string) tea.KeyMsg {
	switch key {
	case " ":
		return tea.KeyMsg{Type: tea.KeySpace, Runes: []rune(" ")}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	}
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
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
	reels := lipgloss.JoinHorizontal(lipgloss.Top, reel(m.prefixLabel(v.Prefix())), reel(v.Stem.Name))

	card := []string{reels, ""}
	if !m.revealed {
		card = append(card, "", metaStyle.Render("What does it mean? What are the forms?"))
	} else {
		card = append(card, m.answer(v))
	}
	body := lipgloss.Place(m.w, m.h-2, lipgloss.Center, lipgloss.Center,
		lipgloss.JoinVertical(lipgloss.Center, card...))
	return body + "\n" + m.buttonRow(m.chips())
}

// answer is the back of the card: one fixed-width block, everything ranged
// left on a label column, with air between the four things being said.
func (m model) answer(v *Verb) string {
	cw := min(64, max(28, m.w-8))
	label := lipgloss.NewStyle().Foreground(muted).Width(13)
	wrap := lipgloss.NewStyle().Width(cw - 2)

	row := func(name, value string) string {
		return label.Render(name) + value
	}
	block := func(heading, text string) []string {
		return []string{headingStyle.Render(heading), wrap.Render("  " + text), ""}
	}

	present, past, perfect := v.Forms()
	head := wordStyle.Render(v.Name)
	if k := kindOf(*v); true {
		head = pad(head, cw-lipgloss.Width(k)) + k
	}
	rows := []string{
		head,
		trackStyle.Render(strings.Repeat("─", cw)),
		"",
		row("er/sie/es", formStyle.Render(present)),
		row("präteritum", formStyle.Render(past)),
		row("perfekt", formStyle.Render(perfect)),
		"",
		row("rektion", useStyle.Render(v.Use)),
		row("nebensatz", bodyStyle.Render(v.Nebensatz())),
		"",
	}
	rows = append(rows, block("offiziell", bodyStyle.Render(v.Official))...)
	rows = append(rows, block("umgangssprachlich", bodyStyle.Render(v.Colloquial))...)
	rows = append(rows, block("beispiel", exampleStyle.Render(v.Example))...)

	// Left-aligned inside a block of known width, so the card as a whole can be
	// centred without the text going ragged.
	for i, r := range rows {
		rows[i] = pad(r, cw)
	}
	return strings.Join(rows, "\n")
}

// buttonRow lays the chips out with the gaps chips() already measured, so a
// tap lands on the button it looks like it lands on.
func (m model) buttonRow(cs []chip) string {
	var b strings.Builder
	for _, c := range cs {
		for lipgloss.Width(b.String()) < c.x0 {
			b.WriteString(" ")
		}
		b.WriteString(chipStyle.Render(c.label))
	}
	return trunc(b.String(), m.w)
}

// geometry mirrors what View draws, so a tap can be turned back into a row.
type geometry struct {
	reelY    int // screen row of the first reel item
	pX0, pX1 int // columns covered by the prefix pane
	sX0, sX1 int
	footerY  int
}

func (m model) geometry() geometry {
	rh := m.reelHeight()
	g := geometry{reelY: 2, footerY: 1 + rh + 2}
	switch {
	case m.w >= mediumWidth:
		g.pX0, g.pX1 = 0, prefixWidth+border-1
		g.sX0, g.sX1 = g.pX1+1, g.pX1+stemWidth+border
	default:
		half := (m.w - 2*border) / 2
		g.pX0, g.pX1 = 0, half+border-1
		g.sX0, g.sX1 = g.pX1+1, m.w-1
		dh := max(1, m.h-chromeH-rh-2*box.GetVerticalFrameSize())
		g.footerY += dh + 2
	}
	return g
}

func (m model) click(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if m.testing {
		if msg.Action != tea.MouseActionPress || msg.Button != tea.MouseButtonLeft {
			return m, nil
		}
		if msg.Y == m.h-1 {
			for _, c := range m.chips() {
				if msg.X >= c.x0 && msg.X <= c.x1 {
					return m.updateTest(keyOf(c.key))
				}
			}
			return m, nil
		}
		return m.updateTest(keyOf(" ")) // a tap on the card itself flips it
	}
	g := m.geometry()
	onPrefix := msg.X >= g.pX0 && msg.X <= g.pX1
	onStem := msg.X >= g.sX0 && msg.X <= g.sX1

	switch msg.Button {
	case tea.MouseButtonWheelUp, tea.MouseButtonWheelDown:
		if m.searching {
			return m, nil
		}
		// A swipe on a phone arrives as a wheel event wherever the finger was,
		// so a wheel away from the reels still spins the focused one.
		if onPrefix {
			m.focus = 0
		} else if onStem {
			m.focus = 1
		}
		if msg.Button == tea.MouseButtonWheelUp {
			return m.press("k")
		}
		return m.press("j")
	case tea.MouseButtonLeft:
		if msg.Action != tea.MouseActionPress {
			return m, nil
		}
		if msg.Y == g.footerY {
			for _, c := range m.chips() {
				if msg.X >= c.x0 && msg.X <= c.x1 {
					if m.searching {
						m.searching = false
					}
					return m.press(c.key)
				}
			}
			return m, nil
		}
		if m.searching {
			return m, nil
		}
		if i := msg.Y - g.reelY; i >= 0 {
			switch {
			case onPrefix && m.ptop+i < len(m.prefixList()):
				m.focus = 0
				m.setPI(m.ptop + i)
			case onStem && m.st+i < len(m.stemList()):
				m.focus = 1
				m.setSI(m.st + i)
			}
			m.rescroll()
		}
	}
	return m, nil
}

func (m model) View() string {
	if m.h == 0 {
		return ""
	}
	if m.testing {
		return m.testView()
	}
	if m.w < 20 || m.h < 8 {
		return "terminal too small\n"
	}
	rh := m.reelHeight()
	pane := func(focused bool, w, h int, s string) string {
		b := box
		if focused {
			b = boxFocused
		}
		return b.Width(w).Height(h).Render(clamp(s, h))
	}
	var prefixes, stems []string
	for _, p := range m.prefixList() {
		prefixes = append(prefixes, m.prefixLabel(p))
	}
	for _, s := range m.stemList() {
		stems = append(stems, s.Name)
	}
	pl, sl := m.prefixList(), m.stemList()
	alive := func(ok func(int) bool) func(int) bool {
		if m.filtered {
			return nil // everything on screen is a real word already
		}
		return ok
	}
	prefixPane := func(w int) string {
		live := alive(func(i int) bool { return m.exists(pl[i], m.stem) })
		return pane(m.focus == 0, w, rh, reel(prefixes, m.pi(), m.ptop, w, rh, m.focus == 0, live))
	}
	stemPane := func(w int) string {
		live := alive(func(i int) bool { return m.exists(m.pfx, sl[i]) })
		return pane(m.focus == 1, w, rh, reel(stems, m.si(), m.st, w, rh, m.focus == 1, live))
	}

	var body string
	switch {
	case m.w >= wideWidth:
		rw := m.w - prefixWidth - stemWidth - centerWidth - 4*border
		body = lipgloss.JoinHorizontal(lipgloss.Top,
			prefixPane(prefixWidth), stemPane(stemWidth),
			pane(false, centerWidth, rh, m.forms(centerWidth)),
			pane(false, rw, rh, m.meanings(rw)))
	case m.w >= mediumWidth:
		dw := m.w - prefixWidth - stemWidth - 3*border
		body = lipgloss.JoinHorizontal(lipgloss.Top,
			prefixPane(prefixWidth), stemPane(stemWidth),
			pane(false, dw, rh, m.compact(dw)))
	default:
		half := (m.w - 2*border) / 2
		reels := lipgloss.JoinHorizontal(lipgloss.Top, prefixPane(half), stemPane(m.w-half-2*border))
		dh := max(1, m.h-chromeH-rh-2*box.GetVerticalFrameSize())
		dw := m.w - border
		body = reels + "\n" + pane(false, dw, dh, m.compact(dw))
	}
	return m.header() + "\n" + body + "\n" + m.footer()
}

// clamp keeps a pane from pushing the layout around when the content is taller
// than the space it was given.
func clamp(s string, h int) string {
	lines := strings.Split(s, "\n")
	if len(lines) <= h {
		return s
	}
	if h < 2 {
		return lines[0]
	}
	return strings.Join(append(lines[:h-1], metaStyle.Render("…")), "\n")
}

func (m model) header() string {
	word := m.pfx + m.stem.Name
	left := wordStyle.Render(word)
	if _, real := m.current(); !real {
		left = ghostStyle.Render(word + "  (kein Wort)")
	}
	right := kickerStyle.Render(fmt.Sprintf("%d Verben", m.count))
	if m.w >= mediumWidth {
		right += metaStyle.Render(fmt.Sprintf(" aus %d Stämmen", len(m.stems))) +
			metaStyle.Render("  ·  ") + m.ribbon()
	}
	gap := m.w - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		return trunc(left, m.w)
	}
	return left + strings.Repeat(" ", gap) + right
}

func (m model) footer() string {
	if m.searching {
		line := keyStyle.Render("/") + bodyStyle.Render(m.query) + wordStyle.Render("█")
		switch {
		case m.query == "":
			line += footerStyle.Render("  type to search prefixes, verbs and meanings")
		case len(m.hits) == 0:
			line += footerStyle.Render("  no match")
		default:
			line += footerStyle.Render(fmt.Sprintf("  %d/%d  ↑↓ cycle · enter keep · esc cancel", m.hit+1, len(m.hits)))
		}
		return trunc(line, m.w)
	}
	return m.buttonRow(m.chips())
}

// A chip is one footer button: a hint for the keyboard and a tap target for a
// touchscreen, where nobody is going to press ^d.
type chip struct {
	label, key string
	x0, x1     int // filled in by chips(), for hit testing
}

func (m model) chips() []chip {
	if m.testing {
		cs := []chip{{label: "  space reveal  ", key: " "}, {label: "  n next  ", key: "n"},
			{label: "  esc back  ", key: "esc"}, {label: "  q quit  ", key: "q"}}
		if m.revealed {
			cs[0].label = "  space next card  "
		}
		return measure(cs, m.w)
	}
	cs := []chip{
		{label: "◀▶ h/l", key: "tab"},
		{label: "▲ k", key: "k"},
		{label: "▼ j", key: "j"},
		{label: "« ^b", key: "ctrl+b"},
		{label: "» ^f", key: "ctrl+f"},
		{label: "/ search", key: "/"},
		{label: "space random", key: " "},
		{label: "f " + m.filterLabel(), key: "f"},
		{label: "t test", key: "t"},
		{label: "q quit", key: "q"},
	}
	return measure(cs, m.w)
}

// measure assigns each chip its columns, dropping the least essential ones
// until the row fits.
func measure(cs []chip, w int) []chip {
	for {
		x := 0
		for i := range cs {
			cs[i].x0 = x
			x += lipgloss.Width(cs[i].label) + 2 // the chip's own padding
			cs[i].x1 = x - 1
			x += 1 // the gap between chips
		}
		if x-1 <= w || len(cs) <= 4 {
			// Centre the row in whatever space is left, hit boxes included.
			if pad := (w - (x - 1)) / 2; pad > 0 {
				for i := range cs {
					cs[i].x0 += pad
					cs[i].x1 += pad
				}
			}
			return cs
		}
		cs = append(cs[:3], cs[4:]...)
	}
}

// hyperlink wraps text in OSC 8, which terminals that support it render as a
// clickable link. Terminals that do not simply ignore the escape.
func hyperlink(text, url string) string {
	return "\x1b]8;;" + url + "\x1b\\" + text + "\x1b]8;;\x1b\\"
}

// ribbon is the contribute link that sits at the top right, beside the counts.
// The badge only appears when there is room for it.
func (m model) ribbon() string {
	link := hyperlink(linkStyle.Render("github.com/progapandist/tja"), repoURL)
	if m.w >= wideWidth {
		return badgeStyle.Render(" contribute ") + " " + link
	}
	return link
}

func (m model) filterLabel() string {
	if m.filtered {
		return "all pairs"
	}
	return "real only"
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

// reel renders one column of the slot machine plus its scrollbar.
func reel(items []string, cur, top, w, h int, focused bool, alive func(int) bool) string {
	lw := max(1, w-padding-2) // the bar takes a gap column and the bar itself
	var b strings.Builder
	for i := top; i < len(items) && i < top+h; i++ {
		style := reelStyle
		if alive != nil && !alive(i) {
			style = deadStyle
		}
		if i == cur {
			style = restingStyle
			if focused {
				style = pickedStyle
			}
		}
		b.WriteString(style.Render(pad(" "+items[i], lw)))
		b.WriteString(scrollbarRow(i-top, h, len(items), top))
		b.WriteString("\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

// scrollbarRow draws row i of a track h tall for total items scrolled to top.
// The thumb is sized to the share of the list on screen; blank when it all fits.
func scrollbarRow(i, h, total, top int) string {
	if total <= h {
		return "  "
	}
	thumb := max(1, h*h/total)
	pos := (h - thumb) * top / (total - h)
	if i >= pos && i < pos+thumb {
		return " " + thumbStyle.Render("┃")
	}
	return " " + trackStyle.Render("│")
}

// separability spells out in words what the trennbar/untrennbar badge shows.
func separability(v Verb) string {
	switch {
	case v.Prefix() == "":
		return "kein Präfix zu trennen"
	case v.Sep:
		return "trennbar: im Hauptsatz hinten, im Nebensatz wieder am Stamm"
	}
	return "untrennbar: bleibt immer am Stamm, kein ge- im Partizip"
}

func kindOf(v Verb) string {
	switch {
	case v.Prefix() == "":
		return metaStyle.Render("stamm")
	case v.Sep:
		return sepStyle.Render("trennbar")
	}
	return insepStyle.Render("untrennbar")
}

func (m model) forms(w int) string {
	s := m.stem
	var rows []string
	if v, real := m.current(); !real {
		present, past, perfect := v.Forms()
		rows = append(rows,
			ghostStyle.Render(v.Name),
			metaStyle.Render("kein belegtes Wort — so ginge es aber"),
			"",
			ghostStyle.Render("er/sie/es   "+present),
			ghostStyle.Render("präteritum  "+past),
			ghostStyle.Render("perfekt     "+perfect),
			"",
			metaStyle.Render("f zeigt nur echte Kombinationen"),
			"",
		)
	}
	for _, v := range m.matches() {
		present, past, perfect := v.Forms()
		line := func(label, val string) string {
			return metaStyle.Render(pad(label, 12)) + formStyle.Render(val)
		}
		rows = append(rows,
			wordStyle.Render(v.Name),
			kindOf(*v)+metaStyle.Render("   "+dash(v.Prefix())+" + "+s.Name),
			"",
			headingStyle.Render("wo der stamm sich ändert"),
			line("er/sie/es", present),
			line("präteritum", past),
			line("perfekt", perfect),
			"",
			headingStyle.Render("rektion"),
			useStyle.Render(v.Use),
			"",
			headingStyle.Render("im nebensatz"),
			bodyStyle.Render(v.Nebensatz()),
			metaStyle.Render(separability(*v)),
			"",
		)
	}
	rows = append(rows,
		headingStyle.Render("stammformen"),
		metaStyle.Render(s.Name+" · "+s.Present+" · "+s.Past+" · "+s.Aux+" "+s.PartII),
		metaStyle.Render(s.Gloss),
	)
	return wrapAll(rows, w)
}

// compact is the narrow-screen detail pane: the forms squeezed onto one line
// so the meanings still fit underneath.
func (m model) compact(w int) string {
	v, real := m.current()
	present, past, perfect := v.Forms()
	head := wordStyle.Render(v.Name) + "   " + kindOf(v)
	if !real {
		head = ghostStyle.Render(v.Name) + "   " + metaStyle.Render("kein belegtes Wort")
	}
	rows := []string{
		head,
		formStyle.Render(present + " · " + past + " · " + perfect),
		useStyle.Render(v.Use),
		bodyStyle.Render(v.Nebensatz()),
		"",
	}
	return wrapAll(rows, w) + "\n" + m.meanings(w)
}

func (m model) meanings(w int) string {
	hits := m.matches()
	if len(hits) == 0 {
		var real []string
		for _, v := range m.stem.Verbs {
			real = append(real, dash(v.Prefix()))
		}
		return wrapAll([]string{
			headingStyle.Render("es gibt stattdessen"),
			bodyStyle.Render(strings.Join(real, ", ") + " + " + m.stem.Name),
		}, w)
	}
	var rows []string
	for _, v := range hits {
		if len(hits) > 1 {
			rows = append(rows, wordStyle.Render(v.Name)+"  "+kindOf(*v))
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
	wrap := lipgloss.NewStyle().Width(w - padding)
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
	if _, err := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
