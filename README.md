# tja

A one-armed-bandit TUI for German verbs: two reels, prefixes and stems, that
filter each other. Whatever they land on is always a real word — spinning the
reels *is* the vocabulary list.

```
go install github.com/progapandist/tja@latest   # or, in a clone:
go run .
```

Prebuilt binaries for macOS and Linux (amd64/arm64) are on the
[releases page](https://github.com/progapandist/tja/releases).

## The panes

- **Prefix reel** — only the prefixes that make a word with the stem showing.
  Separable ones carry the dictionary hyphen (`auf-` vs `be`), inferred from the data.
- **Stem reel** — only the stems that take the prefix showing.
- **Forms** — just the places where the root actually changes: 3rd person singular,
  preterite, perfect with its auxiliary. Plus the rection: which cases and
  prepositions the verb governs.
- **Meanings** — the official meaning, the colloquial or idiomatic one, and an example.

The forms pane also generates a subordinate clause for every verb — the one place
where a separable prefix rejoins its stem: main clause `ruft … an`, Nebensatz
`…, weil sie mich anruft.` The object comes from the rection, so the clause is
something you could actually say.

The header carries the verb count and a clickable repo link (OSC 8). Both reels carry a scrollbar. The layout has three shapes: four panes at 100
columns or more, reels plus one combined pane down to 64, and reels above a
detail pane below that.

`j/k` spin · `h/l` switch reel · `^d/^u` half page · `^f/^b` page · `g/G` ends ·
`/` search · `f` filtering on/off · `space` random combination · `t` flash cards · `q` quit

`f` turns the mutual filtering off: both reels then list everything, prefixes
that make no word with the stem showing (and vice versa) are dimmed rather than
hidden, and landing between two words shows the form the rules *would* build,
marked as not attested. `f` again snaps back to a real word.

## Mouse and touch

The reels scroll with the wheel and select on click or tap. The footer is a row
of buttons — `⇄` switches reel, `▲ ▼` spin, `⇞ ⇟` page, and the rest mirror the
keys — so the whole thing is usable on a phone without a keyboard. In test mode
a tap anywhere reveals the card or deals the next one.

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
