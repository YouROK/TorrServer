package template

import (
	"html/template"

	"github.com/gin-gonic/gin"
)

var ctx *gin.Engine

//
// func InitTemplate(route *gin.Engine) {
// 	ctx = route
//
// 	tmpl := getTemplate("favicon", faviconB64)
// 	tmpl = getTemplate("header", header)
//
// 	route.SetHTMLTemplate(tmpl)
// }
//
// func getTemplate(name, body string) *template.Template {
// 	tmpl, err := template.New(name).ParseFiles(body)
// 	if err != nil {
// 		log.TLogln("error parse template", err)
// 	}
// 	return tmpl
// }

type Template struct {
	templates *template.Template
}

func InitTemplate(c *gin.Engine) *Template {
	temp := new(Template)

	temp.parseMainPage()
	// temp.parseSettingsPage()
	// temp.parseAboutPage()
	// temp.parseCachePage()
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
	_, err := tmpl.Parse(s)
	if err != nil {
		return err
	}
	return nil
}
