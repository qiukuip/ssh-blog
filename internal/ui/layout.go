package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	siteName = "ssh-blog"
	pageHome = iota
	pageArchive
	pageFriends
	pageMessages
	pageAbout
)

type NavItem struct {
	Key   string
	Label string
	Page  int
}

func NavItems() []NavItem {
	return []NavItem{
		{Key: "1", Label: "首页", Page: pageHome},
		{Key: "2", Label: "归档", Page: pageArchive},
		{Key: "3", Label: "友链", Page: pageFriends},
		{Key: "4", Label: "留言", Page: pageMessages},
		{Key: "5", Label: "关于", Page: pageAbout},
	}
}

func pageLabel(p int) string {
	for _, n := range NavItems() {
		if n.Page == p {
			return n.Label
		}
	}
	return ""
}

type FooterConfig struct {
	Author  string
	Copy    string
	Beian   string
	Kaiwang string // "开往" link
	Shinian string // "十年之约" link
}

var defaultFooter = FooterConfig{
	Author:  "longkun",
	Copy:    "© 2018-2026 ssh-blog",
	Beian:   "萌ICP备 0000000 号",
	Kaiwang: "https://开往.cn",
	Shinian: "https://www.foreverblog.cn",
}

// PageWidth returns the inner content width given a terminal width,
// clamping to [MinContentWidth, MaxContentWidth]. The reading view and the
// rest of the body share the same ContentMeasure so layouts line up.
const (
	ContentMeasure = 110
	MinContentWidth = ContentMeasure
	MaxContentWidth = ContentMeasure
	OuterPadding    = 4
)

func PageWidth(terminalW int) int {
	w := terminalW - OuterPadding*2
	if w < MinContentWidth {
		w = terminalW
	}
	if w > MaxContentWidth {
		w = MaxContentWidth
	}
	if w < 20 {
		w = 20
	}
	return w
}

// ContentMeasureWidth returns the standard content column width for the given
// terminal width. Caps at ContentMeasure on wide terminals; otherwise leaves
// room for a 2-cell gutter on each side so the content stays centered.
func ContentMeasureWidth(terminalW int) int {
	if terminalW >= ContentMeasure+4 {
		return ContentMeasure
	}
	return terminalW - 4
}

// ContentMeasurePad returns the horizontal padding needed to center a
// ContentMeasure-wide block in a terminal of the given width.
func ContentMeasurePad(terminalW int) int {
	w := ContentMeasureWidth(terminalW)
	if terminalW <= w {
		return 0
	}
	return (terminalW - w) / 2
}

// ContentBox returns the content rendered centered within the terminal width.

func ContentBox(content string, terminalW int) string {
	inner := PageWidth(terminalW)
	if terminalW < MinContentWidth+OuterPadding*2 {
		return content
	}
	pad := (terminalW - inner) / 2
	if pad < 0 {
		pad = 0
	}
	// Wrap with terminal width so any line that exceeds it gets visually wrapped,
	// but the height budget stays predictable.
	return lipgloss.NewStyle().
		Width(terminalW).
		Padding(0, pad, 0, pad).
		Render(content)
}

func TopBar(currentPage int, theme Theme, themeIdx, totalThemes int, width int) string {
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Primary).
		Render(siteName)

	var navParts []string
	for _, n := range NavItems() {
		label := fmt.Sprintf("%s %s", n.Key, n.Label)
		if n.Page == currentPage {
			navParts = append(navParts, lipgloss.NewStyle().
				Bold(true).
				Foreground(theme.NavActive).
				Background(theme.Primary).
				Padding(0, 1).
				Render(label))
		} else {
			navParts = append(navParts, lipgloss.NewStyle().
				Foreground(theme.NavInactive).
				Padding(0, 1).
				Render(label))
		}
	}
	nav := lipgloss.NewStyle().Render(strings.Join(navParts, " "))

	// Compact theme hint for narrow terminals.
	themeHint := lipgloss.NewStyle().
		Foreground(theme.Muted).
		Italic(true).
		Render(fmt.Sprintf("[T] %s %d/%d", theme.Name, themeIdx+1, totalThemes))

	left := title
	right := themeHint
	boxW := width - 2
	if boxW < 20 {
		boxW = 20
	}
	contentW := boxW - 2 // account for padding

	used := lipgloss.Width(left) + lipgloss.Width(nav) + lipgloss.Width(right) + 1
	gap := contentW - used
	if gap < 1 {
		// Try to fit by trimming nav key prefix when very narrow.
		if width < 80 {
			navParts = navParts[:0]
			for _, n := range NavItems() {
				if n.Page == currentPage {
					navParts = append(navParts, lipgloss.NewStyle().
						Bold(true).
						Foreground(theme.NavActive).
						Background(theme.Primary).
						Render(n.Label))
				} else {
					navParts = append(navParts, lipgloss.NewStyle().
						Foreground(theme.NavInactive).
						Render(n.Label))
				}
			}
			nav = lipgloss.NewStyle().Render(strings.Join(navParts, "·"))
		}
		used = lipgloss.Width(left) + lipgloss.Width(nav) + lipgloss.Width(right) + 1
		gap = contentW - used
		if gap < 1 {
			gap = 1
		}
	}

	bar := left + strings.Repeat(" ", gap) + nav + " " + right

	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(theme.Border).
		Width(boxW).
		Render(bar)
}

// Footer renders a compact 2-line footer (top border + content row) split into
// three slots: left = author/copyright, center = beian + links, right = operation hints.
func Footer(theme Theme, width int, hints string) string {
	cfg := defaultFooter
	muted := lipgloss.NewStyle().Foreground(theme.Muted)
	primary := lipgloss.NewStyle().Foreground(theme.Primary)

	boxW := width
	if boxW < 20 {
		boxW = 20
	}

	leftText := cfg.Author + " · " + cfg.Copy
	centerText := cfg.Beian + " 开往 🚇 十年之约"

	left := muted.Render(leftText)
	center := muted.Render(centerText) + " " + primary.Render("")
	right := muted.Render(hints)

	// Detect if a single-line footer would overflow; if so use a two-line layout.
	totalW := lipgloss.Width(left) + lipgloss.Width(center) + lipgloss.Width(right)
	if totalW <= boxW {
		slotW := boxW / 3
		leftSlot := padRight(left, slotW)
		centerSlot := centerAlign(center, slotW)
		rightSlot := padLeft(right, slotW)
		row := leftSlot + centerSlot + rightSlot
		return lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), true, false, false, false).
			BorderForeground(theme.Border).
			Width(boxW).
			Render(row)
	}

	// Narrow terminal: author+center on one line, hints on a second line.
	line1 := left + " " + center
	line1W := lipgloss.Width(line1)
	if line1W > boxW {
		line1 = truncate(line1, boxW)
	}
	line2 := right
	if lipgloss.Width(line2) > boxW {
		line2 = truncate(line2, boxW)
	}
	body := line1 + "\n" + line2
	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), true, false, false, false).
		BorderForeground(theme.Border).
		Width(boxW).
		Render(body)
}

func padRight(s string, w int) string {
	pad := w - lipgloss.Width(s)
	if pad < 0 {
		return truncate(s, w)
	}
	return s + strings.Repeat(" ", pad)
}

func padLeft(s string, w int) string {
	pad := w - lipgloss.Width(s)
	if pad < 0 {
		return truncate(s, w)
	}
	return strings.Repeat(" ", pad) + s
}

func centerAlign(s string, w int) string {
	if w <= 0 {
		return s
	}
	pad := w - lipgloss.Width(s)
	if pad <= 0 {
		return truncate(s, w)
	}
	left := pad / 2
	right := pad - left
	return strings.Repeat(" ", left) + s + strings.Repeat(" ", right)
}

// HelpBar renders a centered help line.
func HelpBar(parts []string, theme Theme, width int) string {
	sep := lipgloss.NewStyle().Foreground(theme.Muted).Render("  ")
	styled := make([]string, len(parts))
	for i, p := range parts {
		styled[i] = lipgloss.NewStyle().Foreground(theme.Muted).Render(p)
	}
	return ContentBox(strings.Join(styled, sep), width)
}
