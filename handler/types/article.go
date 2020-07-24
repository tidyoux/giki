package types

import (
	"html/template"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/tidyoux/goutils"
)

type Article struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	UpdatedAt int64  `json:"updated_at"`
}

func (a *Article) FormatUpdatedAt(layout string) string {
	return time.Unix(a.UpdatedAt, 0).Format(layout)
}

type Articles struct {
	User     string
	Articles []*Article
}

type ArticleDetail struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	UpdatedAt int64  `json:"updated_at"`
}

var langReg = regexp.MustCompile("(?m)^```([a-z]+)")

func (a *ArticleDetail) RenderContent() template.HTML {
	result, _ := goutils.ExeCmd("multimarkdown", nil, func(c *exec.Cmd) {
		content := langReg.ReplaceAllString(a.Content, "```prettyprint lang-$1")
		c.Stdin = strings.NewReader(content)
	})
	return template.HTML(result)
}
