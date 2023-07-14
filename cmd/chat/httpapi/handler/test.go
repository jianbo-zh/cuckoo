package handler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gofiber/fiber/v2"
	peersvc "github.com/jianbo-zh/dchat/service/peer"
	"github.com/libp2p/go-libp2p/core/peer"
)

var TestHandler = func() fiber.Handler {
	return func(c *fiber.Ctx) error {
		msg := fmt.Sprintf("✋ %s", c.Params("*"))
		return c.SendString(msg) // => ✋ register
	}
}

var SendMsgHandler = func() fiber.Handler {
	return func(c *fiber.Ctx) error {

		peerID, _ := peer.Decode(c.Params("peerid"))
		msgtxt := c.Params("msgtxt")

		err := peersvc.Get().SendTextMessage(context.Background(), peerID, msgtxt)
		if err != nil {
			return err
		}

		return c.SendString("ok") // => ✋ register
	}
}

var GetMsgsHandler = func() fiber.Handler {
	return func(c *fiber.Ctx) error {
		peerID, _ := peer.Decode(c.Params("peerid"))

		fmt.Println("peerID: ", peerID)

		msgs, err := peersvc.Get().GetMessages(context.Background(), peerID, 0, 10)
		if err != nil {
			return err
		}

		if len(msgs) == 0 {
			return c.SendString("nothing")
		}

		msgJson, _ := json.Marshal(msgs)

		return c.SendString(string(msgJson))
	}
}
