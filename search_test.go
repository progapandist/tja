package main

import "testing"

// The point of the search is that one query can span prefix and stem.
func TestSearch(t *testing.T) {
	for _, c := range []struct{ q, want string }{
		{"mitneh", "mitnehmen"},
		{"übernehm", "übernehmen"},
		{"xqbrechen", ""}, // x and q appear nowhere: no match at all
		{"zusbrech", "zusammenbrechen"},
		{"anrufen", "anrufen"},
		{"burgle", "einbrechen"},    // meanings are searched too
		{"ubernehm", "übernehmen"},  // typed without the umlaut
		{"uebernehm", "übernehmen"}, // or spelled out
	} {
		m := newModel()
		m.query = c.q
		m.search()
		got := ""
		if len(m.hits) > 0 {
			got = m.hits[0].Name
		}
		if got != c.want {
			t.Errorf("%q: got %q, want %q", c.q, got, c.want)
		}
	}
}
