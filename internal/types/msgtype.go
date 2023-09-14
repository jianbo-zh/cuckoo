package types

type MsgType string

const (
	MsgTypeText    MsgType = "text"
	MsgTypeImage   MsgType = "image"
	MsgTypeAudio   MsgType = "audio"
	MsgTypeVideo   MsgType = "video"
	MsgTypeUnknown MsgType = "unknown"
)

func (m MsgType) String() string {
	return string(m)
}
