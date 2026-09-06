---
title: Building a TUI with Bubble Tea
date: 2024-04-12
category: tech
tags: [go, bubbletea, tui]
author: longkun
summary: Notes on the Elm-style architecture that powers Bubble Tea.
---

# Building a TUI with Bubble Tea

Bubble Tea is a Go framework for terminal UIs based on [The Elm Architecture](https://guide.elm-lang.org/architecture/).
You implement a single `Model` that holds state, react to messages, and return
a string view.

## The core loop

```go
type Model struct { count int }

func (m Model) Init() tea.Cmd                         { return nil }
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) { /* ... */ }
func (m Model) View() string                          { /* ... */ }
```

1. `Init` returns a startup command (optional).
2. `Update` is called for every event — keystrokes, window resizes, ticks.
3. `View` renders the entire screen as a string.

## Why it works so well

- **Pure functions** make state transitions easy to reason about.
- **Commands** keep side effects (I/O, time, network) out of `Update`.
- **Composability** through small reusable components (`viewport`, `list`,
  `textinput`).

## A minimal counter

```go
type model struct{ count int }

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "q", "ctrl+c":
            return m, tea.Quit
        case "+":
            m.count++
        }
    }
    return m, nil
}

func (m model) View() string {
    return fmt.Sprintf("count: %d\npress + to increment, q to quit", m.count)
}
```

Run with `tea.NewProgram(m).Run()` and you have a working TUI in about 20 lines.
