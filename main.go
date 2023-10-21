package main

import (
	"html/template"
	"io"
	"log"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/stneto1/ws-hub/pkg"
)

func main() {
	db := pkg.CreateDB()

	app := fiber.New(fiber.Config{
		Views: createTemplate(),
	})

	app.Use(logger.New())

	app.Get("/", pkg.HandleIndex)
	app.Get("/ws", websocket.New(pkg.HandleMessage, websocket.Config{}))

	go pkg.RunHub()
	go pkg.RunWorker(db)

	app.Listen(":3000")
}

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

func createTemplate() *Template {
	blobs, err := template.New("").Funcs(template.FuncMap{}).ParseGlob("pkg/views/*.html")

	if err != nil {
		log.Panicln("failed to parse templates", err)
	}

	return &Template{templates: blobs}
}
