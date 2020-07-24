package handler

import (
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/tidyoux/giki/handler/types"
	"github.com/tidyoux/goutils"
)

func ListArticle(c *gin.Context) {
	username, _, _ := c.Request.BasicAuth()
	tag := c.Query("tag")
	articles, err := listArticle(username, tag)
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
	article, err := readArticle(username, id)
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
	article, err := readArticle(username, id)
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

func listArticle(username, tag string) (*types.Wiki, error) {
	dbPath := userDBPath(username)
	if !goutils.FileExist(dbPath) {
		err := os.MkdirAll(dbPath, 0755)
		if err != nil {
			return nil, err
		}

		return types.NewWiki(username, nil, ""), nil
	}

	var articles []*types.Article
	err := filepath.Walk(dbPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if path == dbPath || !info.IsDir() {
			return nil
		}

		article, err := readArticle(username, info.Name())
		if err != nil {
			return errors.Wrapf(err, "read article %s failed", info.Name())
		}

		if article.HasTag(tag) {
			articles = append(articles, article)
		}
		return filepath.SkipDir
	})
	if err != nil {
		return nil, errors.Wrap(err, "read db failed")
	}

	return types.NewWiki(username, articles, tag), nil
}

func readArticle(username, id string) (*types.Article, error) {
	filePath := articleFile(username, id)
	info, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return types.NewArticle(id, string(data), info.ModTime().Unix()), nil
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
