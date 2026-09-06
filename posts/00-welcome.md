---
title: Welcome to ssh-blog
date: 2024-05-01
category: meta
tags: [intro, hello]
author: longkun
summary: A short tour of this terminal-first blog and how to navigate it.
---

# Welcome

This blog lives entirely inside a terminal. To read it, all you need is an SSH
client:

```bash
ssh -p 2222 your.host
```

No accounts. No browsers. Just a TUI.

## Keybindings

- `/` — search across titles, summaries, tags and content
- `c` — browse by **category**
- `t` — browse by **tag**
- `a` — show all posts (clear current filter)
- `enter` — open the highlighted post
- `esc` / `q` — back / quit
- `↑` / `↓` — navigate lists
- `pgup` / `pgdn` — scroll inside an article

## Authoring

Every post is a plain Markdown file under `posts/`. Optional frontmatter
controls metadata:

```yaml
---
title: My Post
date: 2024-05-01
category: tech
tags: [go, ssh, tui]
author: me
summary: A short blurb shown in the list.
---
```

Posts are sorted by date, newest first. Reconnect to pick up changes — there is
no live-reload by design.

Enjoy.
