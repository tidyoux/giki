package handler

import (
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/tidyoux/giki/handler/types"
	"github.com/tidyoux/goutils"
)

func ListArticle(c *gin.Context) {
	username, _, _ := c.Request.BasicAuth()
	articles, err := listArticle(username)
	if err != nil {
		c.String(http.StatusInternalServerError, "list article failed, %v", err)
		return
	}

	rootTmpl.ExecuteTemplate(c.Writer, "list", articles)
}

func CreateArticle(c *gin.Context) {
	var (
		username, _, _ = c.Request.BasicAuth()
		id             = strconv.Itoa(int(time.Now().UnixNano() / 1000000))
	)
	err := saveArticle(username, id, "# This is the title\n\n", true)
	if err != nil {
		c.String(http.StatusInternalServerError, "create article failed, %v", err)
		return
	}

	rootTmpl.ExecuteTemplate(c.Writer, "jump", "/article/"+id)
}

func ViewArticle(c *gin.Context) {
	var (
		username, _, _ = c.Request.BasicAuth()
		id             = c.Param("id")
	)
	article, err := readArticle(username, id, nil)
	if err != nil {
		c.String(http.StatusInternalServerError, "read article failed, %v", err)
		return
	}

	rootTmpl.ExecuteTemplate(c.Writer, "view", article)
}

func EditArticle(c *gin.Context) {
	var (
		username, _, _ = c.Request.BasicAuth()
		id             = c.Param("id")
	)
	article, err := readArticle(username, id, nil)
	if err != nil {
		c.String(http.StatusInternalServerError, "read article failed, %v", err)
		return
	}

	rootTmpl.ExecuteTemplate(c.Writer, "edit", article)
}

func UpdateArticle(c *gin.Context) {
	var (
		username, _, _ = c.Request.BasicAuth()
		id             = c.Param("id")
	)
	content := c.PostForm("content")
	err := saveArticle(username, id, content, false)
	if err != nil {
		c.String(http.StatusInternalServerError, "save article failed, %v", err)
		return
	}

	rootTmpl.ExecuteTemplate(c.Writer, "jump", "/article/"+id)
}

func DeleteArticle(c *gin.Context) {
	var (
		username, _, _ = c.Request.BasicAuth()
		id             = c.Param("id")
	)
	err := os.RemoveAll(articlePath(username, id))
	if err != nil {
		c.String(http.StatusInternalServerError, "delete article failed, %v", err)
		return
	}

	rootTmpl.ExecuteTemplate(c.Writer, "jump", "/article")
}

const (
	DBPath   = "./db"
	FileName = "content.md"
)

func userDBPath(username string) string {
	return filepath.Join(DBPath, username)
}

func articlePath(username, id string) string {
	return filepath.Join(userDBPath(username), id)
}

func articleFile(username, id string) string {
	return filepath.Join(articlePath(username, id), FileName)
}

func listArticle(username string) (*types.Articles, error) {
	var (
		dbPath   = userDBPath(username)
		articles = &types.Articles{
			User: username,
		}
	)
	err := filepath.Walk(dbPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if path == dbPath || !info.IsDir() {
			return nil
		}

		article, err := readArticle(username, info.Name(), info)
		if err != nil {
			return errors.Wrapf(err, "read article %s failed", info.Name())
		}

		articles.Articles = append(articles.Articles, &types.Article{
			ID:        article.ID,
			Title:     article.Title,
			UpdatedAt: article.UpdatedAt,
		})
		return filepath.SkipDir
	})
	if err != nil {
		return nil, errors.Wrap(err, "read db failed")
	}

	sort.Slice(articles.Articles, func(i, j int) bool {
		return articles.Articles[i].UpdatedAt > articles.Articles[j].UpdatedAt
	})

	return articles, nil
}

var (
	titleReg = regexp.MustCompile(`#\s+(.*?)\s*\n`)
)

func readArticle(username, id string, info os.FileInfo) (*types.ArticleDetail, error) {
	path := articleFile(username, id)
	if info == nil {
		var err error
		info, err = os.Stat(path)
		if err != nil {
			return nil, err
		}
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var (
		content = string(data)
		title   string
	)
	match := titleReg.FindStringSubmatch(content)
	if len(match) >= 2 {
		title = match[1]
	}

	return &types.ArticleDetail{
		ID:        id,
		Title:     title,
		Content:   content,
		UpdatedAt: info.ModTime().Unix(),
	}, nil
}

func saveArticle(username, id string, content string, init bool) error {
	dirPath := articlePath(username, id)
	if init && !goutils.FileExist(dirPath) {
		err := os.Mkdir(dirPath, 0755)
		if err != nil {
			return err
		}
	}

	filePath := articleFile(username, id)
	err := ioutil.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		return err
	}

	if !goutils.FileExist(filepath.Join(dirPath, ".git")) {
		cmds := [][]string{
			{"init"},
			{"add", FileName},
			{"commit", "-m", "'Initial commit.'"},
		}
		return gitCmd(dirPath, cmds)
	}

	if err := gitCmd(dirPath, [][]string{{"diff", "--quiet"}}); err != nil {
		return gitCmd(dirPath, [][]string{{"commit", "-am", "'Updated page.'"}})
	}
	return nil
}

func gitCmd(path string, cmds [][]string) error {
	for _, c := range cmds {
		c = append([]string{"-C", path}, c...)
		out, err := goutils.ExeCmd("git", c, nil)
		if err != nil {
			return errors.Wrapf(err, "exec git %v failed, %s", c, string(out))
		}
	}
	return nil
}
