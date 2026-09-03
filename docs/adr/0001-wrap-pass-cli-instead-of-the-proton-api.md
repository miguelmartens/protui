# Wrap pass-cli instead of the Proton API

- **Status:** accepted
- **Date:** 2026-08-31

## Context

protui needs to read and write SSH key items in Proton Pass. Proton Pass is
end-to-end encrypted: items are encrypted client-side with keys derived from the
user's account, and the server never sees plaintext.

Three ways to reach that data:

1. **Reimplement the Proton Pass protocol.** Requires implementing SRP
   authentication, the OpenPGP key hierarchy, share and vault key derivation,
   and the protobuf item schema. Proton publishes no stable client API contract,
   so all of it would be reverse-engineered from the official clients.
2. **Bind to a Proton library.** No Go SDK exists. `pass-cli` is Rust and
   exposes no C ABI, so this would mean maintaining a cgo shim over an internal
   crate.
3. **Shell out to `pass-cli`.** The official CLI already holds the session,
   handles the crypto, and exposes `--output json` on the commands protui needs.

Option 1 means owning a cryptographic implementation on the path to a user's
private keys — the highest-consequence code in the project would be the part
nobody asked for. Option 2 inherits that surface with worse ergonomics.

The cost of option 3 is that protui inherits `pass-cli`'s output format, which
is not a documented stable contract (see
[record-the-upstream-schema-before-implementing](0002-record-the-upstream-schema-before-implementing.md)),
and one process spawn per operation.

## Decision

We will drive the official `pass-cli` binary as a subprocess, and implement no
Proton Pass protocol, cryptography, or authentication of our own.

`pass-cli` is a hard prerequisite. protui checks for the binary and for a valid
session at startup and refuses to open a UI it cannot use.

## Consequences

protui holds no long-term secret. It has no credential store, no session
handling, and no key derivation — all of that stays in the official client,
which is audited by people who own the format.

Users must install and authenticate `pass-cli` separately. This is the most
likely first-run failure, so it gets a dedicated startup check and an error
message naming `pass-cli login` rather than a generic failure.

protui can only do what `pass-cli` exposes. Anything absent from its command
surface is out of reach without upstream changes.

Every operation costs a process spawn, so latency is bounded below by
`pass-cli` startup. This is why deriving the metadata for a list of keys is
rate-limited rather than fanned out without limit — see
[derive-key-metadata-from-public-keys](0006-derive-key-metadata-from-public-keys.md).

We are exposed to upstream output changes, which is what
[isolate-pass-cli-behind-one-package](0003-isolate-pass-cli-behind-one-package.md)
and [tolerate-unknown-json-fields](0012-tolerate-unknown-json-fields.md) exist to
contain.
