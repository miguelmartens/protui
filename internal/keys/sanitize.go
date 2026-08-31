package keys

import (
	"regexp"
	"strings"
	"unicode"
)

// Text that protui displays does not all come from protui. Item titles, vault
// names and key comments are stored in Proton Pass, and Proton Pass supports
// sharing items and vaults — so a title can have been written by somebody else
// and is untrusted input.
//
// Drawn unfiltered into a terminal, such a string is not merely data. A title
// containing ESC [ 2 J clears the screen; ESC ] 0 ; … BEL rewrites the window
// title; and OSC 52, which xterm, kitty, iTerm2, WezTerm and foot all honour,
// writes to the system clipboard. That last one matters here more than
// elsewhere: protui's headline action is putting a public key on the clipboard,
// and a malicious title could replace what the user believes they copied.
//
// Everything from outside is therefore sanitised at the boundary rather than at
// each render site, so nothing downstream has to remember.

// escapeSequence matches the terminal control sequences worth removing whole,
// so that stripping them leaves clean text rather than a visible tail like
// "[2J". Removing ESC alone would be enough for safety, since every sequence
// below needs it, but not enough to keep the display tidy.
var escapeSequence = regexp.MustCompile(
	// OSC: ESC ] … terminated by BEL or ST. This is the clipboard one.
	`\x1b\][^\x07\x1b]*(?:\x07|\x1b\\)?` +
		// CSI: ESC [ parameters intermediates final.
		`|\x1b\[[0-?]*[ -/]*[@-~]?` +
		// Two-character escapes, and a trailing lone ESC.
		`|\x1b[@-Z\\-_]?`,
)

// Sanitize returns s with terminal control sequences and display-reordering
// characters removed, leaving text that is safe to draw.
//
// It is deliberately lossy. Titles, vault names and comments are single-line
// display values, so tabs and newlines collapse to spaces rather than being
// preserved; anything relying on their layout is already being rendered wrong.
func Sanitize(s string) string {
	if s == "" {
		return s
	}

	s = escapeSequence.ReplaceAllString(s, "")

	return strings.Map(func(r rune) rune {
		switch {
		// Keep the word gap that whitespace represented, drop the control.
		case r == '\t', r == '\n', r == '\r', r == '\v', r == '\f':
			return ' '

		case unicode.IsControl(r):
			return -1

		// Explicit bidirectional overrides can reorder rendered text so that a
		// name displays as something other than what it is — the Trojan Source
		// class of attack. Only the explicit overrides are dropped; implicit
		// bidi, which real Arabic and Hebrew text relies on, still works.
		case r >= '\u202a' && r <= '\u202e', // LRE, RLE, PDF, LRO, RLO
			r >= '\u2066' && r <= '\u2069', // LRI, RLI, FSI, PDI
			r == '\u200e', r == '\u200f':   // LRM, RLM
			return -1

		// Line and paragraph separators are not control characters by category,
		// but they still break a single-line field.
		case r == '\u2028', r == '\u2029':
			return ' '
		}

		return r
	}, s)
}
