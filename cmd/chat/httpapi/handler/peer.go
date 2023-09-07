package handler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gofiber/fiber/v2"
	peersvc "github.com/jianbo-zh/dchat/service/peersvc"
	"github.com/libp2p/go-libp2p/core/peer"
)

var GetAddPeerHandler = func() fiber.Handler {
	return func(c *fiber.Ctx) error {
		peerID, _ := peer.Decode(c.Params("peerid"))
		nickname := c.Params("nickname")

		fmt.Println("peerID: ", peerID)

		err := peersvc.Get().AddPeer(context.Background(), peerID, nickname)
		if err != nil {
			return err
		}

		return c.SendString("ok")
	}
}

var GetGetPeersHandler = func() fiber.Handler {
	return func(c *fiber.Ctx) error {

		peers, err := peersvc.Get().GetPeers(context.Background())
		if err != nil {
			return err
		}

		if len(peers) == 0 {
			return c.SendString("nothing")
		}

		msgJson, _ := json.MarshalIndent(peers, "", "  ")

		return c.SendString(string(msgJson))
	}
}
