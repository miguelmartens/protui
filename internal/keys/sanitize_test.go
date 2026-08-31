package keys

import (
	"strings"
	"testing"
	"unicode"
)

func TestSanitize(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "ordinary text is untouched",
			in:   "laptop key (work)",
			want: "laptop key (work)",
		},
		{
			name: "non-ASCII text is untouched",
			in:   "clé — ключ — 鍵 — مفتاح",
			want: "clé — ключ — 鍵 — مفتاح",
		},
		{
			name: "empty stays empty",
			in:   "",
			want: "",
		},
		{
			name: "CSI clear-screen is removed whole",
			in:   "before\x1b[2Jafter",
			want: "beforeafter",
		},
		{
			name: "SGR colour codes are removed whole",
			in:   "\x1b[31mred\x1b[0m",
			want: "red",
		},
		{
			// The attack that matters most here: OSC 52 writes the system
			// clipboard on xterm, kitty, iTerm2, WezTerm and foot.
			name: "OSC 52 clipboard write is removed",
			in:   "key\x1b]52;c;aGVsbG8=\x07name",
			want: "keyname",
		},
		{
			name: "OSC window-title set is removed",
			in:   "a\x1b]0;pwned\x07b",
			want: "ab",
		},
		{
			name: "OSC terminated by ST is removed",
			in:   "a\x1b]0;x\x1b\\b",
			want: "ab",
		},
		{
			name: "a bare escape is removed",
			in:   "a\x1bb",
			want: "ab",
		},
		{
			name: "BEL and NUL are removed",
			in:   "a\x07b\x00c",
			want: "abc",
		},
		{
			name: "DEL is removed",
			in:   "a\x7fb",
			want: "ab",
		},
		{
			name: "newlines and tabs become spaces",
			in:   "one\ttwo\nthree\r\nfour",
			want: "one two three  four",
		},
		{
			// Trojan Source: an override can make a name render as something
			// other than what it is.
			name: "bidi overrides are removed",
			in:   "safe\u202ereversed\u202c",
			want: "safereversed",
		},
		{
			name: "bidi isolates are removed",
			in:   "a\u2066b\u2069c",
			want: "abc",
		},
		{
			name: "line and paragraph separators become spaces",
			in:   "a\u2028b\u2029c",
			want: "a b c",
		},
		{
			name: "legitimate right-to-left text still works",
			// No explicit override characters, so nothing is dropped: real
			// Arabic and Hebrew rely on implicit bidi.
			in:   "مفتاح ssh",
			want: "مفتاح ssh",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := Sanitize(test.in); got != test.want {
				t.Errorf("Sanitize(%q) = %q, want %q", test.in, got, test.want)
			}
		})
	}
}

// TestSanitizeLeavesNoControlRunes is the property that matters, checked
// independently of the table above so a sequence nobody thought to enumerate
// still cannot get through.
func TestSanitizeLeavesNoControlRunes(t *testing.T) {
	var builder strings.Builder
	for r := rune(0); r < 0x100; r++ {
		builder.WriteRune(r)
	}
	builder.WriteString("\u202e\u2066\u200e\u2028")

	for _, r := range Sanitize(builder.String()) {
		if unicode.IsControl(r) {
			t.Errorf("control rune %U survived sanitising", r)
		}
	}
}

// TestSanitizeIsIdempotent guards against a rewrite that only strips one layer,
// which would let a doubled sequence through.
func TestSanitizeIsIdempotent(t *testing.T) {
	for _, in := range []string{
		"\x1b[2J", "\x1b\x1b[2J[2J", "a\x1b]52;c;x\x07b", "plain",
	} {
		once := Sanitize(in)
		if twice := Sanitize(once); once != twice {
			t.Errorf("Sanitize(%q) = %q but sanitising again gave %q", in, once, twice)
		}
	}
}

// TestDescribeSanitizesTheComment covers the comment, which is whatever bytes
// followed the key blob in the file it came from.
func TestDescribeSanitizesTheComment(t *testing.T) {
	const body = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIM+XgkeZS/lofu0u1xq0g4DQUZIOxGcdSqHhQbjKwVEQ"

	info, err := Describe(body + " evil\x1b[2Jcomment")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.Contains(info.Comment, "\x1b") {
		t.Errorf("comment carries an escape: %q", info.Comment)
	}
	if info.Algorithm != AlgorithmED25519 {
		t.Errorf("algorithm = %q, want %q", info.Algorithm, AlgorithmED25519)
	}
}
