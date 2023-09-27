package myerror

import (
	"errors"
	"fmt"
	"testing"
)

func TestWrapStreamError(t *testing.T) {
	err := fmt.Errorf("test")
	streamErr := WrapStreamError("asdfas", err)

	err2 := fmt.Errorf("dadfasf error: %w", streamErr)

	if !errors.Is(err2, streamErr) {
		t.Error("errors is failed")
	}

	if !errors.As(err2, &StreamErr{}) {
		t.Error("errors as failed")
	}
}
