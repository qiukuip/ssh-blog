package blog

import (
	"fmt"
	"strings"
	"time"
)

type Post struct {
	Slug     string    `yaml:"-"`
	Title    string    `yaml:"title"`
	Date     time.Time `yaml:"date"`
	Category string    `yaml:"category"`
	Tags     []string  `yaml:"tags"`
	Author   string    `yaml:"author"`
	Summary  string    `yaml:"summary"`
	Content  string    `yaml:"-"`
}

func (p Post) TagsString() string {
	return strings.Join(p.Tags, ", ")
}

func (p Post) DisplayDate() string {
	if p.Date.IsZero() {
		return "unknown"
	}
	return p.Date.Format("2006-01-02")
}

func (p Post) Header() string {
	var b strings.Builder
	fmt.Fprintf(&b, "**%s**\n\n", p.Title)
	if p.Summary != "" {
		fmt.Fprintf(&b, "%s\n\n", p.Summary)
	}
	fmt.Fprintf(&b, "- 📅 %s", p.DisplayDate())
	if p.Author != "" {
		fmt.Fprintf(&b, " · ✍️ %s", p.Author)
	}
	if p.Category != "" {
		fmt.Fprintf(&b, " · 📂 %s", p.Category)
	}
	if len(p.Tags) > 0 {
		fmt.Fprintf(&b, " · 🏷️ %s", p.TagsString())
	}
	b.WriteString("\n")
	return b.String()
}
