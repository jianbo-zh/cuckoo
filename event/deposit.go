package event

type PushOfflineMessageEvt struct {
	ToPeerID string
	MsgID    string
	MsgData  []byte
}

type PullOfflineMessageEvt struct {
	HasMessage  func(peerID string, msgID string) (bool, error)
	SaveMessage func(peerID string, msgID string, msgData []byte) error
}
