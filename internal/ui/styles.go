package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// The palette is one accent against a neutral ramp. The accent is Proton's
// violet, which is both on-brand for a Proton Pass client and distinct from the
// green/blue most terminal tools default to.
//
// Every colour is adaptive so the UI stays legible on light and dark terminals,
// and each is a hex pair rather than an ANSI index so the two themes can be
// tuned independently — the dark variants are lifted and desaturated, since a
// colour that reads well on white is muddy on black.
var (
	colorAccent    = lipgloss.AdaptiveColor{Light: "#5B34E0", Dark: "#9C82FF"}
	colorAccentDim = lipgloss.AdaptiveColor{Light: "#8B6FEF", Dark: "#6A55B8"}

	// The neutral ramp, brightest to dimmest.
	colorText  = lipgloss.AdaptiveColor{Light: "#18181B", Dark: "#E8E8ED"}
	colorMuted = lipgloss.AdaptiveColor{Light: "#5F6570", Dark: "#9A9AAB"}
	colorFaint = lipgloss.AdaptiveColor{Light: "#9CA3AF", Dark: "#61616F"}
	colorLine  = lipgloss.AdaptiveColor{Light: "#E3E3E8", Dark: "#2C2C36"}

	// colorSurface backs badges and inline chips.
	colorSurface = lipgloss.AdaptiveColor{Light: "#F1EEFD", Dark: "#2A2438"}

	colorOK    = lipgloss.AdaptiveColor{Light: "#15803D", Dark: "#5DD68A"}
	colorWarn  = lipgloss.AdaptiveColor{Light: "#B45309", Dark: "#F0B429"}
	colorError = lipgloss.AdaptiveColor{Light: "#B91C1C", Dark: "#FF7B72"}
)

// Text styles, from loudest to quietest.
var (
	styleAccent = lipgloss.NewStyle().Foreground(colorAccent)
	styleText   = lipgloss.NewStyle().Foreground(colorText)
	styleMuted  = lipgloss.NewStyle().Foreground(colorMuted)
	styleFaint  = lipgloss.NewStyle().Foreground(colorFaint)

	styleOK    = lipgloss.NewStyle().Foreground(colorOK)
	styleWarn  = lipgloss.NewStyle().Foreground(colorWarn)
	styleError = lipgloss.NewStyle().Foreground(colorError)

	// styleWordmark is the product name in the header. There is no filled
	// banner: a coloured block across the top of a terminal reads as chrome,
	// and the list is what deserves attention.
	styleWordmark = lipgloss.NewStyle().Foreground(colorAccent).Bold(true)

	// styleSection labels a group inside the detail pane. Uppercase and faint
	// rather than bold and bright, so it separates without competing.
	styleSection = lipgloss.NewStyle().Foreground(colorFaint)

	// styleLabel is the left column of a key/value row.
	styleLabel = lipgloss.NewStyle().Foreground(colorMuted)

	// styleSelected marks the highlighted row.
	styleSelected = lipgloss.NewStyle().Foreground(colorAccent).Bold(true)

	// styleBadge is an inline chip, used for the key algorithm.
	styleBadge = lipgloss.NewStyle().
			Foreground(colorAccent).
			Background(colorSurface).
			Padding(0, 1)

	styleRule = lipgloss.NewStyle().Foreground(colorLine)
)

// Containers.
var (
	// styleDetail is the right-hand pane, separated by a rule rather than
	// boxed: a full border around a pane that already sits at the screen edge
	// is chrome for its own sake.
	styleDetail = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(colorLine).
			PaddingLeft(2)

	// styleDialog frames a modal. Rounded corners read as softer than square
	// ones, which suits a prompt rather than a warning.
	styleDialog = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorAccentDim).
			Padding(1, 3)

	// styleDanger frames the permanent-delete prompt. Red border and square
	// corners, so the irreversible action does not look like the others.
	styleDanger = lipgloss.NewStyle().
			Border(lipgloss.ThickBorder()).
			BorderForeground(colorError).
			Padding(1, 3)
)

// Glyphs. Kept to widely-supported code points: box drawing, geometric shapes
// and arrows all render in the default fonts of common terminals, whereas
// emoji and Nerd Font glyphs do not.
const (
	glyphMark      = "◆"
	glyphSelected  = "▌"
	glyphSeparator = "·"
	glyphRunning   = "●"
	glyphStopped   = "○"
	glyphDegraded  = "◐"
	glyphOK        = "✓"
	glyphWarn      = "!"
	glyphError     = "✗"
	glyphRule      = "─"
)

// rule draws a horizontal divider.
func rule(width int) string {
	if width < 1 {
		return ""
	}

	return styleRule.Render(strings.Repeat(glyphRule, width))
}

// helpStyles tints the help bar so the keys read as interactive and their
// descriptions recede.
func helpStyles() (key, desc, separator lipgloss.Style) {
	return lipgloss.NewStyle().Foreground(colorAccent),
		lipgloss.NewStyle().Foreground(colorFaint),
		lipgloss.NewStyle().Foreground(colorLine)
}
