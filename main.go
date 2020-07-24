package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/tidyoux/giki/handler"
	"github.com/tidyoux/goutils/cmd"
)

var (
	_port int
)

func main() {
	c := cmd.New(
		"giki",
		"giki is a wiki",
		" ./giki -p 65312",
		run,
	)

	flags := c.Flags()
	flags.IntVarP(&_port, "port", "p", 65321, "service port")

	c.Execute()
}

func run(*cmd.Command) error {
	err := handler.Init()
	if err != nil {
		return err
	}

	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()
	r.Use(gin.BasicAuth(map[string]string{
		"admin": "123456",
	}))

	v1 := r.Group("v1")
	v1.Static("/static", "./static")

	v1.GET("/article", handler.ListArticle)
	v1.POST("/article", handler.CreateArticle)
	v1.GET("/article/:id", handler.ViewArticle)
	v1.GET("/article/:id/edit", handler.EditArticle)
	v1.POST("/article/:id", handler.UpdateArticle)
	v1.POST("/article/:id/delete", handler.DeleteArticle)

	return r.Run(fmt.Sprintf(":%d", _port))
}
