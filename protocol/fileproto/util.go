package fileproto

import (
	"fmt"
	"sync"

	"github.com/jianbo-zh/dchat/internal/mytype"
	pb "github.com/jianbo-zh/dchat/protobuf/pb/filepb"
)

type MultiHandleResult struct {
	curr  int64
	total int64
	mutex sync.Mutex
}

func NewMultiHandleResult(total int64) *MultiHandleResult {
	return &MultiHandleResult{total: total}
}

func (m *MultiHandleResult) IncreOne() (isFinish bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.curr++
	return m.curr >= m.total
}

func encodeFile(file *mytype.FileInfo) *pb.FileInfo {
	if file == nil {
		return nil
	}

	var fileType pb.FileType
	switch file.FileType {
	case mytype.TextFile:
		fileType = pb.FileType_Text
	case mytype.ImageFile:
		fileType = pb.FileType_Image
	case mytype.VoiceFile:
		fileType = pb.FileType_Voice
	case mytype.AudioFile:
		fileType = pb.FileType_Audio
	case mytype.VideoFile:
		fileType = pb.FileType_Video
	default:
		fileType = pb.FileType_Other
	}

	return &pb.FileInfo{
		FileId:      file.FileID,
		FileType:    fileType,
		MimeType:    file.MimeType,
		FileName:    file.FileName,
		FileSize:    file.FileSize,
		ThumbnailId: file.ThumbnailID,
		Width:       file.Width,
		Height:      file.Height,
		Duration:    file.Duration,
	}
}

func chunkFileName(fileID string, chunkIndex int64) string {
	return fmt.Sprintf("%s.part_%d", fileID, chunkIndex)
}
