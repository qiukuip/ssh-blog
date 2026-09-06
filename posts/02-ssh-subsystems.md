---
title: SSH is more than a shell
date: 2024-03-20
category: tech
tags: [ssh, networking, golang]
author: longkun
summary: SSH is a generic secure channel — and you can run any program over it.
---

# SSH is more than a shell

Most people think of SSH as a remote terminal. But under the hood, SSH is just
an authenticated, encrypted transport. Once a client and server agree on
keys, you can run arbitrary subsystems on top of it.

## Subsystems

The SSH protocol defines a `subsystem` request. Servers can register handlers
for subsystem names such as `sftp`. The client then negotiates that subsystem
instead of asking for a shell.

This is how tools like `scp` and `rsync` over SSH work — they request a
specific subsystem or command.

## Wish + Bubble Tea

The [`charmbracelet/wish`](https://github.com/charmbracelet/wish) library lets
you treat a Bubble Tea program as an SSH subsystem. When a user connects, your
TUI runs as their process, with stdin/stdout flowing over the encrypted
channel.

```go
wish.NewServer(
    wish.WithAddress(":2222"),
    wish.WithHostKeyPEM(hostKey),
    wish.WithMiddleware(
        bubbletea.Middleware(teaHandler),
        logging.Middleware(),
    ),
)
```

That's it. No terminal emulation dance, no screen scraping — the client just
runs your TUI.

## Other ideas

- Personal dashboards (calendar, todos, weather)
- Chat over SSH
- Multiplayer games
- A `git`-style protocol for syncing files

Anything you can express as a streaming byte protocol over stdin/stdout can
ride SSH.
