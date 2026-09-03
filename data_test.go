package main

import (
	"sort"
	"strings"
	"testing"
)

// The prefix rules are the only real logic here: separable prefixes detach and
// swallow the ge-, inseparable ones replace it.
func TestForms(t *testing.T) {
	want := map[string][3]string{
		"annehmen":      {"nimmt … an", "nahm … an", "hat angenommen"},
		"übernehmen":    {"übernimmt", "übernahm", "hat übernommen"},
		"nehmen":        {"nimmt", "nahm", "hat genommen"},
		"bekommen":      {"bekommt", "bekam", "hat bekommen"},
		"ankommen":      {"kommt … an", "kam … an", "ist angekommen"},
		"gefallen":      {"gefällt", "gefiel", "hat gefallen"},
		"herunterladen": {"lädt … herunter", "lud … herunter", "hat heruntergeladen"},
		"verbringen":    {"verbringt", "verbrachte", "hat verbracht"},
		"überlegen":     {"überlegt", "überlegte", "hat überlegt"},
		"aufstehen":     {"steht … auf", "stand … auf", "ist aufgestanden"},
	}
	found := 0
	for _, s := range load() {
		for _, v := range s.Verbs {
			w, ok := want[v.Name]
			if !ok {
				continue
			}
			found++
			p, pa, pe := v.Forms()
			if got := [3]string{p, pa, pe}; got != w {
				t.Errorf("%s: got %v, want %v", v.Name, got, w)
			}
		}
	}
	if found != len(want) {
		t.Errorf("checked %d verbs, expected %d", found, len(want))
	}
}

// The clause is generated, so the interesting part is that the prefix rejoins
// the stem and the object matches the rection.
func TestNebensatz(t *testing.T) {
	want := map[string]string{
		"anrufen":    "…, weil sie mich anruft.",
		"aufstehen":  "…, weil sie aufsteht.",
		"teilnehmen": "…, weil sie daran teilnimmt.",
		"benehmen":   "…, weil sie sich benimmt.",
		"übernehmen": "…, weil sie es übernimmt.",
		"vorwerfen":  "…, weil sie mir das vorwirft.",
	}
	seen := 0
	for _, s := range load() {
		for _, v := range s.Verbs {
			if w, ok := want[v.Name]; ok {
				seen++
				if got := v.Nebensatz(); got != w {
					t.Errorf("%s: got %q, want %q", v.Name, got, w)
				}
			}
		}
	}
	if seen != len(want) {
		t.Errorf("checked %d, expected %d", seen, len(want))
	}
}

// A verb that governs a preposition has to keep it. Reading the rection by
// searching for a token anywhere inside it dropped the preposition ("mit
// jdm+D" came out a bare "mir") and picked the wrong object when the pattern
// listed two slots.
func TestNebensatzKeepsPreposition(t *testing.T) {
	want := map[string]string{
		// A person keeps the preposition and takes a pronoun.
		"mitkommen":     "…, weil sie mit mir mitkommt.",
		"vorbeischauen": "…, weil sie bei mir vorbeischaut.",
		"zukommen":      "…, weil sie auf mich zukommt.",
		"einspringen":   "…, weil sie für mich einspringt.",
		// A thing takes the da- compound: German has no "mit es".
		"auskommen":  "…, weil sie damit auskommt.",
		"nachdenken": "…, weil sie darüber nachdenkt.",
		"verstoßen":  "…, weil sie dagegen verstößt.",
		"abhängen":   "…, weil sie davon abhängt.",
		// The first slot is the object, not the first one that happens to match.
		"besprechen":  "…, weil sie es bespricht.",
		"weitergeben": "…, weil sie es weitergibt.",
		// Nothing to stand in for leaves the clause bare.
		"zunehmen": "…, weil sie zunimmt.",
		"gedenken": "…, weil sie gedenkt.",
	}
	seen := 0
	for _, s := range load() {
		for _, v := range s.Verbs {
			if w, ok := want[v.Name]; ok {
				seen++
				if got := v.Nebensatz(); got != w {
					t.Errorf("%s: got %q, want %q", v.Name, got, w)
				}
			}
		}
	}
	if seen != len(want) {
		t.Errorf("checked %d, expected %d", seen, len(want))
	}
	// And nothing leaks the rection notation into the clause.
	for _, s := range load() {
		for _, v := range s.Verbs {
			for _, bad := range []string{"jdm", "jdn", "jds", "etw", "+A", "+D", "+G", "·"} {
				if strings.Contains(v.Nebensatz(), bad) {
					t.Errorf("%s: rection leaked into %q", v.Name, v.Nebensatz())
				}
			}
		}
	}
}

// One spelling, both separabilities, two meanings: umfahren is the standard
// example. Both have to be reachable, which a reel keyed on the prefix alone
// cannot manage.
func TestHomographsAreDistinct(t *testing.T) {
	stems := load()
	byName := map[string][]*Verb{}
	for _, s := range stems {
		for i := range s.Verbs {
			v := &s.Verbs[i]
			byName[v.Name] = append(byName[v.Name], v)
		}
	}
	var pairs []string
	for name, vs := range byName {
		if len(vs) < 2 {
			continue
		}
		pairs = append(pairs, name)
		if len(vs) != 2 {
			t.Errorf("%s: %d entries, expected 2", name, len(vs))
			continue
		}
		// Split only ever follows separability; two senses of one kind belong
		// in a single entry.
		if vs[0].Sep == vs[1].Sep {
			t.Errorf("%s: both entries are sep=%v, so they are not a pair", name, vs[0].Sep)
		}
		if vs[0].Official == vs[1].Official {
			t.Errorf("%s: both entries mean the same thing", name)
		}
	}
	sort.Strings(pairs)
	if got := strings.Join(pairs, ","); got != "umfahren,umgehen,überfahren,übersetzen" {
		t.Errorf("homograph pairs: got %q", got)
	}
}

// The reel gives each of a pair its own row, and stepping between them changes
// the verb without changing the spelling.
func TestReelReachesBothSenses(t *testing.T) {
	m := newModel()
	for _, s := range m.stems {
		if s.Name == "fahren" {
			m.stem = s
		}
	}
	var rows []int
	for i, r := range m.prefixList() {
		if r.p == "um" {
			rows = append(rows, i)
		}
	}
	if len(rows) != 2 {
		t.Fatalf("um- on fahren: %d rows, expected 2", len(rows))
	}
	seen := map[string]bool{}
	for _, i := range rows {
		m.setPI(i)
		v, real := m.current()
		if !real || v.Name != "umfahren" {
			t.Fatalf("row %d: got %q real=%v", i, v.Name, real)
		}
		if m.pi() != i {
			t.Errorf("row %d: pi() came back %d", i, m.pi())
		}
		seen[v.Official] = true
	}
	if len(seen) != 2 {
		t.Errorf("both rows gave the same meaning: %v", seen)
	}
}

// Separable über- and inseparable über are different prefixes that happen to
// be spelled alike, and they take different stems. Matching the text alone
// offered übernehmen while über- was selected, then switched the reel to the
// inseparable one behind your back.
func TestStemReelRespectsSeparability(t *testing.T) {
	m := newModel()
	for _, s := range m.stems {
		if s.Name == "setzen" {
			m.stem = s
		}
	}
	var sepRow, insepRow = -1, -1
	for i, r := range m.prefixList() {
		if r.p != "über" {
			continue
		}
		if r.sep {
			sepRow = i
		} else {
			insepRow = i
		}
	}
	if sepRow < 0 || insepRow < 0 {
		t.Fatalf("über on setzen: sep row %d, insep row %d", sepRow, insepRow)
	}

	names := func() []string {
		var out []string
		for _, s := range m.stemList() {
			out = append(out, s.Name)
		}
		return out
	}
	has := func(xs []string, want string) bool {
		for _, x := range xs {
			if x == want {
				return true
			}
		}
		return false
	}

	m.setPI(sepRow)
	withSep := names()
	if has(withSep, "nehmen") {
		t.Errorf("separable über- offered nehmen, but übernehmen is inseparable: %v", withSep)
	}
	m.setPI(insepRow)
	withInsep := names()
	if !has(withInsep, "nehmen") {
		t.Errorf("inseparable über dropped nehmen: %v", withInsep)
	}
	if len(withInsep) <= len(withSep) {
		t.Errorf("inseparable über takes %d stems, separable %d", len(withInsep), len(withSep))
	}
}

// Whatever the reel offers, picking it has to land there: the prefix and its
// separability both survive a move on the stem reel.
func TestPickingStemKeepsPrefix(t *testing.T) {
	base := newModel()
	moves := 0
	for _, seed := range []string{"setzen", "fahren", "gehen", "nehmen"} {
		for p := 0; ; p++ {
			m := base
			for _, s := range m.stems {
				if s.Name == seed {
					m.stem = s
				}
			}
			if p >= len(m.prefixList()) {
				break
			}
			m.setPI(p)
			before, real := m.current()
			if !real {
				continue
			}
			for i := range m.stemList() {
				m2 := m
				m2.setSI(i)
				got, ok := m2.current()
				moves++
				if !ok || got.Prefix() != before.Prefix() || got.Sep != before.Sep {
					t.Errorf("%s row %d -> stem %d: had %s (sep=%v), got %s (sep=%v)",
						seed, p, i, before.Name, before.Sep, got.Name, got.Sep)
				}
			}
		}
	}
	if moves < 50 {
		t.Errorf("only %d moves checked", moves)
	}
}

// Every card the test mode deals has to be a word someone can be asked about.
// Unfiltered, the prefix reel carries every prefix, greyed where it does not
// pair, and spinning to a random row landed on combinations like
// "herausschicken": forms and a Nebensatz, but nothing under offiziell,
// umgangssprachlich or beispiel.
func TestSpinLandsOnARealWord(t *testing.T) {
	m := newModel()
	m.filtered = false
	for i := 0; i < 2000; i++ {
		m.spin()
		v, real := m.current()
		if !real {
			t.Fatalf("spin dealt a ghost: %s", v.Name)
		}
		if v.Official == "" || v.Example == "" {
			t.Fatalf("%s has no meaning to reveal", v.Name)
		}
	}
}
