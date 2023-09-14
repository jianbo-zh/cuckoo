package groupsvc

import (
	"github.com/jianbo-zh/dchat/internal/types"
	"github.com/jianbo-zh/dchat/service/groupsvc/protocol/message/pb"
)

func decodeMsgType(mtype pb.Message_MsgType) types.MsgType {
	switch mtype {
	case pb.Message_TEXT:
		return types.MsgTypeText
	case pb.Message_IMAGE:
		return types.MsgTypeImage
	case pb.Message_AUDIO:
		return types.MsgTypeAudio
	case pb.Message_VIDEO:
		return types.MsgTypeVideo
	default:
		return types.MsgTypeUnknown
	}
}

func encodeMsgType(mtype types.MsgType) pb.Message_MsgType {
	switch mtype {
	case types.MsgTypeText:
		return pb.Message_TEXT
	case types.MsgTypeImage:
		return pb.Message_IMAGE
	case types.MsgTypeAudio:
		return pb.Message_AUDIO
	case types.MsgTypeVideo:
		return pb.Message_VIDEO
	default:
		return pb.Message_UNKNOWN
	}
}
