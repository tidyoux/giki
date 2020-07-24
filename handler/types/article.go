package types

import (
	"html/template"
	"os/exec"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/tidyoux/goutils"
)

var (
	titleReg = regexp.MustCompile(`#\s+(.*?)\s*\n`)
	tagReg   = regexp.MustCompile(`(?m)^TAG:(.+?)$`)
	langReg  = regexp.MustCompile("(?m)^```([a-z]+)")
)

type Article struct {
	ID        string
	Title     string
	Tags      []string
	Content   string
	UpdatedAt int64
}

func NewArticle(id, content string, updatedAt int64) *Article {
	a := &Article{
		ID:        id,
		Content:   content,
		UpdatedAt: updatedAt,
	}
	a.updateTitleTags()
	return a
}

func (a *Article) updateTitleTags() {
	match := titleReg.FindStringSubmatch(a.Content)
	if len(match) == 2 {
		a.Title = match[1]
	}

	match = tagReg.FindStringSubmatch(a.Content)
	if len(match) == 2 {
		tags := strings.Split(match[1], " ")
		for _, t := range tags {
			if t := strings.TrimSpace(t); len(t) > 0 {
				a.Tags = append(a.Tags, strings.ToUpper(t))
			}
		}
	}
}

func (a *Article) HasTag(tag string) bool {
	if len(tag) == 0 {
		return true
	}

	for _, t := range a.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

func (a *Article) RenderContent() template.HTML {
	result, _ := goutils.ExeCmd("multimarkdown", nil, func(c *exec.Cmd) {
		content := langReg.ReplaceAllString(a.Content, "```prettyprint lang-$1")
		c.Stdin = strings.NewReader(content)
	})
	return template.HTML(result)
}

func (a *Article) FormatUpdatedAt(layout string) string {
	return time.Unix(a.UpdatedAt, 0).Format(layout)
}

type Wiki struct {
	User      string
	Articles  []*Article
	SelectTag string

	Tags []string
}

func NewWiki(user string, articles []*Article, selectTag string) *Wiki {
	w := &Wiki{
		User:      user,
		Articles:  articles,
		SelectTag: selectTag,
	}
	w.updateTags()
	w.sort()
	return w
}

func (w *Wiki) updateTags() {
	tags := make(map[string]struct{})
	for _, a := range w.Articles {
		for _, t := range a.Tags {
			tags[t] = struct{}{}
		}
	}

	w.Tags = make([]string, 0, len(tags))
	for t := range tags {
		w.Tags = append(w.Tags, t)
	}
	sort.Strings(w.Tags)
}

func (w *Wiki) sort() {
	sort.Slice(w.Articles, func(i, j int) bool {
		return w.Articles[i].UpdatedAt > w.Articles[j].UpdatedAt
	})
}
