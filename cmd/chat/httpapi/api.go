package httpapi

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/jianbo-zh/dchat/cmd/chat/httpapi/handler"
)

func Daemon(config Config) {
	app := fiber.New()
	app.Get("/api/test/*", handler.TestHandler())
	app.Get("/api/sendmsg", handler.SendMsgHandler())
	app.Get("/api/getmsgs/:peerid", handler.GetMsgsHandler())
	app.Get("/api/ping/:peerid", handler.PingHandler())
	log.Fatal(app.Listen(fmt.Sprintf("%s:%d", config.Host, config.Port)))
}
