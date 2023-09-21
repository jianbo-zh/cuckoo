package ds

import (
	"strconv"
	"strings"

	"github.com/ipfs/go-datastore/query"
)

type GroupContainFilter struct {
	Keywords string
}

func (filter GroupContainFilter) Filter(e query.Entry) bool {
	if strings.Contains(e.Key, filter.Keywords) {
		return true
	}

	return false
}

type GroupMemberFilter struct{}

func (filter GroupMemberFilter) Filter(e query.Entry) bool {

	if strings.Contains(e.Key, "_member_") {
		return true
	}

	return false
}

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

// GroupOrderByKey
type GroupOrderByKeyDescending struct{}

func (o GroupOrderByKeyDescending) Compare(a, b query.Entry) int {
	akeys := strings.SplitN(a.Key, "_", 2)
	bkeys := strings.SplitN(b.Key, "_", 2)

	if len(akeys) == 1 || len(bkeys) == 1 {
		return -strings.Compare(a.Key, b.Key)
	}

	a1, err := strconv.ParseUint(akeys[0], 10, 64)
	if err != nil {
		return -strings.Compare(a.Key, b.Key)
	}

	b1, err := strconv.ParseUint(bkeys[0], 10, 64)
	if err != nil {
		return -strings.Compare(a.Key, b.Key)
	}

	if a1 < b1 {
		return 1

	} else if a1 > b1 {
		return -1

	} else {
		return -strings.Compare(akeys[1], bkeys[1])
	}
}

func (GroupOrderByKeyDescending) String() string {
	return "desc(GroupKEY)"
}

// GroupOrderByValueTime
type GroupOrderByValueTimeDescending struct{}

func (o GroupOrderByValueTimeDescending) Compare(a, b query.Entry) int {
	a1, _ := strconv.ParseInt(string(a.Value), 10, 64)
	b1, _ := strconv.ParseInt(string(b.Value), 10, 64)

	if a1 <= b1 {
		return 1
	}
	return -1
}

func (GroupOrderByValueTimeDescending) String() string {
	return "desc(ValueTime)"
}