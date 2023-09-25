package ds

import (
	ipfsds "github.com/ipfs/go-datastore"
)

var _ DownloadIface = (*DownloadDataStore)(nil)

type DownloadDataStore struct {
	ipfsds.Batching
}

func DownloadWrap(d ipfsds.Batching) *DownloadDataStore {
	return &DownloadDataStore{Batching: d}
}
