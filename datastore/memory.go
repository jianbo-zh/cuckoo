package datastore

import (
	ds "github.com/ipfs/go-datastore"
)

// 临时存储
var memds ds.Datastore

func init() {
	memds = ds.NewMapDatastore()
}
