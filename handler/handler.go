package handler

import (
	"html/template"

	"github.com/pkg/errors"
)

var rootTmpl *template.Template

func Init() error {
	var err error
	rootTmpl, err = template.ParseGlob("static/tmpl/*.gohtml")
	if err != nil {
		return errors.Wrap(err, "load template failed")
	}
	return nil
}
