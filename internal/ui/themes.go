package ui

import "github.com/charmbracelet/lipgloss"

type Theme struct {
	Name        string
	Primary     lipgloss.Color
	Secondary   lipgloss.Color
	Accent      lipgloss.Color
	Border      lipgloss.Color
	Muted       lipgloss.Color
	Title       lipgloss.Color
	Highlight   lipgloss.Color
	Success     lipgloss.Color
	Error       lipgloss.Color
	NavActive   lipgloss.Color
	NavInactive lipgloss.Color
	PageTitle   lipgloss.Color
}

var themes = []Theme{
	{
		Name:        "default",
		Primary:     lipgloss.Color("#A076FF"),
		Secondary:   lipgloss.Color("#36D399"),
		Accent:      lipgloss.Color("#FF6BAD"),
		Border:      lipgloss.Color("#5C5C70"),
		Muted:       lipgloss.Color("#9A9AB0"),
		Title:       lipgloss.Color("#FAFAFA"),
		Highlight:   lipgloss.Color("#FF6BAD"),
		Success:     lipgloss.Color("#36D399"),
		Error:       lipgloss.Color("#FF6B6B"),
		NavActive:   lipgloss.Color("#FFFFFF"),
		NavInactive: lipgloss.Color("#9A9AB0"),
		PageTitle:   lipgloss.Color("#A076FF"),
	},
	{
		Name:        "solarized",
		Primary:     lipgloss.Color("#3FA6FF"),
		Secondary:   lipgloss.Color("#A3D900"),
		Accent:      lipgloss.Color("#FFAA00"),
		Border:      lipgloss.Color("#7A9BAD"),
		Muted:       lipgloss.Color("#A6B5BB"),
		Title:       lipgloss.Color("#FDF6E3"),
		Highlight:   lipgloss.Color("#FF7733"),
		Success:     lipgloss.Color("#A3D900"),
		Error:       lipgloss.Color("#FF5555"),
		NavActive:   lipgloss.Color("#FFFFFF"),
		NavInactive: lipgloss.Color("#A6B5BB"),
		PageTitle:   lipgloss.Color("#3FA6FF"),
	},
	{
		Name:        "dracula",
		Primary:     lipgloss.Color("#C792FF"),
		Secondary:   lipgloss.Color("#5DFA88"),
		Accent:      lipgloss.Color("#FF89D6"),
		Border:      lipgloss.Color("#7F8FBC"),
		Muted:       lipgloss.Color("#9098B8"),
		Title:       lipgloss.Color("#F8F8F2"),
		Highlight:   lipgloss.Color("#FF89D6"),
		Success:     lipgloss.Color("#5DFA88"),
		Error:       lipgloss.Color("#FF6E6E"),
		NavActive:   lipgloss.Color("#FFFFFF"),
		NavInactive: lipgloss.Color("#9098B8"),
		PageTitle:   lipgloss.Color("#C792FF"),
	},
	{
		Name:        "nord",
		Primary:     lipgloss.Color("#92D5FF"),
		Secondary:   lipgloss.Color("#B5DCA0"),
		Accent:      lipgloss.Color("#FFD580"),
		Border:      lipgloss.Color("#5D708C"),
		Muted:       lipgloss.Color("#8FA1BC"),
		Title:       lipgloss.Color("#ECEFF4"),
		Highlight:   lipgloss.Color("#FF7B95"),
		Success:     lipgloss.Color("#B5DCA0"),
		Error:       lipgloss.Color("#FF7B95"),
		NavActive:   lipgloss.Color("#FFFFFF"),
		NavInactive: lipgloss.Color("#8FA1BC"),
		PageTitle:   lipgloss.Color("#92D5FF"),
	},
	{
		Name:        "sakura",
		Primary:     lipgloss.Color("#F7A8C4"),
		Secondary:   lipgloss.Color("#FBC9D9"),
		Accent:      lipgloss.Color("#FFB59A"),
		Border:      lipgloss.Color("#B98AA6"),
		Muted:       lipgloss.Color("#A89BAA"),
		Title:       lipgloss.Color("#FFF0F5"),
		Highlight:   lipgloss.Color("#FF6B9D"),
		Success:     lipgloss.Color("#A8D5BA"),
		Error:       lipgloss.Color("#FF8A95"),
		NavActive:   lipgloss.Color("#FFFFFF"),
		NavInactive: lipgloss.Color("#C9A8B8"),
		PageTitle:   lipgloss.Color("#F58FB1"),
	},
}

func ThemeByIndex(i int) Theme {
	n := len(themes)
	if n == 0 {
		return Theme{}
	}
	return themes[(i%n+n)%n]
}

func ThemeNames() []string {
	out := make([]string, len(themes))
	for i, t := range themes {
		out[i] = t.Name
	}
	return out
}
