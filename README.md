# tja

A one-armed-bandit TUI for German verbs: two reels, prefixes and stems, that
filter each other. Whatever they land on is always a real word — spinning the
reels *is* the vocabulary list.

```
go run .
```

## The panes

- **Prefix reel** — only the prefixes that make a word with the stem showing.
  Separable ones carry the dictionary hyphen (`auf-` vs `be`), inferred from the data.
- **Stem reel** — only the stems that take the prefix showing.
- **Forms** — just the places where the root actually changes: 3rd person singular,
  preterite, perfect with its auxiliary. Plus the rection: which cases and
  prepositions the verb governs.
- **Meanings** — the official meaning, the colloquial or idiomatic one, and an example.

Both reels carry a scrollbar. The layout has three shapes: four panes at 100
columns or more, reels plus one combined pane down to 64, and reels above a
detail pane below that.

`j/k` spin · `h/l` switch reel · `^d/^u` half page · `^f/^b` page · `g/G` ends ·
`/` search · `space` random combination · `t` flash cards · `q` quit

## Search

`/` searches prefixes, stems and meanings at once, fuzzily: `mitneh` finds
*mitnehmen*, `zusbrech` finds *zusammenbrechen*, `burgle` finds *einbrechen*
through its English gloss. Umlauts are optional — `uber` and `ueber` both find
*übernehmen*. `↑↓` cycles hits, `enter` keeps the one showing, `esc` puts the
reels back.

## Test mode

`t` deals a random real combination as a card: one prefix, one stem, nothing else.
`space` reveals forms, rection and meanings, `space`/`n` deals the next one,
`esc` returns to the reels — parked on the card you just saw.

## Data

Everything lives in `verbs.txt`, one verb per line, pipe-delimited:

```
=stem|gloss|present 3sg|preterite|participle|aux
verb|separable t/f|official|colloquial|example|use|aux override
```

90 stems, ~790 verbs, chosen from the high-frequency verbs where a prefix
genuinely shifts the meaning. Adding more needs no code: append lines.

Conjugated forms are derived, not stored — separable prefixes detach and swallow
the *ge-* (`nimmt … an`, `angenommen`), inseparable ones replace it
(`übernimmt`, `übernommen`). See `Verb.Forms` in `data.go`.
