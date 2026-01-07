package format

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// HeaderStyle is for section headers
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")) // Bright white

	// TitleStyle is for main titles
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("14")). // Cyan
			MarginBottom(1)

	// LabelStyle is for key-value labels
	LabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("12")). // Bright blue
			Bold(true)

	// ValueStyle is for key-value values
	ValueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")) // Bright white

	// DimStyle is for secondary information (IDs, dates, etc.)
	DimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")). // Gray
			Faint(true)

	// URLStyle is for URLs (before hyperlink wrapping)
	URLStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("14")). // Cyan
			Underline(true)

	// TableHeaderStyle is for table headers
	TableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("15")). // Bright white
				Align(lipgloss.Center)

	// TableCellStyle is for regular table cells
	TableCellStyle = lipgloss.NewStyle().
			Padding(0, 1)

	// TableOddRowStyle is for odd table rows
	TableOddRowStyle = TableCellStyle.
				Foreground(lipgloss.Color("7")) // Light gray

	// TableEvenRowStyle is for even table rows
	TableEvenRowStyle = TableCellStyle.
				Foreground(lipgloss.Color("15")) // Bright white

	// BorderStyle is for table borders
	BorderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")) // Gray
)
