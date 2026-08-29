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
