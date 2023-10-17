package service

import (
	"context"
	"fmt"

	"github.com/jianbo-zh/dchat/bind/grpc/proto"
	"github.com/jianbo-zh/dchat/cuckoo"
	"github.com/jianbo-zh/dchat/internal/util"
)

var _ proto.CommonSvcServer = (*CommonSvc)(nil)

type CommonSvc struct {
	proto.UnimplementedCommonSvcServer
}

func NewCommonSvc(getter cuckoo.CuckooGetter) *CommonSvc {
	return &CommonSvc{}
}

// DecodeQRCodeToken 解码Token
func (c *CommonSvc) DecodeQRCodeToken(ctx context.Context, request *proto.DecodeQRCodeTokenRequest) (reply *proto.DecodeQRCodeTokenReply, err error) {

	log.Infoln("DecodeQRCodeToken request: ", request.String())
	defer func() {
		if e := recover(); e != nil {
			log.Panicln("DecodeQRCodeToken panic: ", e)
		} else if err != nil {
			log.Errorln("DecodeQRCodeToken error: ", err.Error())
		} else {
			log.Infoln("DecodeQRCodeToken reply: ", reply.String())
		}
	}()

	token, err := util.DecodeQRCodeToken(request.Token)
	if err != nil {
		return nil, fmt.Errorf("util.DecodeQRCodeToken error: %w", err)
	}

	var tokenType proto.QRCodeToken_TokenType
	switch token.Type {
	case util.TokenTypePeer:
		tokenType = proto.QRCodeToken_Peer
	case util.TokenTypeGroup:
		tokenType = proto.QRCodeToken_Group
	default:
		tokenType = proto.QRCodeToken_Unknown
	}

	reply = &proto.DecodeQRCodeTokenReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		Token: &proto.QRCodeToken{
			TokenType:      tokenType,
			Id:             token.ID,
			Name:           token.Name,
			Avatar:         token.Name,
			DepositAddress: token.DepositAddr,
		},
	}
	return reply, nil
}
