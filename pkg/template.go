package pkg

import (
	"html/template"
	"io"
	"log"

	"github.com/Masterminds/sprig/v3"
)

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, _ ...string) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

// This is for the fiber interface, the blob loading is on createTemplate
func (t *Template) Load() error {
	return nil
}

func CreateTemplate() *Template {
	blobs, err := template.New("").Funcs(sprig.FuncMap()).ParseGlob("pkg/views/*.html")

	if err != nil {
		log.Panicln("failed to parse templates", err)
	}

	return &Template{templates: blobs}
}
