package blog

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

var frontmatterRe = regexp.MustCompile(`(?s)^---\s*\n(.*?)\n---\s*\n?(.*)$`)

func Parse(raw []byte, slug string) (Post, error) {
	m := frontmatterRe.FindSubmatch(raw)
	if m == nil {
		return Post{Slug: slug, Title: slug, Content: string(raw)}, nil
	}
	var p Post
	if err := yaml.Unmarshal(m[1], &p); err != nil {
		return p, fmt.Errorf("parse frontmatter: %w", err)
	}
	p.Content = strings.TrimLeft(string(m[2]), "\n")
	p.Slug = slug
	if p.Title == "" {
		p.Title = slug
	}
	return p, nil
}

func LoadDir(dir string) ([]Post, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read posts dir: %w", err)
	}
	var posts []Post
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		ext := strings.ToLower(filepath.Ext(name))
		if ext != ".md" && ext != ".markdown" {
			continue
		}
		slug := strings.TrimSuffix(name, ext)
		raw, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", name, err)
		}
		p, err := Parse(raw, slug)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", name, err)
		}
		posts = append(posts, p)
	}
	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Date.After(posts[j].Date)
	})
	return posts, nil
}

func GroupByCategory(posts []Post) map[string][]Post {
	m := make(map[string][]Post)
	for _, p := range posts {
		key := p.Category
		if key == "" {
			key = "uncategorized"
		}
		m[key] = append(m[key], p)
	}
	return m
}

func GroupByTag(posts []Post) map[string][]Post {
	m := make(map[string][]Post)
	for _, p := range posts {
		for _, t := range p.Tags {
			m[t] = append(m[t], p)
		}
	}
	return m
}

func Search(posts []Post, query string) []Post {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return posts
	}
	var out []Post
	for _, p := range posts {
		hay := strings.ToLower(strings.Join([]string{
			p.Title, p.Summary, p.Category, p.TagsString(), p.Content,
		}, "\n"))
		if strings.Contains(hay, q) {
			out = append(out, p)
		}
	}
	return out
}

