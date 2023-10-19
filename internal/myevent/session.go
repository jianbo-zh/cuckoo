package myevent

// EvtClearSession 清空会话
type EvtClearSession struct {
	SessionID string
	Result    chan<- error
}

// EvtDeleteSession 删除会话
type EvtDeleteSession struct {
	SessionID string
	Result    chan<- error
}
