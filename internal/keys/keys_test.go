package keys

import (
	"strings"
	"testing"
)

// Throwaway keys generated with ssh-keygen purely as test vectors. They guard
// nothing and correspond to no private key that exists anywhere.
const (
	ed25519Key = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIM+XgkeZS/lofu0u1xq0g4DQUZIOxGcdSqHhQbjKwVEQ laptop@example"
	ed25519FP  = "SHA256:jgNTr/lOf7SWNNgIkhV2xOfVZI2zdWn7WremFIOCmaI"

	rsa2048Key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDDMU3w/a9F5JytBm0CPnKCZ9XlyVxG7d3afM0dJg1+4yZp2dWifkADPAxWAxaLnY7gcJzyCDIC3j5klN1deFpspph4AX09detMoaunhNa/xpxFmWRHin36F+6EQWAKtshvBCoenF8PCcFYrnGsatPPxI2Dbyl0QNacQgfiG4C0YLS1/Ajx/JyY8MwoVqM6hIBmE/hVdAzH+EwVe7R7jLcLNAW63j71At/OYyWOCMeOkPN3JPzRHxadMl666XI/ML0tzmv2akTi3bohkQUtXtFS7rFmVhZF78kYACs4gWoyxNCPpHFh3omYLvGL5xk0qhCgs0O99d6vmNZ2BDb1MNgf"
	rsa2048FP  = "SHA256:QLsAuTpG7EF7VFz0rGC7Jy34NWCKPWprfrJBdoABz/k"

	ecdsaKey = "ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBMtPkKX8SOszgwcAyzKEYDjyg5xeCmZGv9aQa11v7RY8zn043Sb6mB6B15l+kUfNtOkffJ8tUtaqoj4bP1Hn6pk= ci@example"
	ecdsaFP  = "SHA256:6F9nCBAirUvYGJodR4JYpiVk6ciHhU6Z+tBUrNi3U98"
)

// Upstream stores none of algorithm, fingerprint or comment, so every field
// asserted here is derived locally from the public key. See docs/schema.md §1.1.
func TestDescribe(t *testing.T) {
	tests := []struct {
		name            string
		publicKey       string
		wantAlgorithm   Algorithm
		wantFingerprint string
		wantComment     string
		wantBits        int
		wantErr         bool
	}{
		{
			name:            "ed25519 with comment",
			publicKey:       ed25519Key,
			wantAlgorithm:   AlgorithmED25519,
			wantFingerprint: ed25519FP,
			wantComment:     "laptop@example",
			wantBits:        256,
		},
		{
			// ssh-keygen writes no trailing field when -C is empty.
			name:            "rsa2048 without comment",
			publicKey:       rsa2048Key,
			wantAlgorithm:   AlgorithmRSA,
			wantFingerprint: rsa2048FP,
			wantComment:     "",
			wantBits:        2048,
		},
		{
			name:            "ecdsa nistp256",
			publicKey:       ecdsaKey,
			wantAlgorithm:   AlgorithmECDSA,
			wantFingerprint: ecdsaFP,
			wantComment:     "ci@example",
			wantBits:        256,
		},
		{
			name:            "surrounding whitespace is tolerated",
			publicKey:       "  " + ed25519Key + "\n",
			wantAlgorithm:   AlgorithmED25519,
			wantFingerprint: ed25519FP,
			wantComment:     "laptop@example",
			wantBits:        256,
		},
		{
			// An item may legitimately have no public key yet; that is not an
			// error, it just yields unknown metadata.
			name:          "empty key is not an error",
			publicKey:     "",
			wantAlgorithm: AlgorithmUnknown,
		},
		{
			name:          "whitespace only is not an error",
			publicKey:     "   \n\t ",
			wantAlgorithm: AlgorithmUnknown,
		},
		{
			name:          "garbage reports unknown and errors",
			publicKey:     "this is not a key",
			wantAlgorithm: AlgorithmUnknown,
			wantErr:       true,
		},
		{
			name:          "truncated base64 reports unknown and errors",
			publicKey:     "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAA",
			wantAlgorithm: AlgorithmUnknown,
			wantErr:       true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := Describe(test.publicKey)

			if test.wantErr && err == nil {
				t.Fatal("expected an error")
			}
			if !test.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got.Algorithm != test.wantAlgorithm {
				t.Errorf("algorithm = %q, want %q", got.Algorithm, test.wantAlgorithm)
			}
			if got.Fingerprint != test.wantFingerprint {
				t.Errorf("fingerprint = %q, want %q", got.Fingerprint, test.wantFingerprint)
			}
			if got.Comment != test.wantComment {
				t.Errorf("comment = %q, want %q", got.Comment, test.wantComment)
			}
			if got.Bits != test.wantBits {
				t.Errorf("bits = %d, want %d", got.Bits, test.wantBits)
			}
		})
	}
}

// TestDescribeFingerprintMatchesSSHKeygen pins the fingerprints against the
// values `ssh-keygen -lf` produced for the same keys, so a change in the
// hashing or encoding would be caught.
func TestDescribeFingerprintMatchesSSHKeygen(t *testing.T) {
	tests := map[string]string{
		ed25519Key: ed25519FP,
		rsa2048Key: rsa2048FP,
		ecdsaKey:   ecdsaFP,
	}

	for publicKey, want := range tests {
		got, err := Describe(publicKey)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Fingerprint != want {
			t.Errorf("fingerprint = %q, want %q", got.Fingerprint, want)
		}
		if !strings.HasPrefix(got.Fingerprint, "SHA256:") {
			t.Errorf("fingerprint %q is not SHA256-prefixed", got.Fingerprint)
		}
	}
}

func TestKeyWithPublicKey(t *testing.T) {
	t.Run("populates derived metadata", func(t *testing.T) {
		got, err := Key{ID: "id", Title: "laptop"}.WithPublicKey(ed25519Key)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got.Algorithm != AlgorithmED25519 {
			t.Errorf("algorithm = %q, want %q", got.Algorithm, AlgorithmED25519)
		}
		if got.Fingerprint != ed25519FP {
			t.Errorf("fingerprint = %q, want %q", got.Fingerprint, ed25519FP)
		}
		if got.Comment != "laptop@example" {
			t.Errorf("comment = %q, want %q", got.Comment, "laptop@example")
		}
		if got.PublicKey != ed25519Key {
			t.Error("public key was not stored verbatim")
		}
	})

	t.Run("a bad key still leaves a usable row", func(t *testing.T) {
		got, err := Key{ID: "id", Title: "broken"}.WithPublicKey("nonsense")
		if err == nil {
			t.Fatal("expected an error")
		}

		// The error is reported, but the item must remain listable.
		if got.Title != "broken" {
			t.Error("title was clobbered")
		}
		if got.Algorithm != AlgorithmUnknown {
			t.Errorf("algorithm = %q, want %q", got.Algorithm, AlgorithmUnknown)
		}
	})

	t.Run("the receiver is not mutated", func(t *testing.T) {
		original := Key{ID: "id", Title: "laptop"}

		if _, err := original.WithPublicKey(ed25519Key); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if original.PublicKey != "" || original.Fingerprint != "" {
			t.Errorf("WithPublicKey mutated its receiver: %+v", original)
		}
	})
}

func TestKeyLabel(t *testing.T) {
	tests := []struct {
		name string
		key  Key
		want string
	}{
		{
			// RSA is the only algorithm whose size varies in practice, so it
			// is the only one that carries the bit count.
			name: "rsa carries its size",
			key:  Key{Algorithm: AlgorithmRSA, Bits: 4096},
			want: "rsa4096",
		},
		{
			name: "rsa without bits falls back to the bare name",
			key:  Key{Algorithm: AlgorithmRSA},
			want: "rsa",
		},
		{
			name: "ed25519 is fixed size so omits it",
			key:  Key{Algorithm: AlgorithmED25519, Bits: 256},
			want: "ed25519",
		},
		{
			name: "unknown",
			key:  Key{Algorithm: AlgorithmUnknown},
			want: "unknown",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.key.Label(); got != test.want {
				t.Errorf("Label() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestShortFingerprint(t *testing.T) {
	key := Key{Fingerprint: ed25519FP}

	want := strings.TrimPrefix(ed25519FP, "SHA256:")
	if got := key.ShortFingerprint(); got != want {
		t.Errorf("ShortFingerprint() = %q, want %q", got, want)
	}

	// A missing fingerprint must not gain a stray prefix.
	if got := (Key{}).ShortFingerprint(); got != "" {
		t.Errorf("ShortFingerprint() on a zero Key = %q, want empty", got)
	}
}
