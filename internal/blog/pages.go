package blog

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Page struct {
	Slug    string
	Title   string
	Content string
}

func LoadPages(dir string) (map[string]Page, error) {
	out := make(map[string]Page)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return out, nil
		}
		return nil, fmt.Errorf("read pages dir: %w", err)
	}
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
		out[slug] = Page{Slug: slug, Title: p.Title, Content: p.Content}
	}
	return out, nil
}

// GroupPostsByYear returns posts grouped by year (descending), posts within
// each year sorted newest first.
func GroupPostsByYear(posts []Post) map[int][]Post {
	m := make(map[int][]Post)
	for _, p := range posts {
		y := p.Date.Year()
		if p.Date.IsZero() {
			y = 0
		}
		m[y] = append(m[y], p)
	}
	for y := range m {
		ps := m[y]
		sort.Slice(ps, func(i, j int) bool { return ps[i].Date.After(ps[j].Date) })
		m[y] = ps
	}
	return m
}

func Keys(m map[int][]Post) []int {
	out := make([]int, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] > out[j] })
	return out
}
