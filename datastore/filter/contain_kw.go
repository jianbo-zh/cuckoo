package filter

import (
	"strings"

	"github.com/ipfs/go-datastore/query"
)

type GroupContainFilter struct {
	Keywords string
}

func (filter GroupContainFilter) Filter(e query.Entry) bool {
	return strings.Contains(e.Key, filter.Keywords)
}
