package ds

import (
	"strconv"
	"strings"

	"github.com/ipfs/go-datastore/query"
)

type TimePrefixFilter struct {
	StartTime int64
	Sep       string
}

func (filter TimePrefixFilter) Filter(e query.Entry) bool {

	fields := strings.Split(e.Key, "/")
	fields2 := strings.SplitN(fields[len(fields)-1], filter.Sep, 2)

	depositTime, err := strconv.ParseInt(fields2[0], 10, 64)
	if err != nil {
		return false
	}

	return depositTime >= filter.StartTime
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
