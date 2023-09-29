package service

import (
	"github.com/jianbo-zh/dchat/bind/grpc/proto"
	"github.com/jianbo-zh/dchat/internal/mytype"
)

func encodeMsgType(msgType string) proto.MsgType {
	switch msgType {
	case mytype.MsgTypeText:
		return proto.MsgType_Text
	case mytype.MsgTypeImage:
		return proto.MsgType_Image
	case mytype.MsgTypeAudio:
		return proto.MsgType_Audio
	case mytype.MsgTypeVideo:
		return proto.MsgType_Video
	default:
		return proto.MsgType_Unknown
	}
}

func decodeMsgType(msgType proto.MsgType) string {
	switch msgType {
	case proto.MsgType_Text:
		return mytype.MsgTypeText
	case proto.MsgType_Image:
		return mytype.MsgTypeImage
	case proto.MsgType_Audio:
		return mytype.MsgTypeAudio
	case proto.MsgType_Video:
		return mytype.MsgTypeVideo
	default:
		return mytype.MsgTypeUnknown
	}
}

func encodeOnlineState(state mytype.OnlineState) proto.ConnState {
	switch state {
	case mytype.OnlineStateOnline:
		return proto.ConnState_OnlineState
	case mytype.OnlineStateOffline:
		return proto.ConnState_OfflineState
	default:
		return proto.ConnState_UnknownState
	}
}
