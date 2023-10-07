package mytype

const (
	TextMsgType  string = "text"
	ImageMsgType string = "image"
	VoiceMsgType string = "voice"
	AudioMsgType string = "audio"
	VideoMsgType string = "video"
	FileMsgType  string = "file"
)

type MessageState string

const (
	MessageStateSending MessageState = "sending"
	MessageStateSuccess MessageState = "success"
	MessageStateFail    MessageState = "fail"
)
