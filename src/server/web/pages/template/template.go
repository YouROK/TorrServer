package template

import (
	"html/template"

	"github.com/gin-gonic/gin"
)

var ctx *gin.Engine

type Template struct {
	templates *template.Template
}

func InitTemplate(c *gin.Engine) *Template {
	temp := new(Template)

	temp.parsePage("mainPage", mainPage)
	// temp.parsePage("apijsPage", apiJS)
	// temp.parsePage("mainjsPage", mainJS)

	c.SetHTMLTemplate(temp.templates)
	return temp
}

func (t *Template) render(c *gin.Context, code int, name string, data interface{}) {
	c.HTML(code, name, data)
}

func (t *Template) parsePage(name, page string) error {
	s := page
	var tmpl *template.Template
	if t.templates == nil {
		t.templates = template.New(name)
	}
	if name == t.templates.Name() {
		tmpl = t.templates
	} else {
		tmpl = t.templates.New(name)
	}
	tmpl.Delims("<<", ">>")
	_, err := tmpl.Parse(s)
	if err != nil {
		return err
	}
	return nil
}
