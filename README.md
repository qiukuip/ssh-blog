# ssh-blog

A terminal-first blog served over SSH. Connect with any SSH client and read
articles rendered as Markdown inside a TUI. No browser, no accounts, no
JavaScript.

Modeled on the visual layout of modern personal blogs: top navigation, a
centered content column with side whitespace, and a three-part footer.

Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea),
[Glamour](https://github.com/charmbracelet/glamour) and
[Wish](https://github.com/charmbracelet/wish).

## Features

- Top navigation bar (首页 / 归档 / 友链 / 留言 / 关于) with active-page highlight
- Built-in themes; cycle live with `T` — fully redefinable via `config.yaml`
- Home: two-column layout — intro on the left, latest posts on the right
- Archive: all posts grouped by year, sorted newest first
- Reader: article on the right, **TOC sidebar (h1/h2/h3)** on the left, with
  per-heading navigation (`n` / `p`)
- Message board backed by [Waline](https://waline.js.org/) — read existing
  comments and post new ones
- Static pages (about / friends) loaded from `pages/*.md`
- Centered content with side margins; auto-clamps width (configurable)
- Footer: author · copyright · beian · 开往 · 十年之约 · compact nav
- Full-text search (`/`) across title, summary, tags and body
- Single static binary, no runtime dependencies
- **Every customizable value lives in `config.yaml`** — nav labels, page
  titles, hints, themes, keyboard bindings, layout, footer text/beian links,
  Waline endpoint/timeout, post-meta emoji, glamour style, and more

## Quick start

```bash
go build -o ssh-blog .

# default: 0.0.0.0:2222, posts=./posts, pages=./pages
./ssh-blog

# or with custom flags (flags override config)
./ssh-blog -host 0.0.0.0 -port 2222 \
           -posts ./posts \
           -pages ./pages \
           -waline https://waline-comments.happy365.day \
           -key ./host_key

# Visit
ssh -p 2222 localhost
```

## Configuration

Every customizable value lives in a single YAML file. Copy
[`config.example.yaml`](./config.example.yaml) to `./config.yaml` (or
`~/.config/ssh-blog/config.yaml`) and edit what you need — anything you omit
falls back to the built-in default.

```yaml
site:
  name: "qiukui-note"
  author: "qiukui-note"
  copyright: "© 2018-2026 qiukui-note"
  beian: "萌ICP备 20249900 号"

nav:
  - { key: "1", label: "首页",  slug: "" }
  - { key: "2", label: "归档",  slug: "" }
  - { key: "3", label: "友链",  slug: "friends" }
  - { key: "4", label: "留言",  slug: "" }
  - { key: "5", label: "关于",  slug: "about" }

waline:
  base_url: "https://waline-comments.happy365.day"
  messages_path: "messages/index.html"
  page_size: 50
  timeout: 10s

themes:
  - name: "default"
    primary: "#A076FF"
    secondary: "#36D399"
    # ... (see config.example.yaml for the full schema)

keys:
  quit: ["q", "ctrl+c"]
  theme_cycle: ["T", "t"]
  toc_next: "n"
  # ... (every binding is overridable)

layout:
  content_measure: 90
  outer_padding: 4
  home_left_ratio: 3
```

Search order (first file that exists wins):

1. `-config <path>` flag
2. `./config.yaml` / `./config.yml`
3. `~/.config/ssh-blog/config.yaml` / `.yml`
4. Built-in defaults

**Precedence at runtime**: `CLI flag > config.yaml > built-in default`. So
`-port 8080` always wins, and an unset flag picks up the value from
`server.default_port` (which itself falls back to `2222`).

See [`config.example.yaml`](./config.example.yaml) for the complete,
commented schema — every key is optional and documented.

### Flags

| Flag       | Default                                            | Description                                |
| ---------- | -------------------------------------------------- | ------------------------------------------ |
| `-config`  | _(auto-discover)_                                  | Path to config.yaml                        |
| `-host`    | `server.default_host` (`0.0.0.0`)                  | Address to listen on                       |
| `-port`    | `server.default_port` (`2222`)                     | TCP port                                   |
| `-posts`   | `server.default_posts_dir` (`posts`)               | Directory of Markdown blog posts           |
| `-pages`   | `server.default_pages_dir` (`pages`)               | Directory of Markdown static pages         |
| `-waline`  | `waline.base_url`                                  | Waline API base URL; empty disables posting|
| `-key`     | `server.host_key_default_path` (`~/.ssh/ssh-blog_hostkey`) | Path to host key (PEM); generated if absent |
| `-version` | false                                              | Print version (`site.version`) and exit    |

## Keybindings

| Key            | Action                                                |
| -------------- | ----------------------------------------------------- |
| `1` `2` `3` `4` `5` | Jump to 首页 / 归档 / 友链 / 留言 / 关于       |
| `T` (or `t`)   | Cycle theme (default → solarized → dracula → nord)    |
| `↑` `↓` `j` `k` | Move highlight in the current list                   |
| `enter`        | Open highlighted post / start composing a message     |
| `/`            | Search across title, summary, tags and body           |
| `n` / `p`      | Jump to next / previous TOC entry (in reader)         |
| `tab` / `shift+tab` | Switch between fields (when composing)          |
| `esc`          | Back to home / cancel current mode                    |
| `q` / `ctrl+c` | Quit                                                  |

## Page layout

```
┌──────────────────────────────────────────────────────────────────────────┐
│ ssh-blog      1 首页  2 归档  3 友链  4 留言  5 关于        [T] default │
├──────────────────────────────────────────────────────────────────────────┤
│ ┌──────────────────────────────┐  ┌────────────────────────────────────┐  │
│ │ # ssh-blog                   │  │ ▸ Welcome to ssh-blog              │  │
│ │                              │  │   2024-05-01 · meta · intro        │  │
│ │ 靠兴趣发电的个人小站…         │  │   Building a TUI with Bubble Tea    │  │
│ │                              │  │   2024-04-12 · tech · go,bubbletea │  │
│ │ 在算法推荐的时代…             │  │   ...                              │  │
│ └──────────────────────────────┘  └────────────────────────────────────┘  │
├──────────────────────────────────────────────────────────────────────────┤
│ longkun · © 2018-2026 ssh-blog     萌ICP备…号 · 开往 · 十年之约  首页   │
└──────────────────────────────────────────────────────────────────────────┘
```

## Authoring

### Posts (`posts/*.md`)

Drop a Markdown file in `posts/`. Optional frontmatter:

```markdown
---
title: My new post
date: 2024-05-01
category: tech
tags: [go, ssh, tui]
author: you
summary: A short blurb shown in the list view.
---

# My new post

Content goes here. Code blocks, tables, links, images — Glamour renders all
of it.
```

- Files sorted by `date` descending (newest first)
- Missing fields fall back to sensible defaults (filename becomes the title if
  `title` is omitted)

### Pages (`pages/*.md`)

Static pages (`about.md`, `friends.md`, …) are rendered with the same
Markdown pipeline. Filename (without extension) becomes the slug, so
`pages/about.md` powers the **关于** nav item.

### Themes

Themes are defined in `config.yaml` under the `themes:` key — add as many as
you like, each entry with the same 12 color fields (see
[`config.example.yaml`](./config.example.yaml)). Press `T` in-app to cycle
through them. Colors follow
[lipgloss](https://github.com/charmbracelet/lipgloss) (hex strings or named
ANSI colors).

## Message board

The留言 board reads from and writes to a [Waline](https://waline.js.org/)
endpoint. The default URL is `https://waline-comments.happy365.day`,
matching `qiukui-note`'s comments.

- **Read**: pressing `4` (or navigating to 留言) fetches
  `GET {waline.base_url}/comment?path={waline.messages_path}&pageSize={waline.page_size}`
- **Post**: pressing `enter` on the留言 page opens a 4-field form
  (nick / mail / link / message). Fill it in with `tab` between fields and
  hit `enter` on the last field to submit.

Every Waline parameter (`base_url`, `messages_path`, `page_size`, `timeout`,
`cache_ttl`, `admin_type`) is configurable in `config.yaml`. The Waline
adapter lives in `internal/blog/waline.go`.

## Testing

```bash
go test ./...
```

Unit tests drive the Bubble Tea model directly (no SSH needed) and cover
home rendering, navigation, theme cycling, reading + TOC, message page,
search, and footer links.

## License

MIT
