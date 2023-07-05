package service

import (
	"context"

	"github.com/libp2p/go-libp2p/core/peer"
)

type CreateGroupParam struct {
	Name    string
	Members []peer.ID
}

func (svc *TopService) CreateGroup(param CreateGroupParam) error {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	groupID, err := svc.groupSvc.CreateGroup(ctx, param.Name, param.Members)
	if err != nil {
		return err
	}

	for _, memberID := range param.Members {
		err := svc.peerSvc.SendGroupInviteMessage(ctx, memberID, groupID)
		if err != nil {
			return err
		}
	}

	return nil
}

type InviteMemberParam struct {
	GroupID string
	PeerID  peer.ID
}

func (svc *TopService) InviteMember(param InviteMemberParam) error {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := svc.groupSvc.InviteMember(ctx, param.GroupID, param.PeerID); err != nil {
		return err
	}

	if err := svc.peerSvc.SendGroupInviteMessage(ctx, param.PeerID, param.GroupID); err != nil {
		return err
	}

	return nil
}
