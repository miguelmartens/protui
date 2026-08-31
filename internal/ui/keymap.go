package ui

import "github.com/charmbracelet/bubbles/key"

// keyMap holds protui's bindings.
//
// Navigation is genuinely vim: j/k, gg/G, and Ctrl-f/b/d/u for paging, all
// with the same semantics and the same modifiers vim uses.
//
// Actions follow the lazygit and k9s idiom instead: one mnemonic letter each.
// They are not vim operators. In vim, d and y take a motion, which is why dd
// and yy exist — the doubling disambiguates "delete a line" from "delete a
// word". A list has one object under the cursor and no motions, so the second
// keystroke would resolve no ambiguity.
//
// Deletion is deliberately split across two keys. `d` trashes, which upstream
// can undo with `pass-cli item untrash`; `D` permanently destroys, which
// nothing can undo. Both confirm, but only the permanent one requires typing
// the title.
type keyMap struct {
	Up       key.Binding
	Down     key.Binding
	Top      key.Binding
	Bottom   key.Binding
	PageDown key.Binding
	PageUp   key.Binding
	HalfDown key.Binding
	HalfUp   key.Binding

	// ScrollDown and ScrollUp move the detail pane, not the list. vim's
	// Ctrl-e and Ctrl-y scroll the window without moving the cursor, which is
	// exactly the distinction wanted here.
	ScrollDown key.Binding
	ScrollUp   key.Binding
	Filter     key.Binding
	Copy       key.Binding
	New        key.Binding
	Trash      key.Binding
	Delete     key.Binding
	Agent      key.Binding
	Refresh    key.Binding
	Help       key.Binding
	Quit       key.Binding
}

func defaultKeyMap() keyMap {
	return keyMap{
		Up: key.NewBinding(
			key.WithKeys("k", "up"),
			key.WithHelp("k/↑", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("j", "down"),
			key.WithHelp("j/↓", "down"),
		),
		// Two-key sequence, as in vim. A bare g is a prefix there and never
		// acts on its own.
		Top: key.NewBinding(
			key.WithKeys("g"),
			key.WithHelp("gg", "top"),
		),
		Bottom: key.NewBinding(
			key.WithKeys("G", "end"),
			key.WithHelp("G", "bottom"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("ctrl+f", "pgdown"),
			key.WithHelp("^f", "page down"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("ctrl+b", "pgup"),
			key.WithHelp("^b", "page up"),
		),
		HalfDown: key.NewBinding(
			key.WithKeys("ctrl+d"),
			key.WithHelp("^d", "half page down"),
		),
		HalfUp: key.NewBinding(
			key.WithKeys("ctrl+u"),
			key.WithHelp("^u", "half page up"),
		),
		ScrollDown: key.NewBinding(
			key.WithKeys("ctrl+e"),
			key.WithHelp("^e", "scroll detail"),
		),
		ScrollUp: key.NewBinding(
			key.WithKeys("ctrl+y"),
			key.WithHelp("^y", "scroll detail up"),
		),
		Filter: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "filter"),
		),
		// c is the advertised key; y stays bound as an alias, since the
		// navigation around it is vim's and yank is the reflex that comes with
		// that.
		Copy: key.NewBinding(
			key.WithKeys("c", "y"),
			key.WithHelp("c", "copy pubkey"),
		),
		// n, as in lazygit. Binding n to list movement instead would only
		// duplicate j, because `/` here filters rather than searches: there is
		// no hidden next match to jump to.
		New: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "new"),
		),
		Trash: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "trash"),
		),
		Delete: key.NewBinding(
			key.WithKeys("D"),
			key.WithHelp("D", "delete!"),
		),
		Agent: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "agent"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
	}
}

// ShortHelp is the single-line help bar shown under the list.
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Filter, k.Copy, k.New, k.Trash, k.Agent, k.Help, k.Quit}
}

// FullHelp is the expanded help, toggled with `?`.
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Top, k.Bottom},
		{k.PageDown, k.PageUp},
		{k.HalfDown, k.HalfUp, k.ScrollDown, k.ScrollUp},
		{k.Filter, k.Refresh},
		{k.Copy, k.New, k.Trash, k.Delete},
		{k.Agent, k.Help, k.Quit},
	}
}
