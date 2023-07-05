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

		peerID, _ := peer.Decode("12D3KooWNr9xsCWkavXGnyzfvSyuhTzva74x82zTieh6nHFJk7xf")

		err := peersvc.Get().SendTextMessage(context.Background(), peerID, "hello msg")
		if err != nil {
			return err
		}

		return c.SendString("ok") // => ✋ register
	}
}

var PingHandler = func() fiber.Handler {
	return func(c *fiber.Ctx) error {

		peerID, _ := peer.Decode(c.Params("peerid"))

		rtt, err := peersvc.Get().Ping(context.Background(), peerID)
		if err != nil {
			return err
		}

		return c.SendString(rtt.String()) // => ✋ register
	}
}

var GetMsgsHandler = func() fiber.Handler {
	return func(c *fiber.Ctx) error {
		peerID, _ := peer.Decode(c.Params("peerid"))

		fmt.Println("peerID: ", peerID)

		msgs, err := peersvc.Get().GetMessages(context.Background(), peerID)
		if err != nil {
			return err
		}

		if len(msgs) == 0 {
			return c.SendString("nothing")
		}

		mss := make([]map[string]any, len(msgs))
		for i, msg := range msgs {
			mss[i] = map[string]any{
				"id":             msg.GetId(),
				"type":           msg.GetType(),
				"data":           string(msg.GetPayload()),
				"send_id":        msg.GetSenderId(),
				"receive_id":     msg.GetReceiverId(),
				"send_timestamp": msg.GetSendTimestamp(),
				"lamport_time":   msg.GetLamportTime(),
			}
		}

		msgJson, _ := json.Marshal(mss)

		return c.SendString(string(msgJson))
	}
}
