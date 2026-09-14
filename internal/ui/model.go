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
	"github.com/longkun/ssh-blog/internal/config"
)

type state int

const (
	stateBrowse state = iota
	stateReading
	stateSearch
	stateComposing
	stateSubmitting
)

type postItem struct {
	post blog.Post
	meta blog.MetaConfig
}

func (p postItem) Title() string       { return p.post.Title }
func (p postItem) Description() string { return p.post.Header(p.meta) }
func (p postItem) FilterValue() string {
	return strings.ToLower(p.post.Title + " " + p.post.Summary + " " + p.post.TagsString() + " " + p.post.Category)
}

func (m *Model) newPostItem(p blog.Post) postItem {
	return postItem{
		post: p,
		meta: blog.MetaConfig{
			DateFormat:    m.cfg.Post.DateFormat,
			UnknownDate:   m.cfg.Post.UnknownDate,
			DateEmoji:     m.cfg.Post.DateEmoji,
			AuthorEmoji:   m.cfg.Post.AuthorEmoji,
			CategoryEmoji: m.cfg.Post.CategoryEmoji,
			TagEmoji:      m.cfg.Post.TagEmoji,
		},
	}
}

type Model struct {
	cfg *config.Config

	state     state
	page      int
	width     int
	height    int
	bodyH     int
	err       error
	theme     Theme
	themeIdx  int
	themes    []Theme
	nav       []NavItem
	dims      layoutDims
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

	archiveOrder []blog.Post
	archiveIdx   int

	messages    []blog.WalineComment
	msgLoaded   bool
	msgLoading  bool
	msgErr      string
	msgCache    []blog.WalineComment
	msgCacheAt  time.Time
	msgCacheTTL time.Duration
	msgList     viewport.Model

	compNick  textinput.Model
	compMail  textinput.Model
	compLink  textinput.Model
	compBody  textinput.Model
	compField int

	spinner spinner.Model

	waline *blog.WalineClient
}

// Config is what callers (main, tests) pass into New().
type Config struct {
	Cfg     *config.Config
	Posts   []blog.Post
	Pages   map[string]blog.Page
	Waline  *blog.WalineClient
}

func New(cfg Config) *Model {
	if cfg.Cfg == nil {
		cfg.Cfg = config.Default()
	}
	themes := ThemesFromConfig(cfg.Cfg.Themes)
	if len(themes) == 0 {
		themes = DefaultThemes()
	}
	nav := NavItems(cfg.Cfg.Nav)
	dims := dimsFromConfig(cfg.Cfg.Layout)

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

	searchCfg := cfg.Cfg.Search
	si := textinput.New()
	si.Placeholder = searchCfg.Placeholder
	si.Prompt = searchCfg.Prompt
	si.CharLimit = searchCfg.CharLimit

	vp := viewport.New(0, 0)
	mvp := viewport.New(0, 0)
	mvp.MouseWheelEnabled = false

	sp := spinner.New(spinner.WithSpinner(spinner.Dot))
	sp.Style = lipgloss.NewStyle().Foreground(themes[0].Primary)

	formFields := buildFormFields(cfg.Cfg.Pages.Messages.Form)

	var archiveOrder []blog.Post
	if len(cfg.Posts) > 0 {
		groups := blog.GroupPostsByYear(cfg.Posts)
		for _, y := range blog.Keys(groups) {
			archiveOrder = append(archiveOrder, groups[y]...)
		}
	}

	m := &Model{
		cfg:          cfg.Cfg,
		state:        stateBrowse,
		page:         pageHome,
		theme:        themes[0],
		themeIdx:     0,
		themes:       themes,
		nav:          nav,
		dims:         dims,
		posts:        cfg.Posts,
		allPosts:     cfg.Posts,
		pages:        cfg.Pages,
		homeList:     hl,
		archiveList:  al,
		search:       si,
		reading:      vp,
		msgList:      mvp,
		spinner:      sp,
		compNick:     formFields[0],
		compMail:     formFields[1],
		compLink:     formFields[2],
		compBody:     formFields[3],
		waline:       cfg.Waline,
		archiveOrder: archiveOrder,
		width:        80,
		height:       24,
		msgCacheTTL:  cfg.Cfg.Waline.CacheTTL,
	}
	m.buildHomeList()
	m.buildArchiveList()
	m.applyThemeToLists()
	return m
}

func buildFormFields(cfgs []config.FormFieldConfig) [4]textinput.Model {
	var out [4]textinput.Model
	if len(cfgs) < 4 {
		cfgs = config.Default().Pages.Messages.Form
	}
	for i := 0; i < 4; i++ {
		ti := textinput.New()
		ti.Placeholder = cfgs[i].Placeholder
		ti.CharLimit = cfgs[i].Limit
		ti.Prompt = cfgs[i].Prompt
		out[i] = ti
	}
	return out
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
	max := m.cfg.Home.RecentCount
	if max <= 0 {
		max = 10
	}
	if len(posts) > max {
		posts = posts[:max]
	}
	items := make([]list.Item, len(posts))
	for i, p := range posts {
		items[i] = m.newPostItem(p)
	}
	m.homeList.SetItems(items)
	if len(items) > 0 {
		m.homeList.Select(0)
	}
}

func (m *Model) buildArchiveList() {
	items := make([]list.Item, len(m.allPosts))
	for i, p := range m.allPosts {
		items[i] = m.newPostItem(p)
	}
	m.archiveList.SetItems(items)
	if len(items) > 0 {
		m.archiveList.Select(0)
	}
}

func (m *Model) buildListFrom(posts []blog.Post, l *list.Model) {
	items := make([]list.Item, len(posts))
	for i, p := range posts {
		items[i] = m.newPostItem(p)
	}
	l.SetItems(items)
	if len(items) > 0 {
		l.Select(0)
	}
}

func (m *Model) rebuildRenderer() {
	w := m.width - m.cfg.Glamour.WordWrapOffset
	if m.width >= m.dims.minContentWidth+m.dims.outerPadding*2 {
		w = PageWidth(m.width, m.dims) - m.cfg.Glamour.WordWrapOffset
	}
	if w < m.cfg.Glamour.MinRendererWidth {
		w = m.cfg.Glamour.MinRendererWidth
	}
	if m.rendererW == w && m.renderer != nil {
		return
	}
	m.rendererW = w
	opts := []glamour.TermRendererOption{
		glamour.WithStandardStyle(m.cfg.Glamour.Style),
		glamour.WithWordWrap(w),
	}
	if m.cfg.Glamour.EnableEmoji {
		opts = append(opts, glamour.WithEmoji())
	}
	r, err := glamour.NewTermRenderer(opts...)
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

	meta := blog.MetaConfig{
		DateFormat:    m.cfg.Post.DateFormat,
		UnknownDate:   m.cfg.Post.UnknownDate,
		DateEmoji:     m.cfg.Post.DateEmoji,
		AuthorEmoji:   m.cfg.Post.AuthorEmoji,
		CategoryEmoji: m.cfg.Post.CategoryEmoji,
		TagEmoji:      m.cfg.Post.TagEmoji,
	}

	body := p.Content
	rendered := body
	if m.renderer != nil {
		if r, err := m.renderer.Render(body); err == nil {
			rendered = r
		}
	}
	header := ""
	if m.renderer != nil {
		if h, err := m.renderer.Render(p.Header(meta)); err == nil {
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
		if m.waline == nil {
			m.state = stateComposing
			m.compField = 0
			m.compNick.Focus()
			m.compMail.Blur()
			m.compLink.Blur()
			m.compBody.Blur()
			m.sizeMessagesViewport()
			m.refreshMessagesContent()
			return nil
		}
		return m.loadMessages()
	}
	return nil
}

// navByDigit switches to the nav slot matching the Nth digit key (1..len(nav)).
// Used by the global digit-key page nav so it works from any state except
// text input.
func (m *Model) navByDigit(d byte) tea.Cmd {
	idx := int(d - '1')
	if idx < 0 || idx >= len(m.nav) {
		return nil
	}
	target := m.nav[idx].Page
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
	if m.msgLoaded && len(m.msgCache) > 0 && time.Since(m.msgCacheAt) < m.msgCacheTTL {
		m.messages = m.msgCache
		return nil
	}
	m.msgLoading = true
	c := m.waline
	path := m.cfg.Waline.MessagesPath
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
	path := m.cfg.Waline.MessagesPath
	return func() tea.Msg {
		if c == nil {
			return messagePostedMsg{err: fmt.Errorf("%s", m.cfg.Pages.Messages.ErrorNotConfigured)}
		}
		err := c.Post(nick, mail, link, body, path)
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
		m.refreshMessagesContent()
		return m, nil

	case messagePostedMsg:
		m.state = stateBrowse
		m.sizeMessagesViewport()
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
	bodyH := m.height - m.dims.chromeHeight
	if bodyH < m.dims.minBodyHeight {
		bodyH = m.dims.minBodyHeight
	}
	listW := PageWidth(m.width, m.dims) - m.cfg.Layout.Gap
	listH := bodyH
	if listH < m.dims.minListHeight {
		listH = m.dims.minListHeight
	}
	m.homeList.SetSize(listW, listH)
	m.archiveList.SetSize(listW, listH)
	m.reading.Width = listW
	m.reading.Height = bodyH
	m.bodyH = bodyH
	m.sizeMessagesViewport()
	m.rebuildRenderer()
}

func (m *Model) sizeMessagesViewport() {
	formH := 1
	if m.state == stateComposing || m.state == stateSubmitting {
		formH = m.cfg.Layout.MessageFormHeight
	}
	listW := PageWidth(m.width, m.dims) - m.cfg.Layout.Gap
	if listW < m.cfg.Layout.MinViewportWidth {
		listW = m.cfg.Layout.MinViewportWidth
	}
	h := m.bodyH - 3 - formH
	if h < m.cfg.Layout.MinViewportHeight {
		h = m.cfg.Layout.MinViewportHeight
	}
	m.msgList.Width = listW
	m.msgList.Height = h
}

func (m *Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	navDigits := m.cfg.Keys.NavDigits
	digits := []byte(navDigits)
	isDigit := len(key) == 1 && key[0] >= digits[0] && key[0] <= digits[len(digits)-1]
	if isDigit {
		for _, d := range digits {
			if key[0] == d {
				if m.state != stateComposing && m.state != stateSearch {
					return m, m.navByDigit(key[0])
				}
				break
			}
		}
	}

	switch m.state {
	case stateComposing:
		return m.handleComposeKey(key)
	case stateReading:
		if isDigit {
			for _, d := range digits {
				if key[0] == d {
					return m, m.navByDigit(key[0])
				}
			}
		}
		return m.handleReadingKey(msg)
	case stateSearch:
		return m.handleSearchKey(msg.String())
	case stateSubmitting:
		if key == m.cfg.Keys.Esc {
			m.state = stateBrowse
			m.sizeMessagesViewport()
		}
		return m, nil
	}

	switch key {
	case "q", "ctrl+c":
		if contains(m.cfg.Keys.Quit, key) {
			return m, tea.Quit
		}
	case "T", "t":
		if contains(m.cfg.Keys.ThemeCycle, key) {
			m.cycleTheme()
			return m, nil
		}
	case "esc":
		if key == m.cfg.Keys.Esc {
			if m.page != pageHome {
				cmd := m.setPage(pageHome)
				return m, cmd
			}
		}
	case "/":
		if key == m.cfg.Keys.Search {
			if m.page == pageHome || m.page == pageArchive {
				m.state = stateSearch
				m.search.Focus()
				m.search.SetValue("")
				m.buildListFrom(m.allPosts, m.activeList())
				return m, textinput.Blink
			}
		}
	case "enter":
		if key == m.cfg.Keys.Enter {
			if m.page == pageHome || m.page == pageArchive {
				m.enterReading()
				return m, nil
			}
			if m.page == pageMessages {
				m.state = stateComposing
				m.compField = 0
				m.compNick.Focus()
				m.sizeMessagesViewport()
				return m, nil
			}
		}
	}

	if l := m.activeList(); l != nil && m.page != pageArchive {
		var cmd tea.Cmd
		*l, cmd = l.Update(msg)
		return m, cmd
	}

	// Messages page: scroll the list viewport (form is always visible below).
	if m.page == pageMessages {
		switch {
		case contains(m.cfg.Keys.ScrollUp, key):
			m.msgList.LineUp(1)
			return m, nil
		case contains(m.cfg.Keys.ScrollDown, key):
			m.msgList.LineDown(1)
			return m, nil
		case contains(m.cfg.Keys.PageDown, key):
			m.msgList.PageDown()
			return m, nil
		case contains(m.cfg.Keys.PageUp, key):
			m.msgList.PageUp()
			return m, nil
		case contains(m.cfg.Keys.Top, key):
			m.msgList.GotoTop()
			return m, nil
		case contains(m.cfg.Keys.Bottom, key):
			m.msgList.GotoBottom()
			return m, nil
		}
	}

	if m.page == pageArchive && len(m.archiveOrder) > 0 {
		switch {
		case contains(m.cfg.Keys.ScrollUp, key):
			if m.archiveIdx > 0 {
				m.archiveIdx--
			}
			return m, nil
		case contains(m.cfg.Keys.ScrollDown, key):
			if m.archiveIdx < len(m.archiveOrder)-1 {
				m.archiveIdx++
			}
			return m, nil
		}
	}

	return m, nil
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

func (m *Model) handleSearchKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case m.cfg.Keys.Esc:
		m.state = stateBrowse
		m.search.Blur()
		m.buildListFrom(m.allPosts, m.activeList())
		return m, nil
	case m.cfg.Keys.Enter:
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
	k := msg.String()
	switch {
	case contains(m.cfg.Keys.ReadingBack, k):
		m.state = stateBrowse
		return m, nil
	case contains(m.cfg.Keys.ScrollUp, k):
		m.reading.LineUp(1)
		return m, nil
	case contains(m.cfg.Keys.ScrollDown, k):
		m.reading.LineDown(1)
		return m, nil
	case contains(m.cfg.Keys.PageDown, k):
		m.reading.PageDown()
		return m, nil
	case contains(m.cfg.Keys.PageUp, k):
		m.reading.PageUp()
		return m, nil
	case contains(m.cfg.Keys.Top, k):
		m.reading.GotoTop()
		return m, nil
	case contains(m.cfg.Keys.Bottom, k):
		m.reading.GotoBottom()
		return m, nil
	case k == m.cfg.Keys.TOCNext:
		if len(m.toc) == 0 {
			return m, nil
		}
		m.tocIdx = (m.tocIdx + 1) % len(m.toc)
		m.scrollTOCToCurrent()
		return m, nil
	case k == m.cfg.Keys.TOCPrev:
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
}

func (m *Model) handleComposeKey(key string) (tea.Model, tea.Cmd) {
	switch {
	case key == m.cfg.Keys.Esc:
		m.state = stateBrowse
		m.sizeMessagesViewport()
		return m, nil
	case key == m.cfg.Keys.ComposeNextField:
		m.cycleComposeField(1)
		return m, nil
	case key == m.cfg.Keys.ComposePrevField:
		m.cycleComposeField(-1)
		return m, nil
	case contains(m.cfg.Keys.ScrollUp, key):
		m.msgList.LineUp(1)
		return m, nil
	case contains(m.cfg.Keys.ScrollDown, key):
		m.msgList.LineDown(1)
		return m, nil
	case contains(m.cfg.Keys.PageDown, key):
		m.msgList.PageDown()
		return m, nil
	case contains(m.cfg.Keys.PageUp, key):
		m.msgList.PageUp()
		return m, nil
	case contains(m.cfg.Keys.Top, key):
		m.msgList.GotoTop()
		return m, nil
	case contains(m.cfg.Keys.Bottom, key):
		m.msgList.GotoBottom()
		return m, nil
	case key == m.cfg.Keys.ComposeSubmit:
		if m.compField < 3 {
			m.cycleComposeField(1)
			return m, nil
		}
		fields := m.cfg.Pages.Messages.Form
		var nickReq, bodyReq bool
		if len(fields) >= 4 {
			nickReq = fields[0].Required
			bodyReq = fields[3].Required
		}
		if (nickReq && m.compNick.Value() == "") || (bodyReq && m.compBody.Value() == "") {
			m.msgErr = m.cfg.Pages.Messages.ValidationEmpty
			return m, nil
		}
		m.state = stateSubmitting
		m.sizeMessagesViewport()
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
		return m.cfg.Server.InitPlaceholder
	}
	bar := TopBar(m.page, m.cfg.Site.Name, m.nav, m.theme, m.themeIdx, len(m.themes), m.width, "")
	hints := m.hintLine()
	foot := Footer(m.theme, m.width, hints, m.cfg.Footer)

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
		return m.viewPage(m.cfg.Pages.Friends.Slug, m.cfg.Pages.Friends.Title, m.cfg.Pages.Friends.Empty)
	case pageAbout:
		return m.viewPage(m.cfg.Pages.About.Slug, m.cfg.Pages.About.Title, m.cfg.Pages.About.Empty)
	case pageMessages:
		return m.viewMessages()
	}
	return m.viewHome()
}

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

func (m Model) hintLine() string {
	h := m.cfg.Hints
	switch m.state {
	case stateReading:
		return h.Reading
	case stateSearch:
		return h.Search
	case stateComposing:
		return h.Composing
	case stateSubmitting:
		return h.Submitting
	}
	switch m.page {
	case pageHome:
		return h.Home
	case pageArchive:
		return h.Archive
	case pageMessages:
		return h.Messages
	}
	return h.Fallback
}

func (m Model) viewHome() string {
	theme := m.theme
	leftTitle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.PageTitle).
		Render(m.cfg.Home.Title)

	intro := m.cfg.Home.Intro
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
		Render(m.cfg.Home.RecentTitle)
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

	gap := " │ "
	row := lipgloss.JoinHorizontal(lipgloss.Top, leftBlock, gap, rightBlock)
	return ContentBox(row, m.width, m.dims)
}

func (m Model) viewArchive() string {
	theme := m.theme
	pageTitle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.PageTitle).
		Render(m.cfg.Pages.Archive.Title)

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

	innerW := PageWidth(m.width, m.dims) - m.cfg.Layout.Gap
	if innerW < m.cfg.Layout.MinViewportWidth {
		innerW = m.cfg.Layout.MinViewportWidth
	}

	var b strings.Builder
	b.WriteString(pageTitle)
	b.WriteString("\n\n")

	postIdx := 0
	for _, y := range keys {
		b.WriteString(yearStyle.Render(fmt.Sprintf("%s %d", m.cfg.Post.DateEmoji, y)))
		b.WriteString("\n\n")

		posts := groups[y]
		for i, p := range posts {
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
			if i < len(posts)-1 {
				b.WriteString("\n")
			}
			postIdx++
		}
		b.WriteString("\n")
	}

	body := b.String()
	return ContentBox(lipgloss.NewStyle().Padding(0, 1).Render(body), m.width, m.dims)
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

func (m Model) viewPage(slug, title, empty string) string {
	theme := m.theme
	pageTitle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.PageTitle).
		Render(title)
	p, ok := m.pages[slug]
	if !ok {
		note := lipgloss.NewStyle().Foreground(theme.Muted).Render(empty)
		return ContentBox(pageTitle+"\n\n"+note, m.width, m.dims)
	}
	body := p.Content
	if r, err := m.renderMD(body); err == nil {
		body = r
	}
	return ContentBox(pageTitle+"\n\n"+body, m.width, m.dims)
}

func (m Model) viewReading() string {
	if m.currentPost == nil {
		return ""
	}
	theme := m.theme

	contentW := ContentMeasureWidth(m.width, m.dims)
	if contentW < m.cfg.Layout.MinReadingWidth {
		contentW = m.cfg.Layout.MinReadingWidth
	}

	var b strings.Builder
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.PageTitle).
		Render(m.cfg.Reading.TitlePrefix + m.currentPost.Title)
	b.WriteString(title)
	b.WriteString("\n\n")

	tocW := 0
	leftBlock := ""
	if len(m.toc) > 0 {
		tocW = m.homeColWidth()
		leftBlock = m.renderTOC(tocW)
	}

	articleW := contentW - tocW - 2
	if articleW < m.cfg.Reading.ArticleMinWidth {
		articleW = m.cfg.Reading.ArticleMinWidth
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

	pad := ContentMeasurePad(m.width, m.dims)
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

func (m Model) renderTOC(width int) string {
	theme := m.theme
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Primary).
		Padding(0, 1).
		Render(m.cfg.Reading.TOCTitle)
	var b strings.Builder
	b.WriteString(header)
	b.WriteString("\n\n")
	for i, h := range m.toc {
		style := lipgloss.NewStyle().Foreground(theme.Muted)
		marker := "  "
		if i == m.tocIdx {
			marker = "▸ "
			style = lipgloss.NewStyle().Foreground(theme.Highlight).Bold(true)
		}
		indent := strings.Repeat("  ", h.Level-2)
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

	pageTitle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.PageTitle).
		Render(m.cfg.Pages.Messages.Title)

	header := pageTitle
	if m.msgErr != "" {
		errLine := lipgloss.NewStyle().Foreground(theme.Error).Bold(true).Render("⚠ " + m.msgErr)
		header += "\n\n" + errLine
	}

	var form strings.Builder
	if m.state == stateComposing || m.state == stateSubmitting {
		form.WriteString(lipgloss.NewStyle().Bold(true).Foreground(theme.Secondary).Render(m.cfg.Pages.Messages.ComposeTitle) + "\n\n")
		labels := m.composeLabels()
		form.WriteString(m.compFieldView(0, labels[0], &m.compNick) + "\n")
		form.WriteString(m.compFieldView(1, labels[1], &m.compMail) + "\n")
		form.WriteString(m.compFieldView(2, labels[2], &m.compLink) + "\n")
		form.WriteString(m.compFieldView(3, labels[3], &m.compBody) + "\n")
		if m.state == stateSubmitting {
			form.WriteString(m.spinner.View() + m.cfg.Pages.Messages.Sending)
		} else {
			form.WriteString(lipgloss.NewStyle().Foreground(theme.Muted).Render("\n" + m.cfg.Pages.Messages.ComposeHint))
		}
	} else {
		form.WriteString(lipgloss.NewStyle().Foreground(theme.Secondary).Bold(true).Render(m.cfg.Pages.Messages.ComposePrompt) + "\n")
	}

	body := header + "\n\n" + m.msgList.View() + "\n" + form.String()
	return ContentBox(lipgloss.NewStyle().Padding(0, 1).Render(body), m.width, m.dims)
}

func (m Model) composeLabels() [4]string {
	cfg := m.cfg.Pages.Messages.Form
	if len(cfg) < 4 {
		def := config.Default().Pages.Messages.Form
		cfg = def
	}
	var out [4]string
	for i := 0; i < 4; i++ {
		out[i] = cfg[i].Label
	}
	return out
}

func (m *Model) refreshMessagesContent() {
	theme := m.theme
	var listContent strings.Builder
	if m.waline == nil {
		listContent.WriteString(lipgloss.NewStyle().Foreground(theme.Muted).Render(m.cfg.Pages.Messages.NotConfigured) + "\n")
	} else if !m.msgLoaded && !m.msgLoading {
		listContent.WriteString(lipgloss.NewStyle().Foreground(theme.Muted).Render(m.cfg.Pages.Messages.Loading) + "\n")
	} else if m.msgLoading {
		listContent.WriteString(m.spinner.View() + m.cfg.Pages.Messages.LoadingAlt + "\n")
	} else {
		if len(m.messages) == 0 {
			listContent.WriteString(lipgloss.NewStyle().Foreground(theme.Muted).Render(m.cfg.Pages.Messages.Empty) + "\n")
		} else {
			for _, msg := range m.messages {
				listContent.WriteString(m.renderComment(msg, 0))
				for _, ch := range msg.Children {
					listContent.WriteString(m.renderComment(ch, 1))
				}
			}
		}
	}
	m.msgList.SetContent(listContent.String())
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
	if c.Type == m.cfg.Waline.AdminType {
		nick = lipgloss.NewStyle().Foreground(theme.Accent).Bold(true).Render(m.cfg.Post.AdminMarker + nick)
	} else {
		nick = lipgloss.NewStyle().Bold(true).Foreground(theme.Secondary).Render(nick)
	}
	t := c.InsertedAt.Local().Format(m.cfg.Post.CommentTimeFormat)
	head := indent + nick + "  " + lipgloss.NewStyle().Foreground(theme.Muted).Render(t)
	body := indent + "  " + blog.StripHTML(c.Comment)
	divider := indent + "  " + lipgloss.NewStyle().Foreground(theme.Border).Render(strings.Repeat(m.cfg.Post.CommentDivider, m.cfg.Post.CommentDividerLen))
	return head + "\n" + body + "\n" + divider + "\n"
}

func (m Model) homeColWidth() int {
	cw := PageWidth(m.width, m.dims)
	totalContent := cw - m.cfg.Layout.HomeGap
	if totalContent < 8 {
		totalContent = 8
	}
	left := totalContent / m.cfg.Layout.HomeLeftRatio
	if left < 4 {
		left = 4
	}
	return left
}

func (m Model) rightColWidth() int {
	cw := PageWidth(m.width, m.dims)
	totalContent := cw - m.cfg.Layout.HomeGap
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

func (m Model) viewHomeDebugBodyH() int  { return m.bodyH }
func (m Model) viewHomeDebugHomeW() int  { return m.homeColWidth() }
func (m Model) viewHomeDebugRightW() int { return m.rightColWidth() }

func (m Model) StateDebug() int { return int(m.state) }
func (m Model) PageDebug() int  { return m.page }
