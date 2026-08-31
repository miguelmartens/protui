// Package keys holds protui's domain types for SSH key items.
//
// Nothing here mirrors the pass-cli JSON shape: internal/passcli is responsible
// for translating upstream output into these types, so an upstream field rename
// stays contained to that package. See docs/schema.md.
//
// Upstream stores only the public key blob, the private key and a list of
// custom sections; there are no algorithm, fingerprint or comment fields. Those
// three are derived here by parsing the public key. Private key material never
// reaches this package.
package keys

import (
	"crypto/rsa"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

// Algorithm is the key algorithm, derived from the public key blob.
type Algorithm string

// Known algorithms. AlgorithmUnknown is used when the public key is missing or
// cannot be parsed, so a single malformed item never breaks the whole list.
const (
	AlgorithmED25519 Algorithm = "ed25519"
	AlgorithmRSA     Algorithm = "rsa"
	AlgorithmECDSA   Algorithm = "ecdsa"
	AlgorithmDSA     Algorithm = "dsa"
	AlgorithmSKECDSA Algorithm = "ecdsa-sk"
	AlgorithmSKED255 Algorithm = "ed25519-sk"
	AlgorithmUnknown Algorithm = "unknown"
)

// State mirrors an item's lifecycle state. Upstream spells these "Active" and
// "Trashed"; the mapping lives in internal/passcli.
type State string

// Item states.
const (
	StateActive  State = "active"
	StateTrashed State = "trashed"
)

// Vault is a Proton Pass vault. ShareID is the handle every pass-cli item
// subcommand expects; VaultID is carried for display only and is not accepted
// by any item command.
type Vault struct {
	Name    string
	ShareID string
	VaultID string
}

// Key is an SSH key item as protui presents it.
//
// PublicKey is the only key material ever held. Algorithm, Fingerprint, Comment
// and Bits are derived from it locally rather than read from upstream JSON.
type Key struct {
	ID        string
	ShareID   string
	VaultID   string
	VaultName string
	Title     string
	State     State

	PublicKey   string
	Algorithm   Algorithm
	Fingerprint string
	Comment     string
	Bits        int

	// CreatedAt and ModifiedAt are wall-clock times with no zone: upstream
	// serialises jiff::civil::DateTime, which carries no offset. They are
	// display and sort values only, never used for arithmetic across zones.
	CreatedAt  time.Time
	ModifiedAt time.Time
}

// PublicKeyInfo is what can be recovered from an OpenSSH public key line.
type PublicKeyInfo struct {
	Algorithm   Algorithm
	Fingerprint string
	Comment     string
	Bits        int
}

// Describe parses an OpenSSH authorized-keys line and derives the metadata that
// upstream does not store. A blank key yields unknown metadata and no error,
// since an item legitimately may not have one yet.
func Describe(publicKey string) (PublicKeyInfo, error) {
	// A valid key contains no control characters, so this is a no-op for real
	// input and neutralises a hostile blob before it is parsed or displayed.
	trimmed := strings.TrimSpace(Sanitize(publicKey))
	if trimmed == "" {
		return PublicKeyInfo{Algorithm: AlgorithmUnknown}, nil
	}

	pub, comment, _, _, err := ssh.ParseAuthorizedKey([]byte(trimmed))
	if err != nil {
		return PublicKeyInfo{Algorithm: AlgorithmUnknown}, fmt.Errorf("parse public key: %w", err)
	}

	return PublicKeyInfo{
		Algorithm:   algorithmFromKeyType(pub.Type()),
		Fingerprint: ssh.FingerprintSHA256(pub),
		// The comment is whatever bytes the key file carried after the blob.
		Comment: Sanitize(comment),
		Bits:    bitsOf(pub),
	}, nil
}

// WithPublicKey returns a copy of k carrying the public key and the metadata
// derived from it.
//
// A parse failure is returned for the caller to surface but is never fatal: the
// copy is still usable, with unknown metadata, so one malformed key cannot
// break a listing.
func (k Key) WithPublicKey(publicKey string) (Key, error) {
	k.PublicKey = strings.TrimSpace(Sanitize(publicKey))

	info, err := Describe(k.PublicKey)
	k.Algorithm = info.Algorithm
	k.Fingerprint = info.Fingerprint
	k.Comment = info.Comment
	k.Bits = info.Bits

	return k, err
}

// algorithmFromKeyType maps an SSH wire key type to an Algorithm. ECDSA and the
// security-key variants carry the curve in the type name, so they are matched
// by prefix.
func algorithmFromKeyType(keyType string) Algorithm {
	switch {
	case keyType == ssh.KeyAlgoED25519:
		return AlgorithmED25519
	case keyType == ssh.KeyAlgoSKED25519:
		return AlgorithmSKED255
	case keyType == ssh.KeyAlgoRSA, strings.HasPrefix(keyType, "rsa-sha2-"):
		return AlgorithmRSA
	// Matched by wire name rather than ssh.KeyAlgoDSA, which is deprecated:
	// DSA is only secure at key sizes nothing supports any more. pass-cli
	// cannot generate one, but an old imported key could still be listed, and
	// labelling it beats showing "unknown".
	case keyType == "ssh-dss":
		return AlgorithmDSA
	case strings.HasPrefix(keyType, "sk-ecdsa-sha2-"):
		return AlgorithmSKECDSA
	case strings.HasPrefix(keyType, "ecdsa-sha2-"):
		return AlgorithmECDSA
	default:
		return AlgorithmUnknown
	}
}

// bitsOf reports the key size where it is meaningful. ED25519 is fixed at 256
// bits, ECDSA curves are named in the key type, and RSA needs the modulus.
// It returns 0 when the size is unknown.
func bitsOf(pub ssh.PublicKey) int {
	keyType := pub.Type()

	switch {
	case keyType == ssh.KeyAlgoED25519, keyType == ssh.KeyAlgoSKED25519:
		return 256
	case strings.HasSuffix(keyType, "nistp256"):
		return 256
	case strings.HasSuffix(keyType, "nistp384"):
		return 384
	case strings.HasSuffix(keyType, "nistp521"):
		return 521
	}

	// CryptoPublicKey is the only route to the RSA modulus, and not every
	// implementation provides it.
	cryptoKey, ok := pub.(ssh.CryptoPublicKey)
	if !ok {
		return 0
	}

	if rsaKey, ok := cryptoKey.CryptoPublicKey().(*rsa.PublicKey); ok {
		return rsaKey.N.BitLen()
	}

	return 0
}

// Label renders the algorithm for display, including the size where it adds
// information. RSA is the only algorithm whose size is variable in practice.
func (k Key) Label() string {
	if k.Algorithm == AlgorithmRSA && k.Bits > 0 {
		return fmt.Sprintf("rsa%d", k.Bits)
	}
	return string(k.Algorithm)
}

// ShortFingerprint trims the "SHA256:" prefix for narrow columns.
func (k Key) ShortFingerprint() string {
	return strings.TrimPrefix(k.Fingerprint, "SHA256:")
}
