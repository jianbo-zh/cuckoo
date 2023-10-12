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

func DecodeFileID(fileID string) (*FileID, error) {
	s1 := strings.Index(fileID, "_")
	if s1 <= 0 || s1+1 >= len(fileID) {
		return nil, fmt.Errorf("not found hash algo")
	}
	s2 := strings.Index(fileID, ".")
	if s2 <= 0 || s2+1 >= len(fileID) {
		return nil, fmt.Errorf("not found extension")
	}

	return &FileID{
		HashAlgo:  fileID[:s1],
		HashValue: fileID[s1+1 : s2],
		Extension: fileID[s2+1:],
	}, nil
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
