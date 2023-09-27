package myerror

import "fmt"

func WrapStreamError(msg string, err error) StreamErr {
	return StreamErr{msg: msg, inner: err}
}

type StreamErr struct {
	msg   string
	inner error
}

func (s StreamErr) Error() string {
	if s.inner == nil {
		return s.msg
	}
	return fmt.Sprintf("%s: %s", s.msg, s.inner.Error())
}

func (s StreamErr) Unwrap() error {
	return s.inner
}
