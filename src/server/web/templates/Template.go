package templates

import (
	"html/template"
	"io"

	"github.com/labstack/echo"
)

type Template struct {
	templates *template.Template
}

func InitTemplate(e *echo.Echo) {
	temp := new(Template)

	temp.parseMainPage()
	temp.parseSettingsPage()
	temp.parseAboutPage()
	temp.parseCachePage()

	e.Renderer = temp
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func parsePage(temp *Template, name, page string) error {
	s := page
	var tmpl *template.Template
	if temp.templates == nil {
		temp.templates = template.New(name)
	}
	if name == temp.templates.Name() {
		tmpl = temp.templates
	} else {
		tmpl = temp.templates.New(name)
	}
	_, err := tmpl.Parse(s)
	if err != nil {
		return err
	}
	return nil
}
