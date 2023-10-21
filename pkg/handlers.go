package pkg

import "github.com/gofiber/fiber/v2"

func HandleIndex(c *fiber.Ctx) error {
	return c.Render("index.html", fiber.Map{})
}
