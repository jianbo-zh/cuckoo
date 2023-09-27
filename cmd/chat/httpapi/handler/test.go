package handler

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

var TestHandler = func() fiber.Handler {
	return func(c *fiber.Ctx) error {
		msg := fmt.Sprintf("✋ %s", c.Params("*"))
		return c.SendString(msg) // => ✋ register
	}
}
