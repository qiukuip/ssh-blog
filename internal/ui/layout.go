package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/longkun/ssh-blog/internal/config"
)

// Page slot indices. The order MUST match the nav config: home, archive,
// friends, messages, about. Keep these stable so legacy page-int math (e.g.
// archiveIdx handling) still works.
const (
	pageHome = iota
	pageArchive
	pageFriends
	pageMessages
	pageAbout
)

// pageCount is the number of built-in page slots. Slugs that come from the
// nav config are mapped into these slots by index; custom slugs need
// extended handling in the model layer.
const pageCount = 10

type NavItem struct {
	Key   string
	Label string
	Page  int
	Slug  string
}

// NavItems builds the nav strip from the YAML config. Order in cfg.Nav
// determines order of digit keys (1..5).
func NavItems(cfgs []config.NavConfig) []NavItem {
	if len(cfgs) == 0 {
		return nil
	}
	out := make([]NavItem, len(cfgs))
	for i, n := range cfgs {
		page := i
		if page >= pageCount {
			page = pageHome
		}
		out[i] = NavItem{Key: n.Key, Label: n.Label, Page: page, Slug: n.Slug}
	}
	return out
}

// layoutDims collects the layout constants from config so call sites don't
// have to read them off the config each time.
type layoutDims struct {
	contentMeasure   int
	minContentWidth  int
	maxContentWidth  int
	outerPadding     int
	minTerminalWidth int
	narrowNavThresh  int
	chromeHeight     int
	minBodyHeight    int
	minListHeight    int
}

func dimsFromConfig(lc config.LayoutConfig) layoutDims {
	return layoutDims{
		contentMeasure:   lc.ContentMeasure,
		minContentWidth:  lc.MinContentWidth,
		maxContentWidth:  lc.MaxContentWidth,
		outerPadding:     lc.OuterPadding,
		minTerminalWidth: lc.MinTerminalWidth,
		narrowNavThresh:  lc.NarrowNavThresh,
		chromeHeight:     lc.ChromeHeight,
		minBodyHeight:    lc.MinBodyHeight,
		minListHeight:    lc.MinListHeight,
	}
}

// PageWidth returns the inner content width given a terminal width,
// clamping to [MinContentWidth, MaxContentWidth]. The reading view and the
// rest of the body share the same content measure so layouts line up.
func PageWidth(terminalW int, d layoutDims) int {
	w := terminalW - d.outerPadding*2
	if w < d.minContentWidth {
		w = terminalW
	}
	if w > d.maxContentWidth {
		w = d.maxContentWidth
	}
	if w < d.minTerminalWidth {
		w = d.minTerminalWidth
	}
	return w
}

// ContentMeasureWidth returns the standard content column width for the
// given terminal width. Caps at ContentMeasure on wide terminals; otherwise
// leaves room for a 2-cell gutter on each side so the content stays centered.
func ContentMeasureWidth(terminalW int, d layoutDims) int {
	if terminalW >= d.contentMeasure+d.outerPadding {
		return d.contentMeasure
	}
	return terminalW - d.outerPadding
}

// ContentMeasurePad returns the horizontal padding needed to center a
// content-measure-wide block in a terminal of the given width.
func ContentMeasurePad(terminalW int, d layoutDims) int {
	w := ContentMeasureWidth(terminalW, d)
	if terminalW <= w {
		return 0
	}
	return (terminalW - w) / 2
}

// ContentBox returns the content rendered centered within the terminal width.
func ContentBox(content string, terminalW int, d layoutDims) string {
	inner := PageWidth(terminalW, d)
	if terminalW < d.minContentWidth+d.outerPadding*2 {
		return content
	}
	pad := (terminalW - inner) / 2
	if pad < 0 {
		pad = 0
	}
	return lipgloss.NewStyle().
		Width(terminalW).
		Padding(0, pad, 0, pad).
		Render(content)
}

// TopBar renders the navigation bar. siteName and nav come from config.
func TopBar(currentPage int, siteName string, nav []NavItem, theme Theme, themeIdx, totalThemes int, width int, hint string) string {
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Primary).
		Render(siteName)

	var navParts []string
	for _, n := range nav {
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
	navStr := lipgloss.NewStyle().Render(strings.Join(navParts, " "))

	themeHint := lipgloss.NewStyle().
		Foreground(theme.Muted).
		Italic(true).
		Render(fmt.Sprintf("[T] %s %d/%d", theme.Name, themeIdx+1, totalThemes))

	left := title
	right := themeHint
	if hint != "" {
		right = hint + "  " + right
	}
	boxW := width - 2
	if boxW < 20 {
		boxW = 20
	}
	contentW := boxW - 2

	used := lipgloss.Width(left) + lipgloss.Width(navStr) + lipgloss.Width(right) + 1
	gap := contentW - used
	if gap < 1 {
		if width < 80 {
			navParts = navParts[:0]
			for _, n := range nav {
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
			navStr = lipgloss.NewStyle().Render(strings.Join(navParts, "·"))
		}
		used = lipgloss.Width(left) + lipgloss.Width(navStr) + lipgloss.Width(right) + 1
		gap = contentW - used
		if gap < 1 {
			gap = 1
		}
	}

	bar := left + strings.Repeat(" ", gap) + navStr + " " + right

	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(theme.Border).
		Width(boxW).
		Render(bar)
}

// Footer renders a compact 2-line footer (top border + content row) split
// into three slots: left = author/copyright, center = beian + links,
// right = operation hints.
func Footer(theme Theme, width int, hints string, fc config.FooterConfig) string {
	muted := lipgloss.NewStyle().Foreground(theme.Muted)
	primary := lipgloss.NewStyle().Foreground(theme.Primary)

	boxW := width
	if boxW < 20 {
		boxW = 20
	}

	leftText := fc.Author + fc.CenterJoiner + fc.Copyright
	sep := fc.Separator
	if sep == "" {
		sep = " / "
	}
	centerText := fc.Beian + sep + fc.KaiwangLabel + sep + fc.ShinianLabel

	left := muted.Render(leftText)
	center := muted.Render(centerText) + " " + primary.Render("")
	right := muted.Render(hints)

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
func HelpBar(parts []string, theme Theme, width int, d layoutDims) string {
	sep := lipgloss.NewStyle().Foreground(theme.Muted).Render("  ")
	styled := make([]string, len(parts))
	for i, p := range parts {
		styled[i] = lipgloss.NewStyle().Foreground(theme.Muted).Render(p)
	}
	return ContentBox(strings.Join(styled, sep), width, d)
}
