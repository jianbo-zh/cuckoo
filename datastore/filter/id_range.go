package filter

import (
	"strings"

	"github.com/ipfs/go-datastore/query"
)

type IDRangeFilter struct {
	StartID string
	EndID   string
}

func NewIDRangeFilter(startMsgID string, endMsgID string) *IDRangeFilter {
	return &IDRangeFilter{
		StartID: startMsgID,
		EndID:   endMsgID,
	}
}

func (f *IDRangeFilter) Filter(e query.Entry) bool {
	keys := strings.Split(e.Key, "/")
	msgID := keys[len(keys)-1]

	if msgID >= f.StartID && msgID <= f.EndID {
		return true
	}

	return false
}
