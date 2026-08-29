package main

import (
	_ "embed"
	"strings"
)

// verbs.txt is a pipe-delimited flat file. Lines starting with "=" open a stem
// group and carry its principal parts; the lines under it are prefixed verbs.
//
//	=stem|gloss|präsens 3.sg|präteritum|partizip II|aux
//	verb|t or f (separable)|official|colloquial|example|use (case/preposition)|aux override
//
// ponytail: a text file and a Split, not JSON with a schema. Hand-editable.
//
//go:embed verbs.txt
var raw string

type Stem struct {
	Name, Gloss           string
	Present, Past, PartII string
	Aux                   string
	Verbs                 []Verb
}

type Verb struct {
	Name       string
	Sep        bool
	Official   string
	Colloquial string
	Example    string
	Use        string // rection: cases and prepositions the verb takes
	Aux        string
	Stem       *Stem
}

// Prefix is what the stem was decorated with ("" for the bare stem).
func (v Verb) Prefix() string { return strings.TrimSuffix(v.Name, v.Stem.Name) }

// Forms returns the three places where the root actually changes: 3rd person
// singular present, preterite, and the perfect with its auxiliary.
func (v Verb) Forms() (present, past, perfect string) {
	p, s := v.Prefix(), v.Stem
	switch {
	case p == "":
		present, past, perfect = s.Present, s.Past, s.PartII
	case v.Sep:
		// Separable prefixes hop to the end of the main clause, and ge- lands
		// between prefix and root: "nimmt ... an", "angenommen".
		present = s.Present + " … " + p
		past = s.Past + " … " + p
		perfect = p + s.PartII
	default:
		present, past = p+s.Present, p+s.Past
		perfect = p + strings.TrimPrefix(s.PartII, "ge")
	}
	return present, past, v.aux() + " " + perfect
}

func (v Verb) aux() string {
	if v.Aux != "" {
		return v.Aux
	}
	return v.Stem.Aux
}

func load() []*Stem {
	var stems []*Stem
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		f := strings.Split(line, "|")
		if strings.HasPrefix(line, "=") {
			stems = append(stems, &Stem{
				Name: strings.TrimPrefix(f[0], "="), Gloss: f[1],
				Present: f[2], Past: f[3], PartII: f[4], Aux: f[5],
			})
			continue
		}
		s := stems[len(stems)-1]
		v := Verb{Name: f[0], Sep: f[1] == "t", Official: f[2], Colloquial: f[3], Example: f[4], Use: f[5], Stem: s}
		if len(f) > 6 {
			v.Aux = f[6]
		}
		s.Verbs = append(s.Verbs, v)
	}
	return stems
}
