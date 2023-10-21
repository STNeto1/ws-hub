package pkg

import (
	"github.com/gofiber/fiber/v2"
)

func HandleIndex(c *fiber.Ctx) error {
	return c.Render("index.html", fiber.Map{})
}

//
// func HandleConnectionsSse(c *fiber.Ctx) error {
// 	c.Set("Content-Type", "text/event-stream")
// 	c.Set("Cache-Control", "no-cache")
// 	c.Set("Connection", "keep-alive")
// 	c.Set("Transfer-Encoding", "chunked")
//
// 	c.Context().SetBodyStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
// 		var i int
// 		for {
// 			i++
// 			msg := fmt.Sprintf("%d - the time is %v", i, time.Now())
// 			fmt.Fprintf(w, "data: Message: %s\n\n", msg)
//
// 			if err := w.Flush(); err != nil {
// 				// Refreshing page in web browser will establish a new
// 				// SSE connection, but only (the last) one is alive, so
// 				// dead connections must be closed here.
// 				fmt.Printf("Error while flushing: %v. Closing http connection.\n", err)
//
// 				break
// 			}
//
// 		}
// 	}))
//
// 	return nil
//
// }
