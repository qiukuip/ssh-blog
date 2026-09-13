package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Site    SiteConfig    `yaml:"site"`
	Nav     []NavConfig   `yaml:"nav"`
	Home    HomeConfig    `yaml:"home"`
	Footer  FooterConfig  `yaml:"footer"`
	Pages   PagesConfig   `yaml:"pages"`
	Reading ReadingConfig `yaml:"reading"`
	Search  SearchConfig  `yaml:"search"`
	Hints   HintsConfig   `yaml:"hints"`
	Post    PostConfig    `yaml:"post"`
	Waline  WalineConfig  `yaml:"waline"`
	Keys    KeysConfig    `yaml:"keys"`
	Layout  LayoutConfig  `yaml:"layout"`
	Glamour GlamourConfig `yaml:"glamour"`
	Themes  []ThemeConfig `yaml:"themes"`
	Server  ServerConfig  `yaml:"server"`
}

type SiteConfig struct {
	Name      string `yaml:"name"`
	Author    string `yaml:"author"`
	Copyright string `yaml:"copyright"`
	Beian     string `yaml:"beian"`
	Version   string `yaml:"version"`
	Tagline   string `yaml:"tagline"`
}

type NavConfig struct {
	Key   string `yaml:"key"`
	Label string `yaml:"label"`
	Slug  string `yaml:"slug"`
}

type HomeConfig struct {
	Title       string `yaml:"title"`
	RecentTitle string `yaml:"recent_title"`
	Intro       string `yaml:"intro"`
	RecentCount int    `yaml:"recent_count"`
}

type FooterConfig struct {
	Author        string `yaml:"author"`
	Copyright     string `yaml:"copyright"`
	Beian         string `yaml:"beian"`
	KaiwangURL    string `yaml:"kaiwang_url"`
	KaiwangLabel  string `yaml:"kaiwang_label"`
	ShinianURL    string `yaml:"shinian_url"`
	ShinianLabel  string `yaml:"shinian_label"`
	Separator     string `yaml:"separator"`
	CenterJoiner  string `yaml:"center_joiner"`
}

type PagesConfig struct {
	Home     PageEntry      `yaml:"home"`
	Archive  PageEntry      `yaml:"archive"`
	Friends  StaticPageConf `yaml:"friends"`
	About    StaticPageConf `yaml:"about"`
	Messages MessagesConfig `yaml:"messages"`
}

type PageEntry struct {
	Title string `yaml:"title"`
}

type StaticPageConf struct {
	PageEntry `yaml:",inline"`
	Slug      string `yaml:"slug"`
	Empty     string `yaml:"empty"`
}

type MessagesConfig struct {
	PageEntry         `yaml:",inline"`
	ComposeTitle      string        `yaml:"compose_title"`
	ComposePrompt     string        `yaml:"compose_prompt"`
	Sending           string        `yaml:"sending"`
	Empty             string        `yaml:"empty"`
	Loading           string        `yaml:"loading"`
	LoadingAlt        string        `yaml:"loading_alt"`
	NotConfigured     string        `yaml:"not_configured"`
	ErrorNotConfigured string       `yaml:"error_not_configured"`
	ValidationEmpty   string        `yaml:"validation_empty"`
	ComposeHint       string        `yaml:"compose_hint"`
	Form              []FormFieldConfig `yaml:"form"`
}

type FormFieldConfig struct {
	Label       string `yaml:"label"`
	Placeholder string `yaml:"placeholder"`
	Required    bool   `yaml:"required"`
	Limit       int    `yaml:"limit"`
	Prompt      string `yaml:"prompt"`
}

type ReadingConfig struct {
	TitlePrefix    string `yaml:"title_prefix"`
	TOCTitle       string `yaml:"toc_title"`
	TOCMinWidth    int    `yaml:"toc_min_width"`
	TOCMaxWidth    int    `yaml:"toc_max_width"`
	TOCRatio       int    `yaml:"toc_ratio"`
	ArticleMinWidth int   `yaml:"article_min_width"`
}

type SearchConfig struct {
	Placeholder string `yaml:"placeholder"`
	Prompt      string `yaml:"prompt"`
	CharLimit   int    `yaml:"char_limit"`
}

type HintsConfig struct {
	Reading    string `yaml:"reading"`
	Search     string `yaml:"search"`
	Composing  string `yaml:"composing"`
	Submitting string `yaml:"submitting"`
	Home       string `yaml:"home"`
	Archive    string `yaml:"archive"`
	Messages   string `yaml:"messages"`
	Fallback   string `yaml:"fallback"`
}

type PostConfig struct {
	DateFormat        string `yaml:"date_format"`
	UnknownDate       string `yaml:"unknown_date"`
	CommentTimeFormat string `yaml:"comment_time_format"`
	DateEmoji         string `yaml:"date_emoji"`
	AuthorEmoji       string `yaml:"author_emoji"`
	CategoryEmoji     string `yaml:"category_emoji"`
	TagEmoji          string `yaml:"tag_emoji"`
	AdminMarker       string `yaml:"admin_marker"`
	CommentDivider    string `yaml:"comment_divider"`
	CommentDividerLen int    `yaml:"comment_divider_len"`
}

type WalineConfig struct {
	BaseURL      string        `yaml:"base_url"`
	MessagesPath string        `yaml:"messages_path"`
	PageSize     int           `yaml:"page_size"`
	Timeout      time.Duration `yaml:"timeout"`
	CacheTTL     time.Duration `yaml:"cache_ttl"`
	AdminType    string        `yaml:"admin_type"`
}

type KeysConfig struct {
	Quit               []string `yaml:"quit"`
	ThemeCycle         []string `yaml:"theme_cycle"`
	Search             string   `yaml:"search"`
	Enter              string   `yaml:"enter"`
	Esc                string   `yaml:"esc"`
	NavDigits          string   `yaml:"nav_digits"`
	ReadingBack        []string `yaml:"reading_back"`
	ScrollUp           []string `yaml:"scroll_up"`
	ScrollDown         []string `yaml:"scroll_down"`
	PageDown           []string `yaml:"page_down"`
	PageUp             []string `yaml:"page_up"`
	Top                []string `yaml:"top"`
	Bottom             []string `yaml:"bottom"`
	TOCNext            string   `yaml:"toc_next"`
	TOCPrev            string   `yaml:"toc_prev"`
	ComposeNextField   string   `yaml:"compose_next_field"`
	ComposePrevField   string   `yaml:"compose_prev_field"`
	ComposeSubmit      string   `yaml:"compose_submit"`
}

type LayoutConfig struct {
	ContentMeasure    int `yaml:"content_measure"`
	MinContentWidth   int `yaml:"min_content_width"`
	MaxContentWidth   int `yaml:"max_content_width"`
	OuterPadding      int `yaml:"outer_padding"`
	MinTerminalWidth  int `yaml:"min_terminal_width"`
	NarrowNavThresh   int `yaml:"narrow_nav_threshold"`
	ChromeHeight      int `yaml:"chrome_height"`
	MinBodyHeight     int `yaml:"min_body_height"`
	MinListHeight     int `yaml:"min_list_height"`
	MinViewportWidth  int `yaml:"min_viewport_width"`
	MinViewportHeight int `yaml:"min_viewport_height"`
	MinReadingWidth   int `yaml:"min_reading_width"`
	HomeGap           int `yaml:"home_gap"`
	HomeLeftRatio     int `yaml:"home_left_ratio"`
	MessageFormHeight int `yaml:"message_form_height"`
	Gap               int `yaml:"gap"`
}

type GlamourConfig struct {
	Style            string `yaml:"style"`
	WordWrapOffset   int    `yaml:"word_wrap_offset"`
	MinRendererWidth int    `yaml:"min_renderer_width"`
	EnableEmoji      bool   `yaml:"enable_emoji"`
}

type ThemeConfig struct {
	Name        string `yaml:"name"`
	Primary     string `yaml:"primary"`
	Secondary   string `yaml:"secondary"`
	Accent      string `yaml:"accent"`
	Border      string `yaml:"border"`
	Muted       string `yaml:"muted"`
	Title       string `yaml:"title"`
	Highlight   string `yaml:"highlight"`
	Success     string `yaml:"success"`
	Error       string `yaml:"error"`
	NavActive   string `yaml:"nav_active"`
	NavInactive string `yaml:"nav_inactive"`
	PageTitle   string `yaml:"page_title"`
}

type ServerConfig struct {
	DefaultHost        string `yaml:"default_host"`
	DefaultPort        int    `yaml:"default_port"`
	DefaultPostsDir    string `yaml:"default_posts_dir"`
	DefaultPagesDir    string `yaml:"default_pages_dir"`
	ShutdownTimeout    time.Duration `yaml:"shutdown_timeout"`
	HostKeyAlgorithm   string `yaml:"host_key_algorithm"`
	HostKeyDefaultPath string `yaml:"host_key_default_path"`
	ListenBanner       string `yaml:"listen_banner"`
	ShutdownMessage    string `yaml:"shutdown_message"`
	InitPlaceholder    string `yaml:"init_placeholder"`
}

func Default() *Config {
	return &Config{
		Site: SiteConfig{
			Name:      "qiukui-note",
			Author:    "qiukui-note",
			Copyright: "© 2018-2026 qiukui-note",
			Beian:     "萌ICP备 20249900 号",
			Version:   "ssh-blog 0.2.0",
			Tagline:   "靠兴趣发电的个人小站。",
		},
		Nav: []NavConfig{
			{Key: "1", Label: "首页", Slug: ""},
			{Key: "2", Label: "归档", Slug: ""},
			{Key: "3", Label: "友链", Slug: "friends"},
			{Key: "4", Label: "留言", Slug: ""},
			{Key: "5", Label: "关于", Slug: "about"},
		},
		Home: HomeConfig{
			Title:       "🏠 关于本站",
			RecentTitle: "📰 最新文章",
			Intro:       "靠兴趣发电的个人小站。\n\n在算法推荐的时代，留一片安静的角落。\n\n试试 `1-5` 切换页面，\n或按 `enter` 阅读最新文章。",
			RecentCount: 10,
		},
		Footer: FooterConfig{
			Author:       "qiukui-note",
			Copyright:    "© 2018-2026 qiukui-note",
			Beian:        "萌ICP备 20249900 号",
			KaiwangURL:   "https://www.travellings.cn",
			KaiwangLabel: "开往",
			ShinianURL:   "https://www.foreverblog.cn",
			ShinianLabel: "十年之约",
			Separator:    " / ",
			CenterJoiner: " · ",
		},
		Pages: PagesConfig{
			Home:    PageEntry{Title: "🏠 关于本站"},
			Archive: PageEntry{Title: "📅 归档"},
			Friends: StaticPageConf{
				PageEntry: PageEntry{Title: "🔗 友链"},
				Slug:      "friends",
				Empty:     "暂无内容，请在 pages/friends.md 添加 Markdown",
			},
			About: StaticPageConf{
				PageEntry: PageEntry{Title: "👤 关于"},
				Slug:      "about",
				Empty:     "暂无内容，请在 pages/about.md 添加 Markdown",
			},
			Messages: MessagesConfig{
				PageEntry:          PageEntry{Title: "💬 留言板"},
				ComposeTitle:       "✎ 我要留言",
				ComposePrompt:      "[ enter ] 写留言",
				Sending:            " 正在发送…\n",
				Empty:              "还没有留言，留下第一条吧。",
				Loading:            "正在加载留言…",
				LoadingAlt:         "加载留言中…",
				NotConfigured:      "（留言后端未配置，仅可本地填写，不会发送）",
				ErrorNotConfigured: "留言后端未配置 (WALINE_URL 为空)",
				ValidationEmpty:    "昵称和留言不能为空",
				ComposeHint:        "tab 切换字段 · ↑/↓ 行 · ctrl+f/b 翻页 · enter 提交",
				Form: []FormFieldConfig{
					{Label: "昵称 *", Placeholder: "昵称 *", Required: true, Limit: 32, Prompt: "> "},
					{Label: "邮箱", Placeholder: "邮箱（可选）", Required: false, Limit: 64, Prompt: "> "},
					{Label: "主页", Placeholder: "主页（可选）", Required: false, Limit: 128, Prompt: "> "},
					{Label: "留言内容", Placeholder: "说点什么…（多行用 \\n）", Required: true, Limit: 400, Prompt: "> "},
				},
			},
		},
		Reading: ReadingConfig{
			TitlePrefix:     "📖 ",
			TOCTitle:        "📑 目录",
			TOCMinWidth:     18,
			TOCMaxWidth:     32,
			TOCRatio:        5,
			ArticleMinWidth: 30,
		},
		Search: SearchConfig{
			Placeholder: "搜索标题/标签/正文…",
			Prompt:      "/ ",
			CharLimit:   64,
		},
		Hints: HintsConfig{
			Reading:    "↑/↓ 行 · ctrl+f/b 翻页 · esc 返回",
			Search:     "输入搜索 · enter 确认 · esc",
			Composing:  "tab 切换 · ctrl+f/b 翻页 · enter 提交",
			Submitting: "提交中… · esc 关闭",
			Home:       "enter 阅读 · / 搜索 · 1-5 · T 主题 · q",
			Archive:    "↑/↓ 选择 · enter 阅读 · 1-5 · T 主题 · q",
			Messages:   "↑/↓ 滚动 · ctrl+f/b 翻页 · enter 留言 · esc",
			Fallback:   "esc · T 主题 · 1-5 · q",
		},
		Post: PostConfig{
			DateFormat:        "2006-01-02",
			UnknownDate:       "unknown",
			CommentTimeFormat: "2006-01-02 15:04",
			DateEmoji:         "📅",
			AuthorEmoji:       "✍️",
			CategoryEmoji:     "📂",
			TagEmoji:          "🏷️",
			AdminMarker:       "★ ",
			CommentDivider:    "─",
			CommentDividerLen: 40,
		},
		Waline: WalineConfig{
			BaseURL:      "https://waline-comments.happy365.day",
			MessagesPath: "messages/index.html",
			PageSize:     50,
			Timeout:      10 * time.Second,
			CacheTTL:     30 * time.Second,
			AdminType:    "administrator",
		},
		Keys: KeysConfig{
			Quit:             []string{"q", "ctrl+c"},
			ThemeCycle:       []string{"T", "t"},
			Search:           "/",
			Enter:            "enter",
			Esc:              "esc",
			NavDigits:        "12345",
			ReadingBack:      []string{"q", "esc", "b"},
			ScrollUp:         []string{"up", "k"},
			ScrollDown:       []string{"down", "j"},
			PageDown:         []string{"ctrl+f", "pgdown", "pagedown", "space"},
			PageUp:           []string{"ctrl+b", "pgup", "pageup"},
			Top:              []string{"home", "g"},
			Bottom:           []string{"end", "G"},
			TOCNext:          "n",
			TOCPrev:          "p",
			ComposeNextField: "tab",
			ComposePrevField: "shift+tab",
			ComposeSubmit:    "enter",
		},
		Layout: LayoutConfig{
			ContentMeasure:    90,
			MinContentWidth:   90,
			MaxContentWidth:   90,
			OuterPadding:      4,
			MinTerminalWidth:  20,
			NarrowNavThresh:   80,
			ChromeHeight:      4,
			MinBodyHeight:     6,
			MinListHeight:     4,
			MinViewportWidth:  30,
			MinViewportHeight: 3,
			MinReadingWidth:   60,
			HomeGap:           6,
			HomeLeftRatio:     3,
			MessageFormHeight: 7,
			Gap:               4,
		},
		Glamour: GlamourConfig{
			Style:            "notty",
			WordWrapOffset:   8,
			MinRendererWidth: 30,
			EnableEmoji:      true,
		},
		Themes: []ThemeConfig{
			{Name: "default", Primary: "#A076FF", Secondary: "#36D399", Accent: "#FF6BAD", Border: "#5C5C70", Muted: "#9A9AB0", Title: "#FAFAFA", Highlight: "#FF6BAD", Success: "#36D399", Error: "#FF6B6B", NavActive: "#FFFFFF", NavInactive: "#9A9AB0", PageTitle: "#A076FF"},
			{Name: "solarized", Primary: "#3FA6FF", Secondary: "#A3D900", Accent: "#FFAA00", Border: "#7A9BAD", Muted: "#A6B5BB", Title: "#FDF6E3", Highlight: "#FF7733", Success: "#A3D900", Error: "#FF5555", NavActive: "#FFFFFF", NavInactive: "#A6B5BB", PageTitle: "#3FA6FF"},
			{Name: "dracula", Primary: "#C792FF", Secondary: "#5DFA88", Accent: "#FF89D6", Border: "#7F8FBC", Muted: "#9098B8", Title: "#F8F8F2", Highlight: "#FF89D6", Success: "#5DFA88", Error: "#FF6E6E", NavActive: "#FFFFFF", NavInactive: "#9098B8", PageTitle: "#C792FF"},
			{Name: "nord", Primary: "#92D5FF", Secondary: "#B5DCA0", Accent: "#FFD580", Border: "#5D708C", Muted: "#8FA1BC", Title: "#ECEFF4", Highlight: "#FF7B95", Success: "#B5DCA0", Error: "#FF7B95", NavActive: "#FFFFFF", NavInactive: "#8FA1BC", PageTitle: "#92D5FF"},
			{Name: "sakura", Primary: "#F7A8C4", Secondary: "#FBC9D9", Accent: "#FFB59A", Border: "#B98AA6", Muted: "#A89BAA", Title: "#FFF0F5", Highlight: "#FF6B9D", Success: "#A8D5BA", Error: "#FF8A95", NavActive: "#FFFFFF", NavInactive: "#C9A8B8", PageTitle: "#F58FB1"},
		},
		Server: ServerConfig{
			DefaultHost:        "0.0.0.0",
			DefaultPort:        2222,
			DefaultPostsDir:    "posts",
			DefaultPagesDir:    "pages",
			ShutdownTimeout:    5 * time.Second,
			HostKeyAlgorithm:   "ed25519",
			HostKeyDefaultPath: "~/.ssh/ssh-blog_hostkey",
			ListenBanner:       "ssh-blog listening on %s:%d, posts=%s, pages=%s",
			ShutdownMessage:    "shutting down",
			InitPlaceholder:    "Initializing...",
		},
	}
}

// Load reads a YAML config from one of the candidate paths (first that exists),
// unmarshals it on top of Default(), and returns the merged result.
func Load(path string) (*Config, error) {
	cfg := Default()

	candidates := searchPaths(path)
	for _, p := range candidates {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("parse %s: %w", p, err)
		}
		return cfg, nil
	}
	return cfg, nil
}

func searchPaths(explicit string) []string {
	if explicit != "" {
		return []string{explicit}
	}
	home, _ := os.UserHomeDir()
	paths := []string{
		"config.yaml",
		"config.yml",
	}
	if home != "" {
		paths = append(paths,
			filepath.Join(home, ".config", "ssh-blog", "config.yaml"),
			filepath.Join(home, ".config", "ssh-blog", "config.yml"),
		)
	}
	return paths
}
