# tja

A three-pane TUI for German verbs, grouped by stem.

Left: every prefixed verb built on a stem (`nehmen` → `annehmen`, `mitnehmen`, `übernehmen`, …).
Centre: only the forms where the root actually changes — 3rd person singular, preterite, perfect —
plus whether the prefix is separable.
Right: the official meaning, the colloquial/idiomatic one, and an example.

```
go run .          # needs ~72 columns
```

`j/k` verb · `n/p` stem · `/` search · `x` clear · `q` quit

## Data

Everything lives in `verbs.txt`, one verb per line, pipe-delimited:

```
=stem|gloss|present 3sg|preterite|participle|aux
verb|separable t/f|official|colloquial|example|aux override
```

Conjugated forms are derived from the stem, not stored — see `Verb.Forms` in `data.go`.
28 stems, ~310 verbs, chosen from the high-frequency verbs where a prefix genuinely
shifts the meaning.
