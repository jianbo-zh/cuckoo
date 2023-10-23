package service

import (
	"fmt"

	"github.com/jianbo-zh/dchat/bind/grpc/proto"
	"github.com/jianbo-zh/dchat/internal/mytype"
	goproto "google.golang.org/protobuf/proto"
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

func encodeOnlineState(state mytype.OnlineState) proto.OnlineState {
	switch state {
	case mytype.OnlineStateOnline:
		return proto.OnlineState_IsOnlineState
	case mytype.OnlineStateOffline:
		return proto.OnlineState_IsOfflineState
	default:
		return proto.OnlineState_UnknownOnlineState
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

func encodeFileInfo(file *mytype.FileInfo) *proto.FileInfo {
	var fileType proto.FileType
	switch file.FileType {
	case mytype.TextFile:
		fileType = proto.FileType_TextFileType
	case mytype.ImageFile:
		fileType = proto.FileType_ImageFileType
	case mytype.VoiceFile:
		fileType = proto.FileType_VoiceFileType
	case mytype.AudioFile:
		fileType = proto.FileType_AudioFileType
	case mytype.VideoFile:
		fileType = proto.FileType_VideoFileType
	default:
		fileType = proto.FileType_OtherFileType
	}

	return &proto.FileInfo{
		FileId:      file.FileID,
		FileType:    fileType,
		MimeType:    file.MimeType,
		Name:        file.FileName,
		Size:        file.FileSize,
		ThumbnailId: file.ThumbnailID,
		Width:       file.Width,
		Height:      file.Height,
		Duration:    file.Duration,
	}
}

func parseContactMessageFile(msg *mytype.ContactMessage) (file *mytype.FileInfo, err error) {

	switch msg.MsgType {
	case mytype.ImageMsgType:
		var pFile proto.ImageMessagePayload
		if err := goproto.Unmarshal(msg.Payload, &pFile); err != nil {
			return nil, fmt.Errorf("proto.Unmarshal payload error: %w", err)
		}
		file = &mytype.FileInfo{
			FileID:      pFile.ImageId,
			FileName:    pFile.Name,
			FileSize:    pFile.Size,
			FileType:    mytype.ImageFile,
			MimeType:    msg.MimeType,
			ThumbnailID: pFile.ThumbnailId,
			Width:       pFile.Width,
			Height:      pFile.Height,
		}
	case mytype.AudioMsgType:
		var pFile proto.AudioMessagePayload
		if err := goproto.Unmarshal(msg.Payload, &pFile); err != nil {
			return nil, fmt.Errorf("proto.Unmarshal payload error: %w", err)
		}
		file = &mytype.FileInfo{
			FileID:   pFile.AudioId,
			FileName: pFile.Name,
			FileSize: pFile.Size,
			FileType: mytype.AudioFile,
			MimeType: msg.MimeType,
			Duration: pFile.Duration,
		}

	case mytype.VideoMsgType:
		var pFile proto.VideoMessagePayload
		if err := goproto.Unmarshal(msg.Payload, &pFile); err != nil {
			return nil, fmt.Errorf("proto.Unmarshal payload error: %w", err)
		}
		file = &mytype.FileInfo{
			FileID:   pFile.VideoId,
			FileName: pFile.Name,
			FileSize: pFile.Size,
			FileType: mytype.VideoFile,
			MimeType: msg.MimeType,
			Duration: pFile.Duration,
		}

	case mytype.FileMsgType:
		var pFile proto.FileMessagePayload
		if err := goproto.Unmarshal(msg.Payload, &pFile); err != nil {
			return nil, fmt.Errorf("proto.Unmarshal payload error: %w", err)
		}
		file = &mytype.FileInfo{
			FileID:   pFile.FileId,
			FileName: pFile.Name,
			FileSize: pFile.Size,
			FileType: mytype.OtherFile,
			MimeType: msg.MimeType,
		}
	default:
		// not file msg
		return nil, fmt.Errorf("msg no file")
	}

	return file, nil
}

func parseGroupMessageFile(msg *mytype.GroupMessage) (file *mytype.FileInfo, err error) {
	switch msg.MsgType {
	case mytype.ImageMsgType:
		var pFile proto.ImageMessagePayload
		if err := goproto.Unmarshal(msg.Payload, &pFile); err != nil {
			return nil, fmt.Errorf("proto.Unmarshal payload error: %w", err)
		}
		file = &mytype.FileInfo{
			FileID:      pFile.ImageId,
			FileName:    pFile.Name,
			FileSize:    pFile.Size,
			FileType:    mytype.ImageFile,
			MimeType:    msg.MimeType,
			ThumbnailID: pFile.ThumbnailId,
			Width:       pFile.Width,
			Height:      pFile.Height,
		}
	case mytype.AudioMsgType:
		var pFile proto.AudioMessagePayload
		if err := goproto.Unmarshal(msg.Payload, &pFile); err != nil {
			return nil, fmt.Errorf("proto.Unmarshal payload error: %w", err)
		}
		file = &mytype.FileInfo{
			FileID:   pFile.AudioId,
			FileName: pFile.Name,
			FileSize: pFile.Size,
			FileType: mytype.AudioFile,
			MimeType: msg.MimeType,
			Duration: pFile.Duration,
		}

	case mytype.VideoMsgType:
		var pFile proto.VideoMessagePayload
		if err := goproto.Unmarshal(msg.Payload, &pFile); err != nil {
			return nil, fmt.Errorf("proto.Unmarshal payload error: %w", err)
		}
		file = &mytype.FileInfo{
			FileID:   pFile.VideoId,
			FileName: pFile.Name,
			FileSize: pFile.Size,
			FileType: mytype.VideoFile,
			MimeType: msg.MimeType,
			Duration: pFile.Duration,
		}

	case mytype.FileMsgType:
		var pFile proto.FileMessagePayload
		if err := goproto.Unmarshal(msg.Payload, &pFile); err != nil {
			return nil, fmt.Errorf("proto.Unmarshal payload error: %w", err)
		}
		file = &mytype.FileInfo{
			FileID:   pFile.FileId,
			FileName: pFile.Name,
			FileSize: pFile.Size,
			FileType: mytype.OtherFile,
			MimeType: msg.MimeType,
		}
	default:
		// not file msg
		return nil, fmt.Errorf("msg no file")
	}

	return file, nil
}
