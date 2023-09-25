package ds

import (
	"strings"

	"github.com/ipfs/go-datastore/query"
)

type FromKeyFilter struct {
	StartID string
}

func NewFromKeyFilter(startMsgID string) *FromKeyFilter {
	return &FromKeyFilter{
		StartID: startMsgID,
	}
}

func (f *FromKeyFilter) Filter(e query.Entry) bool {
	keys := strings.Split(e.Key, "/")
	return keys[len(keys)-1] > f.StartID
}

type WithSuffixFilter struct {
	Suffix []string
}

func (filter WithSuffixFilter) Filter(e query.Entry) bool {
	for _, suffix := range filter.Suffix {
		if strings.HasSuffix(e.Key, suffix) {
			return true
		}
	}

	return false
}
