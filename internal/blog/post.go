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

// MetaConfig holds the emoji/format tokens used when rendering post headers.
// All fields have safe zero-value defaults so callers can leave it empty.
type MetaConfig struct {
	DateFormat        string
	UnknownDate       string
	DateEmoji         string
	AuthorEmoji       string
	CategoryEmoji     string
	TagEmoji          string
}

// Header returns the Markdown header for a post (title + meta line).
func (p Post) Header(meta MetaConfig) string {
	dateFmt := meta.DateFormat
	if dateFmt == "" {
		dateFmt = "2006-01-02"
	}
	unknown := meta.UnknownDate
	if unknown == "" {
		unknown = "unknown"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "**%s**\n\n", p.Title)
	if p.Summary != "" {
		fmt.Fprintf(&b, "%s\n\n", p.Summary)
	}

	var dateStr string
	if p.Date.IsZero() {
		dateStr = unknown
	} else {
		dateStr = p.Date.Format(dateFmt)
	}
	if meta.DateEmoji != "" {
		fmt.Fprintf(&b, "- %s %s", meta.DateEmoji, dateStr)
	} else {
		fmt.Fprintf(&b, "- %s", dateStr)
	}
	if p.Author != "" {
		emoji := meta.AuthorEmoji
		if emoji == "" {
			emoji = "✍️"
		}
		fmt.Fprintf(&b, " · %s %s", emoji, p.Author)
	}
	if p.Category != "" {
		emoji := meta.CategoryEmoji
		if emoji == "" {
			emoji = "📂"
		}
		fmt.Fprintf(&b, " · %s %s", emoji, p.Category)
	}
	if len(p.Tags) > 0 {
		emoji := meta.TagEmoji
		if emoji == "" {
			emoji = "🏷️"
		}
		fmt.Fprintf(&b, " · %s %s", emoji, p.TagsString())
	}
	b.WriteString("\n")
	return b.String()
}
