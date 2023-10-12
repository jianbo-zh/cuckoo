package mytype

import (
	"fmt"
	"strings"
)

type FileID struct {
	HashAlgo  string
	HashValue string
	Extension string
}

func (f *FileID) String() string {
	return fmt.Sprintf("%s_%s.%s", f.HashAlgo, f.HashValue, strings.TrimLeft(f.Extension, "."))
}

type FileType string

const (
	TextFile  FileType = "text"
	ImageFile FileType = "image"
	VoiceFile FileType = "voice"
	AudioFile FileType = "audio"
	VideoFile FileType = "video"
	OtherFile FileType = "other"
)

type FileInfo struct {
	FileID      string
	FileName    string
	FileSize    int64
	FileType    FileType
	MimeType    string
	ThumbnailID string
	Width       int32
	Height      int32
	Duration    int32
}
