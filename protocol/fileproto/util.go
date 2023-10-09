package fileproto

import (
	"encoding/json"
	"fmt"

	goproto "github.com/gogo/protobuf/proto"
	"github.com/jianbo-zh/dchat/internal/mytype"
	pb "github.com/jianbo-zh/dchat/protobuf/pb/filepb"
	"google.golang.org/protobuf/proto"
)

func convertFile(file *mytype.FileInfo) (*pb.FileInfo, error) {
	if file == nil {
		return nil, nil
	}

	var err error
	var fileType pb.FileType
	var metadata []byte
	switch file.FileType {
	case mytype.TextFile:
		fileType = pb.FileType_Text
	case mytype.ImageFile:
		fileType = pb.FileType_Image
		if file.Extension != nil {
			var extension mytype.ImageFileMetadata
			if err = json.Unmarshal(file.Extension, &extension); err != nil {
				return nil, fmt.Errorf("json.Unmarshal error: %w", err)

			}

			if metadata, err = proto.Marshal(&pb.ImageFileMetadata{
				Width:  extension.Width,
				Height: extension.Height,
			}); err != nil {
				return nil, fmt.Errorf("proto.Marshal error: %w", err)

			}
		}
	case mytype.VoiceFile:
		fileType = pb.FileType_Voice
		if file.Extension != nil {
			var extension mytype.VoiceFileMetadata
			if err = json.Unmarshal(file.Extension, &extension); err != nil {
				return nil, fmt.Errorf("json.Unmarshal error: %w", err)

			}

			if metadata, err = goproto.Marshal(&pb.VoiceFileMetadata{
				Duration: extension.Duration,
			}); err != nil {
				return nil, fmt.Errorf("proto.Marshal error: %w", err)

			}
		}
	case mytype.AudioFile:
		fileType = pb.FileType_Audio
		if file.Extension != nil {
			var extension mytype.AudioFileMetadata
			if err = json.Unmarshal(file.Extension, &extension); err != nil {
				return nil, fmt.Errorf("json.Unmarshal error: %w", err)

			}

			if metadata, err = goproto.Marshal(&pb.AudioFileMetadata{
				Duration: extension.Duration,
			}); err != nil {
				return nil, fmt.Errorf("proto.Marshal error: %w", err)

			}
		}
	case mytype.VideoFile:
		fileType = pb.FileType_Video
		if file.Extension != nil {
			var extension mytype.VideoFileMetadata
			if err = json.Unmarshal(file.Extension, &extension); err != nil {
				return nil, fmt.Errorf("json.Unmarshal error: %w", err)

			}

			if metadata, err = goproto.Marshal(&pb.VideoFileMetadata{
				Duration: extension.Duration,
			}); err != nil {
				return nil, fmt.Errorf("proto.Marshal error: %w", err)

			}
		}
	default:
		fileType = pb.FileType_Other
	}

	return &pb.FileInfo{
		Id:       file.FileID,
		FileType: fileType,
		MimeType: file.MimeType,
		FileName: file.FileName,
		FileSize: file.FileSize,
		FilePath: file.FileID,
		Metadata: metadata,
	}, nil
}
