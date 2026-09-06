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
- Four built-in themes; cycle live with `T`
- Home: two-column layout — intro on the left, latest 10 posts on the right
- Archive: all posts grouped by year, sorted newest first
- Reader: article on the right, **TOC sidebar (h1/h2/h3)** on the left, with
  per-heading navigation (`n` / `p`)
- Message board backed by [Waline](https://waline.js.org/) — read existing
  comments and post new ones
- Static pages (about / friends) loaded from `pages/*.md`
- Centered content with side margins; auto-clamps width between 60 and 110
- Footer: author · copyright · beian · 开往 · 十年之约 · compact nav
- Full-text search (`/`) across title, summary, tags and body
- Single static binary, no runtime dependencies

## Quick start

```bash
go build -o ssh-blog .

# default: 0.0.0.0:2222, posts=./posts, pages=./pages
./ssh-blog

# or with custom flags
./ssh-blog -host 0.0.0.0 -port 2222 \
           -posts ./posts \
           -pages ./pages \
           -waline https://waline-comments.happy365.day \
           -key ./host_key

# Visit
ssh -p 2222 localhost
```

### Flags

| Flag       | Default                              | Description                                |
| ---------- | ------------------------------------ | ------------------------------------------ |
| `-host`    | `0.0.0.0`                            | Address to listen on                       |
| `-port`    | `2222`                               | TCP port                                   |
| `-posts`   | `posts`                              | Directory of Markdown blog posts           |
| `-pages`   | `pages`                              | Directory of Markdown static pages         |
| `-waline`  | `https://waline-comments.happy365.day` | Waline API base URL for the message board |
| `-key`     | _(in-memory)_                        | Path to host key (PEM); generated if absent|
| `-version` | false                                | Print version and exit                     |

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

Themes are defined in Go (`internal/ui/themes.go`). To add a new one, append
a `Theme{...}` struct to the `themes` slice — color names follow
[lipgloss](https://github.com/charmbracelet/lipgloss).

## Message board

The留言 board reads from and writes to a [Waline](https://waline.js.org/)
endpoint. The default URL is `https://waline-comments.happy365.day`,
matching `qiukui-note`'s comments.

- **Read**: pressing `4` (or navigating to 留言) fetches
  `GET /comment?path=messages/index.html&pageSize=50`
- **Post**: pressing `enter` on the留言 page opens a 4-field form
  (nick / mail / link / message). Fill it in with `tab` between fields and
  hit `enter` on the last field to submit.

The Waline adapter lives in `internal/blog/waline.go` and can be pointed at
any compatible endpoint via the `-waline` flag.

## Testing

```bash
go test ./...
```

Unit tests drive the Bubble Tea model directly (no SSH needed) and cover
home rendering, navigation, theme cycling, reading + TOC, message page,
search, and footer links.

## License

MIT
