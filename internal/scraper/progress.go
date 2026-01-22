package scraper

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Model is the bubbletea model for the progress display
type Model struct {
	// State
	total    int
	quitting bool

	// Active lookups (in progress) - map for fast lookup, slice for stable order
	activeLookups map[string]*activeEntry
	activeOrder   []string

	// Counts
	processed       int
	found           int
	notFound        int
	skipped         int
	errors          int
	mediaDownloaded int
	cacheHits       int

	// Thread info
	maxThreads int

	// Timing
	startTime time.Time

	// Components
	spinner  spinner.Model
	progress progress.Model

	// Updates channel
	updatesCh <-chan ProgressUpdate

	// Stats provider for rate/ETA calculation
	getStats        func() RateLimiterStats
	mediaTypesCount int
}

// activeEntry tracks an in-progress lookup
type activeEntry struct {
	name         string
	mediaTotal   int
	mediaDone    int
	mediaFailed  int
	mediaMissing int
	currentMedia string
	startTime    time.Time
}

// Styles
var (
	spinnerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205"))

	foundStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("10"))

	notFoundStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("9"))

	skippedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("11"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")).
			Bold(true)

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	dotDoneStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("10"))

	dotFailedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("9"))

	dotMissingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	dotPendingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))
)

// NewModel creates a new progress model
func NewModel(total, maxThreads, mediaTypesCount int, updatesCh <-chan ProgressUpdate, getStats func() RateLimiterStats) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = spinnerStyle

	p := progress.New(progress.WithDefaultGradient())

	return Model{
		total:           total,
		maxThreads:      maxThreads,
		mediaTypesCount: mediaTypesCount,
		startTime:       time.Now(),
		spinner:         s,
		progress:        p,
		activeLookups:   make(map[string]*activeEntry),
		updatesCh:       updatesCh,
		getStats:        getStats,
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		waitForUpdate(m.updatesCh),
	)
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			// Print all in-progress entries as cancelled before quitting
			var cmds []tea.Cmd
			for _, name := range m.activeOrder {
				entry := m.activeLookups[name]
				if entry == nil {
					continue
				}
				duration := formatDuration(entry)
				line := fmt.Sprintf(" %s  %-42s %s  %s",
					dimStyle.Render("⊘"),
					truncate(name, 42),
					renderDots(entry.mediaDone, entry.mediaFailed, entry.mediaMissing, entry.mediaTotal),
					dimStyle.Render("cancelled  "+duration))
				cmds = append(cmds, tea.Println(line))
			}
			cmds = append(cmds, tea.Quit)
			return m, tea.Sequence(cmds...)
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		return m, cmd

	case ProgressUpdate:
		var printCmd tea.Cmd
		m, printCmd = m.handleUpdate(msg)

		// Check if done
		if m.processed >= m.total {
			return m, tea.Sequence(printCmd, tea.Quit)
		}

		return m, tea.Batch(printCmd, waitForUpdate(m.updatesCh))

	case doneMsg:
		return m, tea.Quit
	}

	return m, nil
}

func (m Model) handleUpdate(update ProgressUpdate) (Model, tea.Cmd) {
	switch update.Type {
	case UpdateTypeStarted:
		// Add to active lookups
		m.activeLookups[update.EntryName] = &activeEntry{
			name:       update.EntryName,
			mediaTotal: update.MediaTotal,
			mediaDone:  0,
			startTime:  time.Now(),
		}
		m.activeOrder = append(m.activeOrder, update.EntryName)

	case UpdateTypeProgress:
		// Update media progress
		if entry, ok := m.activeLookups[update.EntryName]; ok {
			entry.mediaDone = update.MediaDone
			entry.mediaFailed = update.MediaFailed
			entry.mediaMissing = update.MediaMissing
			entry.currentMedia = update.CurrentMedia
		}

	case UpdateTypeFound:
		m.processed++
		m.found++
		m.mediaDownloaded += update.MediaDone
		m.cacheHits += update.CacheHits
		duration := formatDuration(m.activeLookups[update.EntryName])
		delete(m.activeLookups, update.EntryName)
		m.activeOrder = removeFromOrder(m.activeOrder, update.EntryName)

		// Print completed entry to scrollback
		// Total possible API calls = 1 (game info) + media count
		totalCalls := 1 + update.MediaTotal
		apiCalls := totalCalls - update.CacheHits
		cacheInfo := ""
		if apiCalls == 0 {
			cacheInfo = " (all cached)"
		} else if update.CacheHits > 0 {
			cacheInfo = fmt.Sprintf(" (%d/%d cached)", update.CacheHits, totalCalls)
		}

		line := fmt.Sprintf(" %s  %-42s %s  %s%s",
			foundStyle.Render("✓"),
			truncate(update.EntryName, 42),
			renderDots(update.MediaDone, update.MediaFailed, update.MediaMissing, update.MediaTotal),
			dimStyle.Render(fmt.Sprintf("%d media", update.MediaDone)),
			dimStyle.Render(cacheInfo))
		line += dimStyle.Render("  " + duration)
		return m, tea.Println(line)

	case UpdateTypeNotFound:
		m.processed++
		m.notFound++
		duration := formatDuration(m.activeLookups[update.EntryName])
		delete(m.activeLookups, update.EntryName)
		m.activeOrder = removeFromOrder(m.activeOrder, update.EntryName)

		line := fmt.Sprintf(" %s  %-42s %s  %s",
			notFoundStyle.Render("✗"),
			truncate(update.EntryName, 42),
			dimStyle.Render("not found"),
			dimStyle.Render(duration))
		return m, tea.Println(line)

	case UpdateTypeSkipped:
		m.processed++
		m.skipped++
		duration := formatDuration(m.activeLookups[update.EntryName])
		delete(m.activeLookups, update.EntryName)
		m.activeOrder = removeFromOrder(m.activeOrder, update.EntryName)

		line := fmt.Sprintf(" %s  %-42s %s  %s",
			skippedStyle.Render("⊘"),
			truncate(update.EntryName, 42),
			dimStyle.Render("skipped"),
			dimStyle.Render(duration))
		return m, tea.Println(line)

	case UpdateTypeError:
		m.processed++
		m.errors++
		duration := formatDuration(m.activeLookups[update.EntryName])
		delete(m.activeLookups, update.EntryName)
		m.activeOrder = removeFromOrder(m.activeOrder, update.EntryName)

		errMsg := "error"
		if update.Error != nil {
			errMsg = truncate(update.Error.Error(), 25)
		}
		line := fmt.Sprintf(" %s  %-42s %s  %s",
			errorStyle.Render("!"),
			truncate(update.EntryName, 42),
			errorStyle.Render(errMsg),
			dimStyle.Render(duration))
		return m, tea.Println(line)
	}

	return m, nil
}

// View renders the live UI (in-progress items + stats)
func (m Model) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder

	// In-progress lookups with media dots (in stable order)
	for _, name := range m.activeOrder {
		entry := m.activeLookups[name]
		if entry == nil {
			continue
		}
		dots := renderDots(entry.mediaDone, entry.mediaFailed, entry.mediaMissing, entry.mediaTotal)
		elapsed := fmtDuration(time.Since(entry.startTime))
		var status string
		if entry.currentMedia != "" {
			status = fmt.Sprintf("  fetching %s  %s", entry.currentMedia, elapsed)
		} else {
			status = fmt.Sprintf("  identifying  %s", elapsed)
		}
		b.WriteString(fmt.Sprintf(" %s %-42s %s%s\n",
			m.spinner.View(),
			truncate(entry.name, 42),
			dots,
			dimStyle.Render(status)))
	}

	// Separator
	b.WriteString(strings.Repeat("━", 60) + "\n")

	// Progress bar
	pct := float64(m.processed) / float64(m.total)
	b.WriteString(" Progress  ")
	b.WriteString(m.progress.ViewAs(pct))
	b.WriteString(fmt.Sprintf("  %d/%d (%.0f%%)\n\n", m.processed, m.total, pct*100))

	// Stats
	b.WriteString(fmt.Sprintf(" Found: %s    Not Found: %s    Skipped: %s    Errors: %s\n",
		foundStyle.Render(fmt.Sprintf("%d", m.found)),
		notFoundStyle.Render(fmt.Sprintf("%d", m.notFound)),
		skippedStyle.Render(fmt.Sprintf("%d", m.skipped)),
		errorStyle.Render(fmt.Sprintf("%d", m.errors)),
	))
	b.WriteString(fmt.Sprintf(" Media: %d downloaded    Cache hits: %d\n\n",
		m.mediaDownloaded,
		m.cacheHits,
	))

	// Timing
	elapsed := time.Since(m.startTime).Round(time.Second)

	// Get API stats for rate calculation
	var apiRate float64
	var totalAPICalls int
	if m.getStats != nil {
		stats := m.getStats()
		apiRate = stats.RequestsPerSecond
		totalAPICalls = stats.TotalRequests
	}

	// Calculate worst-case ETA (assumes all remaining need full API calls)
	remaining := m.total - m.processed
	callsPerGame := 1 + m.mediaTypesCount // 1 game info + N media calls
	worstCaseRemaining := remaining * callsPerGame

	var eta string
	if apiRate > 0.1 {
		etaDur := time.Duration(float64(worstCaseRemaining)/apiRate) * time.Second
		eta = "~" + etaDur.Round(time.Second).String()
	} else {
		eta = "calculating..."
	}
	b.WriteString(fmt.Sprintf(" Elapsed: %s    API: %d calls (~%.1f/s)    ETA: %s\n", elapsed, totalAPICalls, apiRate, eta))

	return b.String()
}

// Message types
type doneMsg struct{}

// waitForUpdate creates a command that waits for the next update
func waitForUpdate(ch <-chan ProgressUpdate) tea.Cmd {
	return func() tea.Msg {
		update, ok := <-ch
		if !ok {
			return doneMsg{}
		}
		return update
	}
}

func renderDots(done, failed, missing, total int) string {
	var b strings.Builder
	processed := 0

	// Done (green filled)
	for i := 0; i < done && processed < total; i++ {
		b.WriteString(dotDoneStyle.Render("●"))
		processed++
	}

	// Failed (red X)
	for i := 0; i < failed && processed < total; i++ {
		b.WriteString(dotFailedStyle.Render("⊗"))
		processed++
	}

	// Missing (gray dash)
	for i := 0; i < missing && processed < total; i++ {
		b.WriteString(dotMissingStyle.Render("○"))
		processed++
	}

	// Pending (gray empty)
	for processed < total {
		b.WriteString(dotPendingStyle.Render("◌"))
		processed++
	}

	return b.String()
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

func removeFromOrder(order []string, name string) []string {
	for i, n := range order {
		if n == name {
			return append(order[:i], order[i+1:]...)
		}
	}
	return order
}

func formatDuration(entry *activeEntry) string {
	if entry == nil {
		return ""
	}
	return fmtDuration(time.Since(entry.startTime))
}

func fmtDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	// Always show one decimal place for seconds to avoid flicker
	secs := d.Seconds()
	return fmt.Sprintf("%.1fs", secs)
}
