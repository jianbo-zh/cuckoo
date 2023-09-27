package handler

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/jianbo-zh/dchat/cmd/chat/httpapi/cuckobj"
	"github.com/jianbo-zh/dchat/internal/types"
)

var CreateAccountHandler = func() fiber.Handler {
	return func(c *fiber.Ctx) error {
		name := "mac pro"
		avatar := "md5_8463a44c7ea4a84236e3dc7cf49c0ab5.jpg"

		accountSvc, err := cuckobj.GetCuckoo().GetAccountSvc()
		if err != nil {
			return fmt.Errorf("get account svc error: %w", err)
		}

		accountSvc.CreateAccount(context.Background(), types.Account{
			Name:                 name,
			Avatar:               avatar,
			AutoAddContact:       true,
			AutoJoinGroup:        true,
			EnableDepositService: true,
		})
		return c.SendString("ok")
	}
}
