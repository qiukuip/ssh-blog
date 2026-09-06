package ui

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"

	"github.com/longkun/ssh-blog/internal/blog"
)

type state int

const (
	stateBrowse state = iota
	stateReading
	stateSearch
	stateComposing
	stateSubmitting
)

type postItem struct{ post blog.Post }

func (p postItem) Title() string       { return p.post.Title }
func (p postItem) Description() string { return p.post.Header() }
func (p postItem) FilterValue() string {
	return strings.ToLower(p.post.Title + " " + p.post.Summary + " " + p.post.TagsString() + " " + p.post.Category)
}

type Model struct {
	state     state
	page      int
	width     int
	height    int
	bodyH     int
	err       error
	theme     Theme
	themeIdx  int
	themes    []Theme
	renderer  *glamour.TermRenderer
	rendererW int

	posts    []blog.Post
	allPosts []blog.Post
	pages    map[string]blog.Page

	homeList    list.Model
	archiveList list.Model
	search      textinput.Model

	reading     viewport.Model
	currentPost *blog.Post
	toc         []blog.Heading
	tocIdx      int

	// archiveOrder holds posts in the order they're rendered in the archive
	// view (grouped by year). archiveIdx is the cursor for arrow-key nav.
	archiveOrder []blog.Post
	archiveIdx   int

	messages    []blog.WalineComment
	msgLoaded   bool
	msgLoading  bool
	msgErr      string
	msgCache    []blog.WalineComment
	msgCacheAt  time.Time
	msgCacheTTL time.Duration

	// compose fields
	compNick  textinput.Model
	compMail  textinput.Model
	compLink  textinput.Model
	compBody  textinput.Model
	compField int

	// spinner
	spinner spinner.Model

	waline *blog.WalineClient
}

type Config struct {
	Posts  []blog.Post
	Pages  map[string]blog.Page
	Waline *blog.WalineClient
}

func New(cfg Config) *Model {
	themes := append([]Theme{}, themes...)

	delegate := newPostDelegate(themes[0])
	hl := list.New(nil, delegate, 0, 0)
	hl.Title = ""
	hl.SetShowStatusBar(false)
	hl.SetFilteringEnabled(false)
	hl.DisableQuitKeybindings()
	hl.Styles.HelpStyle = lipgloss.NewStyle().Foreground(themes[0].Muted)

	al := list.New(nil, newPostDelegate(themes[0]), 0, 0)
	al.Title = ""
	al.SetShowStatusBar(false)
	al.SetFilteringEnabled(false)
	al.DisableQuitKeybindings()

	si := textinput.New()
	si.Placeholder = "搜索标题/标签/正文…"
	si.Prompt = "/ "
	si.CharLimit = 64

	vp := viewport.New(0, 0)

	sp := spinner.New(spinner.WithSpinner(spinner.Dot))
	sp.Style = lipgloss.NewStyle().Foreground(themes[0].Primary)

	nick := textinput.New()
	nick.Placeholder = "昵称 *"
	nick.CharLimit = 32
	nick.Prompt = "> "

	mail := textinput.New()
	mail.Placeholder = "邮箱（可选）"
	mail.CharLimit = 64
	mail.Prompt = "> "

	link := textinput.New()
	link.Placeholder = "主页（可选）"
	link.CharLimit = 128
	link.Prompt = "> "

	body := textinput.New()
	body.Placeholder = "说点什么…（多行用 \\n）"
	body.CharLimit = 400
	body.Prompt = "> "

	// Build archive order: posts grouped by year, newest year first.
	var archiveOrder []blog.Post
	if len(cfg.Posts) > 0 {
		groups := blog.GroupPostsByYear(cfg.Posts)
		for _, y := range blog.Keys(groups) {
			archiveOrder = append(archiveOrder, groups[y]...)
		}
	}

	m := &Model{
		state:        stateBrowse,
		page:         pageHome,
		theme:        themes[0],
		themeIdx:     0,
		themes:       themes,
		posts:        cfg.Posts,
		allPosts:     cfg.Posts,
		pages:        cfg.Pages,
		homeList:     hl,
		archiveList:  al,
		search:       si,
		reading:      vp,
		spinner:      sp,
		compNick:     nick,
		compMail:     mail,
		compLink:     link,
		compBody:     body,
		waline:       cfg.Waline,
		archiveOrder: archiveOrder,
		width:        80,
		height:       24,
		msgCacheTTL: 30 * time.Second,
	}
	m.buildHomeList()
	m.buildArchiveList()
	m.applyThemeToLists()
	return m
}

type postDelegate struct{ theme Theme }

func newPostDelegate(t Theme) postDelegate { return postDelegate{theme: t} }

func (d postDelegate) Height() int  { return 2 }
func (d postDelegate) Spacing() int { return 1 }
func (d postDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd {
	return nil
}

func (d postDelegate) Render(w io.Writer, m list.Model, index int, it list.Item) {
	p := it.(postItem).post
	title := p.Title
	if index == m.Index() {
		title = lipgloss.NewStyle().
			Bold(true).
			Foreground(d.theme.Highlight).
			Render("▸ " + title)
	} else {
		title = lipgloss.NewStyle().
			Foreground(d.theme.Title).
			Render("  " + title)
	}
	meta := lipgloss.NewStyle().
		Foreground(d.theme.Muted).
		Render(p.DisplayDate())
	fmt.Fprintf(w, "%s\n  %s", title, meta)
}

func (m *Model) applyThemeToLists() {
	d1 := newPostDelegate(m.theme)
	d2 := newPostDelegate(m.theme)
	m.homeList.SetDelegate(d1)
	m.archiveList.SetDelegate(d2)
	m.spinner.Style = lipgloss.NewStyle().Foreground(m.theme.Primary)
}

func (m *Model) buildHomeList() {
	posts := m.allPosts
	if len(posts) > 10 {
		posts = posts[:10]
	}
	items := make([]list.Item, len(posts))
	for i, p := range posts {
		items[i] = postItem{post: p}
	}
	m.homeList.SetItems(items)
	if len(items) > 0 {
		m.homeList.Select(0)
	}
}

func (m *Model) buildArchiveList() {
	items := make([]list.Item, len(m.allPosts))
	for i, p := range m.allPosts {
		items[i] = postItem{post: p}
	}
	m.archiveList.SetItems(items)
	if len(items) > 0 {
		m.archiveList.Select(0)
	}
}

func (m *Model) buildListFrom(posts []blog.Post, l *list.Model) {
	items := make([]list.Item, len(posts))
	for i, p := range posts {
		items[i] = postItem{post: p}
	}
	l.SetItems(items)
	if len(items) > 0 {
		l.Select(0)
	}
}

func (m *Model) rebuildRenderer() {
	w := m.width - 8
	if m.width >= MinContentWidth+OuterPadding*2 {
		w = PageWidth(m.width) - 8
	}
	if w < 30 {
		w = 30
	}
	if m.rendererW == w && m.renderer != nil {
		return
	}
	m.rendererW = w
	r, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle("notty"),
		glamour.WithWordWrap(w),
		glamour.WithEmoji(),
	)
	if err != nil {
		m.renderer = nil
		return
	}
	m.renderer = r
}

func (m *Model) cycleTheme() {
	m.themeIdx = (m.themeIdx + 1) % len(m.themes)
	m.theme = m.themes[m.themeIdx]
	m.applyThemeToLists()
	m.renderer = nil
	m.rendererW = 0
	m.rebuildRenderer()
}

func (m *Model) enterReading() {
	if m.page == pageArchive {
		if m.archiveIdx >= 0 && m.archiveIdx < len(m.archiveOrder) {
			m.openPost(m.archiveOrder[m.archiveIdx])
		}
		return
	}
	if m.page != pageHome {
		return
	}
	l := m.activeList()
	if l == nil {
		return
	}
	it, ok := l.SelectedItem().(postItem)
	if !ok {
		return
	}
	m.openPost(it.post)
}

func (m *Model) openPost(p blog.Post) {
	m.currentPost = &p
	m.toc = blog.ExtractTOC(p.Content)
	m.tocIdx = 0

	body := p.Content
	rendered := body
	if m.renderer != nil {
		if r, err := m.renderer.Render(body); err == nil {
			rendered = r
		}
	}
	header := ""
	if m.renderer != nil {
		if h, err := m.renderer.Render(p.Header()); err == nil {
			header = h
		}
	}
	m.reading.SetContent(header + "\n" + rendered)
	m.reading.GotoTop()
	m.state = stateReading
}

func (m *Model) activeList() *list.Model {
	if m.page == pageHome {
		return &m.homeList
	}
	if m.page == pageArchive {
		return &m.archiveList
	}
	return nil
}

func (m *Model) setPage(p int) tea.Cmd {
	m.page = p
	if p == pageMessages {
		return m.loadMessages()
	}
	return nil
}

// navByDigit switches to the nav slot matching digit '1'..'5'. Used by the
// global 1-5 page nav so it works from any state except text input.
func (m *Model) navByDigit(d byte) tea.Cmd {
	pages := []int{pageHome, pageArchive, pageFriends, pageMessages, pageAbout}
	target := pages[int(d-'1')]
	// Drop back to browse so the target page actually renders; setPage
	// alone wouldn't escape reading/composing states.
	m.state = stateBrowse
	return m.setPage(target)
}

func (m Model) Init() tea.Cmd { return nil }

type messagesLoadedMsg struct {
	msgs []blog.WalineComment
	err  error
}

type messagePostedMsg struct {
	err error
}

func (m *Model) loadMessages() tea.Cmd {
	if m.waline == nil || m.msgLoading {
		return nil
	}
	// Serve from cache if fresh.
	if m.msgLoaded && len(m.msgCache) > 0 && time.Since(m.msgCacheAt) < m.msgCacheTTL {
		m.messages = m.msgCache
		return nil
	}
	m.msgLoading = true
	c := m.waline
	path := "messages/index.html"
	return func() tea.Msg {
		msgs, err := c.List(path)
		return messagesLoadedMsg{msgs: msgs, err: err}
	}
}

func (m *Model) postMessage() tea.Cmd {
	c := m.waline
	nick := m.compNick.Value()
	mail := m.compMail.Value()
	link := m.compLink.Value()
	body := m.compBody.Value()
	return func() tea.Msg {
		err := c.Post(nick, mail, link, body, "messages/index.html")
		return messagePostedMsg{err: err}
	}
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if msg.Width == 0 || msg.Height == 0 {
			return m, nil
		}
		m.width = msg.Width
		m.height = msg.Height
		m.resizeComponents()
		return m, nil

	case messagesLoadedMsg:
		m.msgLoading = false
		m.msgLoaded = true
		if msg.err != nil {
			m.msgErr = msg.err.Error()
		} else {
			m.msgErr = ""
			m.messages = msg.msgs
			m.msgCache = msg.msgs
			m.msgCacheAt = time.Now()
		}
		return m, nil

	case messagePostedMsg:
		m.state = stateBrowse
		// Invalidate cache so the new message shows on next visit.
		m.msgCacheAt = time.Time{}
		if msg.err != nil {
			m.msgErr = msg.err.Error()
			return m, nil
		}
		m.msgErr = ""
		m.compNick.Reset()
		m.compMail.Reset()
		m.compLink.Reset()
		m.compBody.Reset()
		m.compField = 0
		m.msgLoaded = false
		return m, m.loadMessages()

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

func (m *Model) resizeComponents() {
	// Chrome: top bar (2) + footer (2) = 4
	const chrome = 4
	bodyH := m.height - chrome
	if bodyH < 6 {
		bodyH = 6
	}
	listW := PageWidth(m.width) - 4
	listH := bodyH
	if listH < 4 {
		listH = 4
	}
	m.homeList.SetSize(listW, listH)
	m.archiveList.SetSize(listW, listH)
	m.reading.Width = listW
	m.reading.Height = bodyH
	m.bodyH = bodyH
	m.rebuildRenderer()
}

func (m *Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// Global nav: 1-5 always works except while typing (compose/search).
	// In composing/search states the digits are passed through to the text
	// input so the user can type numbers in messages / search.
	if len(key) == 1 && key[0] >= '1' && key[0] <= '5' {
		if m.state != stateComposing && m.state != stateSearch {
			return m, m.navByDigit(key[0])
		}
	}

	switch m.state {
	case stateComposing:
		return m.handleComposeKey(key)
	case stateReading:
		// Global nav also works in reading mode.
		if len(key) == 1 && key[0] >= '1' && key[0] <= '5' {
			return m, m.navByDigit(key[0])
		}
		return m.handleReadingKey(msg)
	case stateSearch:
		return m.handleSearchKey(msg.String())
	case stateSubmitting:
		if key == "esc" {
			m.state = stateBrowse
		}
		return m, nil
	}

	// Browse / page nav
	switch key {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "T":
		m.cycleTheme()
		return m, nil
	case "t":
		// legacy tag browse -> removed, but keep `t` as theme cycle alias
		m.cycleTheme()
		return m, nil
	case "esc":
		if m.page != pageHome {
			cmd := m.setPage(pageHome)
			return m, cmd
		}
	case "/":
		if m.page == pageHome || m.page == pageArchive {
			m.state = stateSearch
			m.search.Focus()
			m.search.SetValue("")
			m.buildListFrom(m.allPosts, m.activeList())
			return m, textinput.Blink
		}
	case "enter":
		if m.page == pageHome || m.page == pageArchive {
			m.enterReading()
			return m, nil
		}
		if m.page == pageMessages {
			m.state = stateComposing
			m.compField = 0
			m.compNick.Focus()
			return m, nil
		}
	}

	if l := m.activeList(); l != nil && m.page != pageArchive {
		var cmd tea.Cmd
		*l, cmd = l.Update(msg)
		return m, cmd
	}

	// Archive page: custom arrow-key navigation over archiveOrder.
	if m.page == pageArchive && len(m.archiveOrder) > 0 {
		switch key {
		case "up", "k":
			if m.archiveIdx > 0 {
				m.archiveIdx--
			}
			return m, nil
		case "down", "j":
			if m.archiveIdx < len(m.archiveOrder)-1 {
				m.archiveIdx++
			}
			return m, nil
		}
	}

	return m, nil
}

func (m *Model) handleSearchKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "esc":
		m.state = stateBrowse
		m.search.Blur()
		m.buildListFrom(m.allPosts, m.activeList())
		return m, nil
	case "enter":
		m.state = stateBrowse
		m.search.Blur()
		return m, nil
	}
	var cmd tea.Cmd
	m.search, cmd = m.search.Update(keyToMsg(key))
	q := m.search.Value()
	results := blog.Search(m.allPosts, q)
	m.buildListFrom(results, m.activeList())
	return m, cmd
}

func keyToMsg(s string) tea.Msg {
	// textinput accepts tea.KeyMsg directly; convenience for string keys.
	if len(s) == 1 {
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
	}
	switch s {
	case "backspace":
		return tea.KeyMsg{Type: tea.KeyBackspace}
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	case "tab":
		return tea.KeyMsg{Type: tea.KeyTab}
	case "left":
		return tea.KeyMsg{Type: tea.KeyLeft}
	case "right":
		return tea.KeyMsg{Type: tea.KeyRight}
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	case "home":
		return tea.KeyMsg{Type: tea.KeyHome}
	case "end":
		return tea.KeyMsg{Type: tea.KeyEnd}
	}
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

func (m *Model) handleReadingKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc", "b":
		m.state = stateBrowse
		return m, nil
	case "ctrl+f", "pagedown", "space":
		m.reading.PageDown()
		return m, nil
	case "ctrl+b", "pageup":
		m.reading.PageUp()
		return m, nil
	case "n":
		if len(m.toc) == 0 {
			return m, nil
		}
		m.tocIdx = (m.tocIdx + 1) % len(m.toc)
		m.scrollTOCToCurrent()
		return m, nil
	case "p":
		if len(m.toc) == 0 {
			return m, nil
		}
		m.tocIdx = (m.tocIdx - 1 + len(m.toc)) % len(m.toc)
		m.scrollTOCToCurrent()
		return m, nil
	}
	var cmd tea.Cmd
	m.reading, cmd = m.reading.Update(msg)
	return m, cmd
}

func (m *Model) scrollTOCToCurrent() {
	// No-op: real scrolling would need text positions from glamour output.
	// We keep this as a placeholder; visual highlight already follows tocIdx.
}

func (m *Model) handleComposeKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "esc":
		m.state = stateBrowse
		return m, nil
	case "tab", "down":
		m.cycleComposeField(1)
		return m, nil
	case "shift+tab", "up":
		m.cycleComposeField(-1)
		return m, nil
	case "enter":
		if m.compField < 3 {
			m.cycleComposeField(1)
			return m, nil
		}
		if m.compBody.Value() == "" || m.compNick.Value() == "" {
			m.msgErr = "昵称和留言不能为空"
			return m, nil
		}
		m.state = stateSubmitting
		return m, m.postMessage()
	}
	fields := []*textinput.Model{&m.compNick, &m.compMail, &m.compLink, &m.compBody}
	var cmd tea.Cmd
	*fields[m.compField], cmd = fields[m.compField].Update(keyToMsg(key))
	return m, cmd
}

func (m *Model) cycleComposeField(delta int) {
	fields := []*textinput.Model{&m.compNick, &m.compMail, &m.compLink, &m.compBody}
	fields[m.compField].Blur()
	m.compField = (m.compField + delta + len(fields)) % len(fields)
	fields[m.compField].Focus()
}

func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Initializing..."
	}
	bar := TopBar(m.page, m.theme, m.themeIdx, len(m.themes), m.width)
	hints := m.hintLine()
	foot := Footer(m.theme, m.width, hints)

	body := m.viewBody()
	barH := lipgloss.Height(bar)
	footH := lipgloss.Height(foot)
	budget := m.height - barH - footH
	if budget < 1 {
		budget = 1
	}
	body = fitHeight(body, budget)

	return lipgloss.JoinVertical(lipgloss.Left, bar, body, foot)
}

// viewBody returns the rendered body for the current state+page.
func (m Model) viewBody() string {
	switch m.state {
	case stateReading:
		return m.viewReading()
	case stateComposing, stateSubmitting:
		return m.viewMessages()
	}
	switch m.page {
	case pageHome:
		return m.viewHome()
	case pageArchive:
		return m.viewArchive()
	case pageFriends:
		return m.viewPage(pageFriends, "🔗 友链", "friends")
	case pageAbout:
		return m.viewPage(pageAbout, "👤 关于", "about")
	case pageMessages:
		return m.viewMessages()
	}
	return m.viewHome()
}

// fitHeight pads or truncates s so its line count equals target.
func fitHeight(s string, target int) string {
	lines := strings.Count(s, "\n") + 1
	if lines < target {
		return s + strings.Repeat("\n", target-lines)
	}
	if lines > target {
		return truncateLines(s, target)
	}
	return s
}

// hintLine returns the operation-hint text that goes into the footer's right slot.
func (m Model) hintLine() string {
	switch m.state {
	case stateReading:
		return "↑/↓ · ctrl+f/b 翻页 · esc 返回"
	case stateSearch:
		return "输入搜索 · enter 确认 · esc"
	case stateComposing:
		return "tab 切换 · enter 提交 · esc"
	case stateSubmitting:
		return "提交中… · esc 关闭"
	}
	switch m.page {
	case pageHome:
		return "enter 阅读 · / 搜索 · 1-5 · T 主题 · q"
	case pageArchive:
		return "↑/↓ 选择 · enter 阅读 · 1-5 · T 主题 · q"
	case pageMessages:
		return "enter 留言 · esc · T 主题 · q"
	}
	return "esc · T 主题 · 1-5 · q"
}

func (m Model) viewHome() string {
	theme := m.theme
	leftTitle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.PageTitle).
		Render("🏠 关于本站")

	intro := strings.Join([]string{
		"",
		"靠兴趣发电的个人小站。",
		"",
		"在算法推荐的时代，留一片安静的角落。",
		"",
		"试试 `1-5` 切换页面，",
		"或按 `enter` 阅读最新文章。",
	}, "\n")
	var introBox string
	if r, err := m.renderMD(intro); err == nil {
		introBox = r
	} else {
		introBox = intro
	}

	leftContent := leftTitle + "\n" + introBox

	rightTitle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.PageTitle).
		Render("📰 最新文章")
	recent := rightTitle + "\n" + m.homeList.View()

	innerH := m.bodyH
	if innerH < 4 {
		innerH = 4
	}

	leftW := m.homeColWidth()
	rightW := m.rightColWidth()

	leftBlock := lipgloss.NewStyle().Width(leftW).Render(
		truncateLines(leftContent, innerH))
	rightBlock := lipgloss.NewStyle().Width(rightW).Render(
		truncateLines(recent, innerH))

	// Two columns with a vertical separator — no panel borders so the
	// visual style matches the other plain-content pages.
	gap := " │ "
	row := lipgloss.JoinHorizontal(lipgloss.Top, leftBlock, gap, rightBlock)
	return ContentBox(row, m.width)
}

func (m Model) viewArchive() string {
	theme := m.theme
	pageTitle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.PageTitle).
		Render("📅 归档")

	yearStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Primary)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Title)

	titleSelectedStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Highlight)

	dateStyle := lipgloss.NewStyle().
		Foreground(theme.Muted)

	groups := blog.GroupPostsByYear(m.allPosts)
	keys := blog.Keys(groups)

	innerW := PageWidth(m.width) - 4
	if innerW < 30 {
		innerW = 30
	}

	var b strings.Builder
	b.WriteString(pageTitle)
	b.WriteString("\n\n")

	postIdx := 0
	for _, y := range keys {
		b.WriteString(yearStyle.Render(fmt.Sprintf("📅 %d", y)))
		b.WriteString("\n\n")

		for _, p := range groups[y] {
			title := p.Title
			if lipgloss.Width(title) > innerW {
				title = truncateWithEllipsis(title, innerW)
			}
			selected := postIdx == m.archiveIdx
			tStyle := titleStyle
			if selected {
				tStyle = titleSelectedStyle
			}
			if selected {
				b.WriteString("▸ ")
			} else {
				b.WriteString("  ")
			}
			b.WriteString(tStyle.Render(title))
			b.WriteString("\n")
			b.WriteString("  ")
			b.WriteString(dateStyle.Render(p.DisplayDate()))
			b.WriteString("\n")
			postIdx++
		}
		b.WriteString("\n")
	}

	body := b.String()
	return ContentBox(lipgloss.NewStyle().Padding(0, 1).Render(body), m.width)
}

func truncateWithEllipsis(s string, n int) string {
	if n <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= n {
		return s
	}
	if n <= 1 {
		return "…"
	}
	target := n - 1
	var b strings.Builder
	w := 0
	for _, r := range s {
		rw := runewidth.RuneWidth(r)
		if w+rw > target {
			break
		}
		b.WriteRune(r)
		w += rw
	}
	b.WriteString("…")
	return b.String()
}

func (m Model) viewPage(_ int, title, slug string) string {
	theme := m.theme
	pageTitle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.PageTitle).
		Render(title)
	p, ok := m.pages[slug]
	if !ok {
		note := lipgloss.NewStyle().Foreground(theme.Muted).Render(
			"暂无内容，请在 pages/" + slug + ".md 添加 Markdown")
		return ContentBox(pageTitle+"\n\n"+note, m.width)
	}
	body := p.Content
	if r, err := m.renderMD(body); err == nil {
		body = r
	}
	return ContentBox(pageTitle+"\n\n"+body, m.width)
}

func (m Model) viewReading() string {
	if m.currentPost == nil {
		return ""
	}
	theme := m.theme

	contentW := ContentMeasureWidth(m.width)
	if contentW < 60 {
		contentW = 60
	}

	var b strings.Builder
	// Page title (always above; spans full content width).
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.PageTitle).
		Render("📖 " + m.currentPost.Title)
	b.WriteString(title)
	b.WriteString("\n\n")

	// TOC on the left, content on the right; total = contentW.
	// TOC shrinks gracefully on narrower content.
	tocW := 0
	leftBlock := ""
	if len(m.toc) > 0 {
		tocW = contentW / 5
		if tocW < 18 {
			tocW = 18
		}
		if tocW > 32 {
			tocW = 32
		}
		leftBlock = m.renderTOC(tocW)
	}

	articleW := contentW - tocW - 2
	if articleW < 30 {
		articleW = 30
	}

	inner := m.reading.View()
	if articleW > 0 {
		inner = lipgloss.NewStyle().Width(articleW).Render(inner)
	}

	var row string
	if tocW > 0 && len(m.toc) > 0 {
		row = lipgloss.JoinHorizontal(lipgloss.Top,
			lipgloss.NewStyle().Width(tocW).Render(leftBlock),
			"  ",
			inner,
		)
	} else {
		row = inner
	}

	b.WriteString(row)

	// Center the whole reading block (TOC + content) within terminal.
	pad := ContentMeasurePad(m.width)
	if pad > 0 {
		return padLinesLeft(b.String(), pad)
	}
	return b.String()
}

func padLinesLeft(s string, n int) string {
	pad := strings.Repeat(" ", n)
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		lines[i] = pad + l
	}
	return strings.Join(lines, "\n")
}

// renderTOC returns the table of contents for the current post as a
// bordered sidebar block of the given width. Used on the left of the
// reading view so the whole reading area stays at ContentMeasure.
func (m Model) renderTOC(width int) string {
	theme := m.theme
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Primary).
		Padding(0, 1).
		Render("📑 目录")
	var b strings.Builder
	b.WriteString(header)
	for i, h := range m.toc {
		style := lipgloss.NewStyle().Foreground(theme.Muted)
		marker := "  "
		if i == m.tocIdx {
			marker = "▸ "
			style = lipgloss.NewStyle().Foreground(theme.Highlight).Bold(true)
		}
		indent := strings.Repeat("  ", h.Level-1)
		text := truncate(h.Text, width-4-len(indent)-2)
		b.WriteString(style.Render(indent+marker+text) + "\n")
	}
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Border).
		Padding(0, 1).
		Width(width - 2).
		Render(strings.TrimRight(b.String(), "\n"))
	return box
}

func (m Model) viewMessages() string {
	theme := m.theme
	var b strings.Builder

	pageTitle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.PageTitle).
		Render("💬 留言板")
	b.WriteString(pageTitle + "\n\n")

	if m.msgErr != "" {
		errLine := lipgloss.NewStyle().Foreground(theme.Error).Bold(true).Render("⚠ " + m.msgErr)
		b.WriteString(errLine + "\n\n")
	}

	if !m.msgLoaded && !m.msgLoading {
		b.WriteString(lipgloss.NewStyle().Foreground(theme.Muted).Render("正在加载留言…") + "\n")
	} else if m.msgLoading {
		b.WriteString(m.spinner.View() + " 加载留言中…\n")
	} else {
		if len(m.messages) == 0 {
			b.WriteString(lipgloss.NewStyle().Foreground(theme.Muted).Render("还没有留言，留下第一条吧。") + "\n")
		} else {
			for _, msg := range m.messages {
				b.WriteString(m.renderComment(msg, 0))
				for _, ch := range msg.Children {
					b.WriteString(m.renderComment(ch, 1))
				}
			}
		}
	}

	b.WriteString("\n")

	if m.state == stateComposing || m.state == stateSubmitting {
		b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(theme.Secondary).Render("✎ 我要留言") + "\n")
		b.WriteString(m.compFieldView(0, "昵称 *", &m.compNick) + "\n")
		b.WriteString(m.compFieldView(1, "邮箱", &m.compMail) + "\n")
		b.WriteString(m.compFieldView(2, "主页", &m.compLink) + "\n")
		b.WriteString(m.compFieldView(3, "留言内容", &m.compBody) + "\n")
		if m.state == stateSubmitting {
			b.WriteString(m.spinner.View() + " 正在发送…\n")
		} else {
			b.WriteString(lipgloss.NewStyle().Foreground(theme.Muted).Render("tab 切换字段，末段 enter 提交") + "\n")
		}
	} else {
		b.WriteString(lipgloss.NewStyle().Foreground(theme.Secondary).Bold(true).Render("[ enter ] 写留言") + "\n")
	}

	return ContentBox(lipgloss.NewStyle().Padding(0, 1).Render(b.String()), m.width)
}

func (m Model) compFieldView(idx int, label string, ti *textinput.Model) string {
	theme := m.theme
	marker := "  "
	style := lipgloss.NewStyle().Foreground(theme.Muted)
	if m.compField == idx {
		marker = "▸ "
		style = lipgloss.NewStyle().Foreground(theme.Highlight).Bold(true)
	}
	return style.Render(marker+label+": ") + ti.View()
}

func (m Model) renderComment(c blog.WalineComment, depth int) string {
	theme := m.theme
	indent := strings.Repeat("  ", depth+1)
	nick := c.Nick
	if c.Type == "administrator" {
		nick = lipgloss.NewStyle().Foreground(theme.Accent).Bold(true).Render("★ " + nick)
	} else {
		nick = lipgloss.NewStyle().Bold(true).Foreground(theme.Secondary).Render(nick)
	}
	t := c.InsertedAt.Local().Format("2006-01-02 15:04")
	head := indent + nick + "  " + lipgloss.NewStyle().Foreground(theme.Muted).Render(t)
	body := indent + "  " + blog.StripHTML(c.Comment)
	divider := indent + "  " + lipgloss.NewStyle().Foreground(theme.Border).Render(strings.Repeat("─", 40))
	return head + "\n" + body + "\n" + divider + "\n"
}

// homeColWidth returns the content width of the left home panel.
// The two panels + gap + borders must fit within PageWidth so the joined row
// doesn't wrap and double its height.
func (m Model) homeColWidth() int {
	cw := PageWidth(m.width)
	// 2 borders (left+right of each panel) = 4, gap = 2 → fixed 6
	totalContent := cw - 6
	if totalContent < 8 {
		totalContent = 8
	}
	left := totalContent / 3
	if left < 4 {
		left = 4
	}
	return left
}

func (m Model) rightColWidth() int {
	cw := PageWidth(m.width)
	totalContent := cw - 6
	if totalContent < 8 {
		totalContent = 8
	}
	right := totalContent - m.homeColWidth()
	if right < 4 {
		right = 4
	}
	return right
}

func (m Model) renderMD(s string) (string, error) {
	if m.renderer == nil {
		return s, nil
	}
	return m.renderer.Render(s)
}

func truncate(s string, n int) string {
	if n <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	if n <= 1 {
		return "…"
	}
	return string(r[:n-1]) + "…"
}

func truncateLines(s string, maxLines int) string {
	if maxLines <= 0 {
		return ""
	}
	lines := strings.Split(s, "\n")
	if len(lines) <= maxLines {
		return s
	}
	return strings.Join(lines[:maxLines], "\n")
}


func (m *Model) RenderMDDebug(s string) (string, error) { return m.renderMD(s) }

func (m *Model) ViewHomeDebug() string { return m.viewHome() }

func (m Model) viewHomeDebugBodyH() int { return m.bodyH }
func (m Model) viewHomeDebugHomeW() int { return m.homeColWidth() }
func (m Model) viewHomeDebugRightW() int { return m.rightColWidth() }

func (m Model) StateDebug() int  { return int(m.state) }
func (m Model) PageDebug() int   { return m.page }
