package service

import (
	"github.com/jianbo-zh/dchat/bind/grpc/proto"
	"github.com/jianbo-zh/dchat/internal/types"
)

func encodeMsgType(msgType string) proto.MsgType {
	switch msgType {
	case types.MsgTypeText:
		return proto.MsgType_Text
	case types.MsgTypeImage:
		return proto.MsgType_Image
	case types.MsgTypeAudio:
		return proto.MsgType_Audio
	case types.MsgTypeVideo:
		return proto.MsgType_Video
	default:
		return proto.MsgType_Unknown
	}
}

func decodeMsgType(msgType proto.MsgType) string {
	switch msgType {
	case proto.MsgType_Text:
		return types.MsgTypeText
	case proto.MsgType_Image:
		return types.MsgTypeImage
	case proto.MsgType_Audio:
		return types.MsgTypeAudio
	case proto.MsgType_Video:
		return types.MsgTypeVideo
	default:
		return types.MsgTypeUnknown
	}
}

func encodeOnlineState(state types.OnlineState) proto.ConnState {
	switch state {
	case types.OnlineStateOnline:
		return proto.ConnState_OnlineState
	case types.OnlineStateOffline:
		return proto.ConnState_OfflineState
	default:
		return proto.ConnState_UnknownState
	}
}
