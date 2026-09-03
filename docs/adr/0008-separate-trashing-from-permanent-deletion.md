# Separate trashing from permanent deletion

- **Status:** accepted
- **Date:** 2026-08-31

## Context

The original brief for protui specified a single delete action, on the stated
premise that "upstream delete is permanent, no trash".

Half of that is right. `pass-cli item delete` is permanent and prompts for
nothing. But `item trash` and `item untrash` both exist, and `item list
--filter-state trashed` lists what is in the trash. The recoverable path was
available all along.

That matters because the two actions differ in exactly the way that should drive
interface design: one is reversible and one destroys an SSH key that may be the
only copy of a credential. A single `d` key bound to the irreversible one, with
a `y/n` prompt, gives the same two keystrokes to "recoverable mistake" and
"permanent loss".

A single confirmation is also weak protection for the destructive case. `d` then
`y` is a common muscle-memory pair; it can be typed faster than it can be
reconsidered.

## Decision

We will expose both operations on separate keys, with confirmation strength
matched to reversibility:

- `d` moves the item to the trash after a `y/n` confirmation. The prompt says
  it can be restored with `pass-cli item untrash`.
- `D` deletes permanently. The prompt states that the key is destroyed rather
  than trashed, and requires the user to **type the item's exact title** before
  it will proceed. It also mentions that `d` trashes instead.

The default action — the lowercase key, the one reachable by habit — is the
recoverable one.

## Consequences

The common case is safe. A mistaken `d` costs an `untrash` from the CLI.

The destructive case cannot be performed by muscle memory. Typing a title is
deliberately slow and forces the user to read which item is selected, which also
catches the "wrong row highlighted" error that a `y/n` prompt does not.

Two keys and two prompts is more UI than one. The alternative was to make every
deletion permanent, which was the brief's assumption and was based on a
misreading of the upstream surface.

protui does not list or restore trashed items in v1: `--filter-state active` is
passed when listing, so a trashed key disappears from the UI. Restoring is
`pass-cli item untrash`, which the trash prompt names. Adding a trash view is a
later decision, not a gap this one leaves open.

`item trash` also accepts `--item-title`, but protui addresses items by
`--share-id` and `--item-id` throughout, since titles are not unique and
resolution by title is ambiguous.
