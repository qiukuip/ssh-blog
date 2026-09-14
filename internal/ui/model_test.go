package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/longkun/ssh-blog/internal/blog"
	"github.com/longkun/ssh-blog/internal/config"
)

func renderAt(m tea.Model, w, h int) string {
	updated, _ := m.Update(tea.WindowSizeMsg{Width: w, Height: h})
	return updated.View()
}

func loadSamplePosts(t *testing.T) []blog.Post {
	t.Helper()
	posts, err := blog.LoadDir("../../posts")
	if err != nil {
		t.Fatalf("load posts: %v", err)
	}
	if len(posts) < 4 {
		t.Fatalf("expected at least 4 sample posts, got %d", len(posts))
	}
	return posts
}

func newModel(t *testing.T) tea.Model {
	pages, _ := blog.LoadPages("../../pages")
	return New(Config{
		Cfg:   config.Default(),
		Posts: loadSamplePosts(t),
		Pages: pages,
	})
}

func TestHomeHasNavAndLatest(t *testing.T) {
	m := newModel(t)
	out := renderAt(m, 120, 40)
	if !strings.Contains(out, "首页") {
		t.Fatalf("expected '首页' in nav, got:\n%s", out)
	}
	if !strings.Contains(out, "归档") {
		t.Fatalf("expected '归档' in nav, got:\n%s", out)
	}
	if !strings.Contains(out, "友链") {
		t.Fatalf("expected '友链' in nav, got:\n%s", out)
	}
	if !strings.Contains(out, "留言") {
		t.Fatalf("expected '留言' in nav, got:\n%s", out)
	}
	if !strings.Contains(out, "关于") {
		t.Fatalf("expected '关于' in nav, got:\n%s", out)
	}
	if !strings.Contains(out, "ssh-blog") {
		t.Fatalf("expected site name in top bar, got:\n%s", out)
	}
	if !strings.Contains(out, "Welcome to ssh-blog") {
		t.Fatalf("expected latest post title, got:\n%s", out)
	}
	if !strings.Contains(out, "©") {
		t.Fatalf("expected copyright in footer, got:\n%s", out)
	}
}

func TestNavKeysJumpToPages(t *testing.T) {
	m := newModel(t)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	// 2 -> archive
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	out := m.View()
	if !strings.Contains(out, "📅") {
		t.Fatalf("expected year markers in archive, got:\n%s", out)
	}

	// 5 -> about (falls back to placeholder when pages/about.md missing)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'5'}})
	out = m.View()
	if !strings.Contains(out, "关于") && !strings.Contains(out, "暂无内容") {
		t.Fatalf("expected about page output, got:\n%s", out)
	}

	// esc -> back home
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	out = m.View()
	if !strings.Contains(out, "Welcome to ssh-blog") {
		t.Fatalf("expected home with posts after esc, got:\n%s", out)
	}
}

func TestArchiveSelection(t *testing.T) {
	m := newModel(t)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	// Open archive.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	out := m.View()
	if !strings.Contains(out, "▸ Welcome to ssh-blog") {
		t.Fatalf("expected first post highlighted with ▸, got:\n%s", out)
	}

	// Down arrow → cursor moves to second post.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	out = m.View()
	if strings.Contains(out, "▸ Welcome to ssh-blog") {
		t.Fatalf("cursor should have moved off first post:\n%s", out)
	}
	if !strings.Contains(out, "▸ Building a TUI with Bubble Tea") {
		t.Fatalf("expected second post highlighted with ▸, got:\n%s", out)
	}

	// Enter opens the reader.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	out = m.View()
	if !strings.Contains(out, "Building a TUI with Bubble Tea") {
		t.Fatalf("expected reader for 'Building a TUI...', got:\n%s", out)
	}
	if !strings.Contains(out, "Elm-style architecture") {
		t.Fatalf("expected article body content in reader, got:\n%s", out)
	}

	// Esc returns to archive with the previous selection restored.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	out = m.View()
	if !strings.Contains(out, "▸ Building a TUI with Bubble Tea") {
		t.Fatalf("expected cursor to stay on second post after esc, got:\n%s", out)
	}
}

func TestThemeCycle(t *testing.T) {
	m := newModel(t)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	mm := m.(*Model)
	first := mm.theme.Name
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'T'}})
	mm = m.(*Model)
	if mm.theme.Name == first {
		t.Fatalf("expected theme to change after pressing T, still %q", first)
	}
}

func TestQuitFromHome(t *testing.T) {
	m := newModel(t)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Fatalf("expected quit command on 'q'")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatalf("expected tea.QuitMsg, got %T", cmd())
	}
}

func TestReadingTOC(t *testing.T) {
	m := newModel(t)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	out := m.View()
	if !strings.Contains(out, "目录") {
		t.Fatalf("expected TOC header in reading view, got:\n%s", out)
	}
	if !strings.Contains(out, "📅") {
		t.Fatalf("expected post metadata in reading view, got:\n%s", out)
	}

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	out = m.View()
	if !strings.Contains(out, "Welcome to ssh-blog") {
		t.Fatalf("expected to return to home, got:\n%s", out)
	}
}

func TestMessagesPage(t *testing.T) {
	pages, _ := blog.LoadPages("../../pages")
	var m tea.Model = New(Config{Posts: loadSamplePosts(t), Pages: pages, Waline: nil})
	m, _ = m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'4'}})
	out := m.View()
	if !strings.Contains(out, "留言板") {
		t.Fatalf("expected message board header, got:\n%s", out)
	}
}

func TestExtractTOC(t *testing.T) {
	md := `# Title

intro

## Section A

text

### Subsection

### Other

## Section B

#### not included
`
	toc := blog.ExtractTOC(md)
	if len(toc) != 4 {
		t.Fatalf("expected 4 headings (h1 and h4 skipped), got %d: %+v", len(toc), toc)
	}
	if toc[0].Level != 2 || toc[0].Text != "Section A" {
		t.Errorf("unexpected first heading: %+v", toc[0])
	}
	if toc[2].Level != 3 || toc[2].Text != "Other" {
		t.Errorf("unexpected 3rd heading: %+v", toc[2])
	}
}

func TestSearchFiltersList(t *testing.T) {
	m := newModel(t)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	for _, r := range "wish" {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	out := m.View()
	if strings.Contains(out, "Welcome to ssh-blog") {
		t.Fatalf("expected Welcome filtered out, got:\n%s", out)
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
}

func TestFooterHasLinks(t *testing.T) {
	m := newModel(t)
	out := renderAt(m, 120, 40)
	if !strings.Contains(out, "开往") {
		t.Fatalf("expected '开往' in footer, got:\n%s", out)
	}
	if !strings.Contains(out, "十年之约") {
		t.Fatalf("expected '十年之约' in footer, got:\n%s", out)
	}
}

func TestStripHTML(t *testing.T) {
	in := "<p>hello <b>world</b></p>"
	if got := blog.StripHTML(in); got != "hello world" {
		t.Fatalf("StripHTML: got %q", got)
	}
}

func TestLoadPagesEmpty(t *testing.T) {
	pages, err := blog.LoadPages("../../nonexistent-dir")
	if err != nil {
		t.Fatalf("missing dir should not error: %v", err)
	}
	if len(pages) != 0 {
		t.Fatalf("expected empty map for missing dir, got %d", len(pages))
	}
}

func TestGlobalNav(t *testing.T) {
	m := newModel(t)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	// Start on home, jump to archive (2), then friends (3) without returning home.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	if out := m.View(); !strings.Contains(out, "归档") {
		t.Fatalf("expected archive after pressing 2, got:\n%s", out)
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})
	if out := m.View(); !strings.Contains(out, "友链") {
		t.Fatalf("expected friends after pressing 3, got:\n%s", out)
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'5'}})
	if out := m.View(); !strings.Contains(out, "关于") {
		t.Fatalf("expected about after pressing 5, got:\n%s", out)
	}

	// Enter reading mode, then press 1 to go home.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	out := m.View()
	if !strings.Contains(out, "📖") {
		t.Fatalf("expected reading view, got:\n%s", out)
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	out = m.View()
	if !strings.Contains(out, "最新文章") {
		t.Fatalf("expected home (最新文章) after pressing 1 in reading, got:\n%s", out)
	}
}

func TestContentWidthMatchesReading(t *testing.T) {
	m := newModel(t)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 200, Height: 30})

	home := m.View()
	lines := strings.Split(home, "\n")
	if len(lines) < 6 {
		t.Fatalf("view has too few lines: %d", len(lines))
	}
	// Drop top bar (lines 0–1) and footer (last 2 lines); only inspect body.
	body := lines[2 : len(lines)-2]
	maxW := 0
	for _, l := range body {
		t := strings.TrimSpace(l)
		if t == "" || isBorderLine(t) {
			continue
		}
		w := lipgloss.Width(t)
		if w > maxW {
			maxW = w
		}
	}
	cm := config.Default().Layout.ContentMeasure
	if maxW > cm+8 {
		t.Fatalf("home body content width %d exceeds %d+8 (should be ~%d)", maxW, cm, cm)
	}
}

func isBorderLine(s string) bool {
	nonBorder := 0
	for _, r := range s {
		switch r {
		case '─', '╭', '╮', '╰', '╯', '═':
		default:
			nonBorder++
		}
	}
	// A border line has at most 2 non-border chars (e.g.╭─╮ with corner).
	return nonBorder <= 2
}

func TestContentMeasureWidth(t *testing.T) {
	dims := dimsFromConfig(config.Default().Layout)
	cases := []struct {
		term int
		want int
	}{
		{200, 90}, // well above cap → cap
		{94, 90},  // exactly at cap (ContentMeasure + OuterPadding) → cap
		{93, 89},  // just below cap → terminalW - OuterPadding
		{84, 80},  // narrow terminal → terminalW - OuterPadding
	}
	for _, c := range cases {
		if got := ContentMeasureWidth(c.term, dims); got != c.want {
			t.Errorf("ContentMeasureWidth(%d) = %d, want %d", c.term, got, c.want)
		}
	}
}
