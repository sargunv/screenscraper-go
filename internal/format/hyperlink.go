package format

import (
	"github.com/charmbracelet/x/ansi"
)

// Hyperlink wraps text in OSC 8 escape sequences for clickable URLs in terminals.
// If the terminal doesn't support hyperlinks, it will just display the text.
func Hyperlink(url, text string) string {
	if url == "" {
		return text
	}
	return ansi.SetHyperlink(url) + text + ansi.SetHyperlink("")
}
