package datastore

import (
	"bytes"
	"context"
	"errors"

	ds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	netpb "github.com/jianbo-zh/dchat/service/group/protocol/network/pb"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-varint"
	"google.golang.org/protobuf/proto"
)

func (gds *GroupDataStore) CachePeerConnect(ctx context.Context, groupID GroupID, pbmsg *netpb.GroupConnect) error {

	var pbmsgID string
	if pbmsg.PeerIdA < pbmsg.PeerIdB {
		pbmsgID = pbmsg.PeerIdA + "_" + pbmsg.PeerIdB
	} else {
		pbmsgID = pbmsg.PeerIdB + "_" + pbmsg.PeerIdA
	}

	key := ds.KeyWithNamespaces([]string{"dchat", "group", string(groupID), "network", "connects", pbmsgID})
	bs, err := memds.Get(ctx, key)
	if err != nil {
		return err
	}

	bs2, err := proto.Marshal(pbmsg)
	if err != nil {
		return err
	}

	var oldmsg netpb.GroupConnect
	err = proto.Unmarshal(bs, &oldmsg)
	if err != nil {
		if errors.Is(err, ds.ErrNotFound) {
			return memds.Put(ctx, key, bs2)
		}
		return err
	}

	if pbmsg.Lamportime > oldmsg.Lamportime {
		return memds.Put(ctx, key, bs2)
	}

	return nil
}

func (gds *GroupDataStore) GetPeerConnect(ctx context.Context, groupID GroupID, peerID1 peer.ID, peerID2 peer.ID) (*netpb.GroupConnect, error) {

	var pbmsgID string
	if peerID1.String() < peerID2.String() {
		pbmsgID = peerID1.String() + "_" + peerID2.String()
	} else {
		pbmsgID = peerID2.String() + "_" + peerID1.String()
	}

	key := ds.KeyWithNamespaces([]string{"dchat", "group", string(groupID), "network", "connects", pbmsgID})
	bs, err := memds.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	var pbmsg netpb.GroupConnect
	if err := proto.Unmarshal(bs, &pbmsg); err != nil {
		return nil, err
	}

	return &pbmsg, nil
}

func (gds *GroupDataStore) GetGroupConnects(ctx context.Context, groupID GroupID) ([]*netpb.GroupConnect, error) {

	results, err := memds.Query(ctx, query.Query{
		Prefix: "/dchat/group/" + string(groupID) + "/network/connects/",
	})
	if err != nil {
		return nil, err
	}

	var networks []*netpb.GroupConnect
	for result := range results.Next() {
		if result.Error != nil {
			return nil, err
		}

		var network netpb.GroupConnect
		if err := proto.Unmarshal(result.Entry.Value, &network); err != nil {
			return nil, err
		}

		networks = append(networks, &network)
	}

	return networks, nil
}

func (gds *GroupDataStore) GetNetworkLamportTime(ctx context.Context, groupID GroupID) (uint64, error) {
	gds.networkLamportMutex.Lock()
	defer gds.networkLamportMutex.Unlock()

	key := ds.KeyWithNamespaces([]string{"dchat", "group", string(groupID), "network", "lamportime"})

	tbs, err := gds.Get(ctx, key)
	if err != nil {
		if errors.Is(err, ds.ErrNotFound) {
			return 0, nil
		}
		return 0, err
	}

	return varint.ReadUvarint(bytes.NewReader(tbs))
}

func (gds *GroupDataStore) SetNetworkLamportTime(ctx context.Context, groupID GroupID, lamptime uint64) error {
	gds.networkLamportMutex.Lock()
	defer gds.networkLamportMutex.Unlock()

	key := ds.KeyWithNamespaces([]string{"dchat", "group", string(groupID), "network", "lamportime"})

	buff := make([]byte, varint.MaxLenUvarint63)
	len := varint.PutUvarint(buff, lamptime)

	return gds.Put(ctx, key, buff[:len])
}

func (gds *GroupDataStore) TickNetworkLamportTime(ctx context.Context, groupID GroupID) (uint64, error) {
	gds.networkLamportMutex.Lock()
	defer gds.networkLamportMutex.Unlock()

	key := ds.KeyWithNamespaces([]string{"dchat", "group", string(groupID), "network", "lamportime"})

	lamptime := uint64(0)

	if tbs, err := gds.Get(ctx, key); err != nil {
		if !errors.Is(err, ds.ErrNotFound) {
			return 0, err
		}

	} else if lamptime, err = varint.ReadUvarint(bytes.NewReader(tbs)); err != nil {
		return 0, err
	}

	buff := make([]byte, varint.MaxLenUvarint63)
	len := varint.PutUvarint(buff, lamptime+1)

	if err := gds.Put(ctx, key, buff[:len]); err != nil {
		return 0, err
	}

	return lamptime + 1, nil
}
