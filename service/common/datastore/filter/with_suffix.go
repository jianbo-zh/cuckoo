package filter

import (
	"strings"

	"github.com/ipfs/go-datastore/query"
)

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
