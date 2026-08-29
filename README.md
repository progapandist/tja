# tja

A three-pane TUI for German verbs, grouped by stem.

Two reels — prefixes and stems — spin independently, one-armed-bandit style.
Whatever they land on is spelled out at the top right: a real word, or a greyed-out
one that does not exist (with its forms shown anyway, because the rules still apply).
Prefixes that make no word with the stem currently showing are dimmed, and vice versa.

Centre: only the forms where the root actually changes — 3rd person singular, preterite,
perfect — plus whether the prefix is separable. Right: official meaning, colloquial
meaning, example.

```
go run .          # needs ~96 columns
```

`j/k` spin · `h/l` switch reel · `J/K` next prefix that makes a real word ·
`space` random combination · `t` flash cards · `q` quit

## Test mode

`t` deals a random real combination as a card: one prefix, one stem, nothing else.
`space` reveals forms and meanings, `space`/`n` deals the next one, `esc` returns to
the reels — parked on the card you just saw.

## Data

Everything lives in `verbs.txt`, one verb per line, pipe-delimited:

```
=stem|gloss|present 3sg|preterite|participle|aux
verb|separable t/f|official|colloquial|example|aux override
```

Conjugated forms are derived from the stem, not stored — see `Verb.Forms` in `data.go`.
28 stems, ~310 verbs, chosen from the high-frequency verbs where a prefix genuinely
shifts the meaning.
