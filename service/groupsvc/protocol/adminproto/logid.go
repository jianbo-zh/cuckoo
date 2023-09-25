package adminproto

import (
	"fmt"

	"github.com/jianbo-zh/dchat/internal/util"
	"github.com/jianbo-zh/dchat/service/groupsvc/protocol/adminproto/ds"
	"github.com/libp2p/go-libp2p/core/peer"
)

func logIdCreate(lamptime uint64, creator peer.ID) string {
	return fmt.Sprintf("%s%s%s", util.Lamptime019(lamptime), ds.KwCreate, creator.String())
}
func logIdName(lamptime uint64, creator peer.ID) string {
	return fmt.Sprintf("%s%s%s", util.Lamptime019(lamptime), ds.KwName, creator.String())
}

func logIdAvatar(lamptime uint64, creator peer.ID) string {
	return fmt.Sprintf("%s%s%s", util.Lamptime019(lamptime), ds.KwAvatar, creator.String())
}

func logIdNotice(lamptime uint64, creator peer.ID) string {
	return fmt.Sprintf("%s%s%s", util.Lamptime019(lamptime), ds.KwNotice, creator.String())
}

func logIdAutojoin(lamptime uint64, creator peer.ID) string {
	return fmt.Sprintf("%s%s%s", util.Lamptime019(lamptime), ds.KwAutoJoin, creator.String())
}

func logIdDepositPeer(lamptime uint64, depositPeer peer.ID) string {
	return fmt.Sprintf("%s%s%s", util.Lamptime019(lamptime), ds.KwDepositPeer, depositPeer.String())
}

func logIdMember(lamptime uint64, creator peer.ID) string {
	return fmt.Sprintf("%s%s%s", util.Lamptime019(lamptime), ds.KwMember, creator.String())
}

func logIdExit(lamptime uint64, creator peer.ID) string {
	return fmt.Sprintf("%s%s%s", util.Lamptime019(lamptime), ds.KwExit, creator.String())
}

func logIdDisband(lamptime uint64, creator peer.ID) string {
	return fmt.Sprintf("%s%s%s", util.Lamptime019(lamptime), ds.KwDisband, creator.String())
}
