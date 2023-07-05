package datastore

import (
	"strconv"
	"strings"

	"github.com/ipfs/go-datastore/query"
)

type Group struct {
	ID   string
	Name string
}

type GroupFilter struct{}

func (filter GroupFilter) Filter(e query.Entry) bool {

	keys := strings.Split(strings.Trim(e.Key, "/"), "/")
	if len(keys) == 4 && keys[0] == "dchat" && keys[1] == "group" && keys[3] == "name" {
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
