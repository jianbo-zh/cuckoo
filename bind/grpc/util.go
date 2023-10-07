package service

import (
	"github.com/jianbo-zh/dchat/bind/grpc/proto"
	"github.com/jianbo-zh/dchat/internal/mytype"
)

func encodeMsgType(msgType string) proto.MsgType {
	switch msgType {
	case mytype.TextMsgType:
		return proto.MsgType_Text
	case mytype.ImageMsgType:
		return proto.MsgType_Image
	case mytype.VoiceMsgType:
		return proto.MsgType_Voice
	case mytype.AudioMsgType:
		return proto.MsgType_Audio
	case mytype.VideoMsgType:
		return proto.MsgType_Video
	case mytype.FileMsgType:
		return proto.MsgType_File
	default:
		return proto.MsgType_Unknown
	}
}

func decodeMsgType(msgType proto.MsgType) string {
	switch msgType {
	case proto.MsgType_Text:
		return mytype.TextMsgType
	case proto.MsgType_Image:
		return mytype.ImageMsgType
	case proto.MsgType_Voice:
		return mytype.VoiceMsgType
	case proto.MsgType_Audio:
		return mytype.AudioMsgType
	case proto.MsgType_Video:
		return mytype.VideoMsgType
	case proto.MsgType_File:
		return mytype.FileMsgType
	default:
		return ""
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

func encodeMessageState(state mytype.MessageState) proto.MsgState {
	switch state {
	case mytype.MessageStateSuccess:
		return proto.MsgState_SendSucc
	case mytype.MessageStateFail:
		return proto.MsgState_SendFail
	default:
		return proto.MsgState_Sending
	}
}
