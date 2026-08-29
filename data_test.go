package main

import "testing"

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
