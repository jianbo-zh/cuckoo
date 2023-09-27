package httpapi

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/jianbo-zh/dchat/cmd/chat/httpapi/cuckobj"
	"github.com/jianbo-zh/dchat/cmd/chat/httpapi/handler"
	"github.com/jianbo-zh/dchat/cuckoo"
)

func Daemon(cuckoo *cuckoo.Cuckoo, config Config) {
	cuckobj.SetCuckoo(cuckoo)

	app := fiber.New()
	app.Get("/api/test/*", handler.TestHandler())
	app.Get("/api/account/create", handler.CreateAccountHandler())
	log.Fatal(app.Listen(fmt.Sprintf("%s:%d", config.Host, config.Port)))
}
