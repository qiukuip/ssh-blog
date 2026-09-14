package blog

import (
	"regexp"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

type Heading struct {
	Level  int
	Text   string
	Anchor string
}

// ExtractTOC walks the markdown AST and returns h2/h3 headings in order.
// h1 (the post title) is excluded since it duplicates the article heading.
// Content inside fenced code blocks is ignored (handled by the parser).
func ExtractTOC(md string) []Heading {
	src := []byte(md)
	mdParser := goldmark.New().Parser()
	root := mdParser.Parse(text.NewReader(src))

	var out []Heading
	_ = ast.Walk(root, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		h, ok := n.(*ast.Heading)
		if !ok {
			return ast.WalkContinue, nil
		}
		level := h.Level
		if level < 2 || level > 3 {
			return ast.WalkSkipChildren, nil
		}
		txt := headingText(h, src)
		txt = strings.TrimSpace(txt)
		if txt == "" {
			return ast.WalkSkipChildren, nil
		}
		out = append(out, Heading{
			Level:  level,
			Text:   txt,
			Anchor: slugify(txt),
		})
		return ast.WalkSkipChildren, nil
	})
	return out
}

func headingText(h *ast.Heading, src []byte) string {
	var b strings.Builder
	for c := h.FirstChild(); c != nil; c = c.NextSibling() {
		b.Write(collectText(c, src))
	}
	return b.String()
}

func collectText(n ast.Node, src []byte) []byte {
	var b []byte
	switch v := n.(type) {
	case *ast.Text:
		b = append(b, v.Segment.Value(src)...)
		// Handle soft line breaks within a heading.
		if v.HardLineBreak() || v.SoftLineBreak() {
			b = append(b, ' ')
		}
	case *ast.String:
		b = append(b, v.Value...)
	case *ast.AutoLink:
		b = append(b, v.URL(src)...)
	case *ast.RawHTML:
		// Skip inline HTML.
	default:
		// Recurse into children (e.g., emphasis, code spans).
		for c := n.FirstChild(); c != nil; c = c.NextSibling() {
			b = append(b, collectText(c, src)...)
		}
	}
	return b
}

var slugReplacer = regexp.MustCompile(`[^\p{L}\p{N}\-_]+`)

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = slugReplacer.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		return "section"
	}
	return s
}
