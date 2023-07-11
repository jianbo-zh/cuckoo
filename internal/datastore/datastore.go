package datastore

// 扩展的 datastore （扩展包括 group相关操作、peer相关操作）

import (
	ds "github.com/ipfs/go-datastore"
	leveldb "github.com/ipfs/go-ds-leveldb"
)

type Config struct {
	Path string
}

func New(config Config) (ds.Batching, error) {
	return leveldb.NewDatastore(config.Path, nil)
}
