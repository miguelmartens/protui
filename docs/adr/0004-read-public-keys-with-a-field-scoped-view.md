# Read public keys with a field-scoped view

- **Status:** accepted
- **Date:** 2026-08-31

## Context

protui shows a public key, its fingerprint, its algorithm and its comment. It
never needs a private key: the SSH agent daemon handles signing inside
`pass-cli`, in a separate process.

The obvious way to fetch an item's detail is `pass-cli item view --output json`.
Reading the upstream source shows what that actually returns
(`pass-cli/src/commands/item/view.rs:154-157`, as of upstream `51a4c9b1`): the
whole `Item` struct, serialised with no redaction, including
`content.content.SshKey.private_key`.

So the natural call hands protui the user's private key on every detail view.
Nothing would visibly break — the UI would render only the public fields — but
the private key would be in protui's address space, in a decoded JSON struct,
for every key the user looks at. It would reach any error message that echoed
raw output, any debug logging added later, and any crash dump.

`pass-cli` also offers `item view --field <name>`, which prints exactly one
field as bare text. The accepted aliases are `private_key`, `private key`,
`public_key` and `public key`
(`pass-domain/src/models/item/field.rs:355-366`).

## Decision

We will read public keys with `item view --share-id <id> --item-id <id> --field
public_key`, and never call `item view --output json`.

More strongly: no code path in protui names the private key. `private_key` does
not appear as a string literal anywhere outside documentation explaining why it
is absent.

The same reasoning rules out `item list --show-secrets`, which swaps the
redacted item summary for full items including private keys. The default
summary form is safe by construction; upstream annotates it with
`// Fields here must never carry user-provided secret material`.

## Consequences

Private key material cannot enter protui's memory, because no call returns it.
This is a structural guarantee rather than a discipline — it does not depend on
remembering to avoid a field.

It costs one subprocess per key rather than one per item batch. The list call
returns no key material at all, so algorithm and fingerprint need a separate
fetch per item — see
[derive-key-metadata-from-public-keys](0006-derive-key-metadata-from-public-keys.md).

The field-scoped output is bare text, not JSON, so it is handled as a trimmed
string rather than decoded.

If protui ever needs a private key — exporting one, say — this decision must be
revisited explicitly rather than quietly relaxed, because the guarantee is
"no code path names it", and the first exception removes the guarantee for
everything.

Tests assert the constructed arguments: `--field public_key` is present,
`--output` is absent, and no argument contains "private".
