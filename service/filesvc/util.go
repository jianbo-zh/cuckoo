package filesvc

import (
	"github.com/jianbo-zh/dchat/internal/mytype"
	pb "github.com/jianbo-zh/dchat/service/filesvc/protobuf/pb/filepb"
)

func decodeFile(file *pb.FileInfo) *mytype.FileInfo {
	if file == nil {
		return nil
	}

	var fileType mytype.FileType
	switch file.FileType {
	case pb.FileType_Text:
		fileType = mytype.TextFile
	case pb.FileType_Image:
		fileType = mytype.ImageFile
	case pb.FileType_Voice:
		fileType = mytype.VoiceFile
	case pb.FileType_Audio:
		fileType = mytype.AudioFile
	case pb.FileType_Video:
		fileType = mytype.VideoFile
	default:
		fileType = mytype.OtherFile
	}

	return &mytype.FileInfo{
		FileID:      file.FileId,
		FileName:    file.FileName,
		FileSize:    file.FileSize,
		FileType:    fileType,
		MimeType:    file.MimeType,
		ThumbnailID: file.ThumbnailId,
		Width:       file.Width,
		Height:      file.Height,
		Duration:    file.Duration,
	}
}
