package network

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	gevent "github.com/jianbo-zh/dchat/event"
	"github.com/jianbo-zh/dchat/service/groupsvc/protocol/networkproto/pb"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-msgio/pbio"
)

func (n *NetworkProto) connect(groupID GroupID, peerID peer.ID) error {

	ctx := context.Background()
	stream, err := n.host.NewStream(ctx, peerID, CONN_ID)
	if err != nil {
		return fmt.Errorf("host new stream error: %w", err)
	}

	fmt.Println("connect peer: ", peerID.String())

	hostBootTs := n.bootTs
	hostConnTimes := n.incrConnTimes() // 连接或断开一次则1

	// 发送组网请求，同步Lamport时钟
	wt := pbio.NewDelimitedWriter(stream)
	if err := wt.WriteMsg(&pb.ConnectInit{GroupId: groupID, BootTs: hostBootTs, ConnTimes: hostConnTimes}); err != nil {
		stream.Reset()
		return fmt.Errorf("wt write msg error: %w", err)
	}

	rd := pbio.NewDelimitedReader(stream, maxMsgSize)

	var connInit pb.ConnectInit
	if err := rd.ReadMsg(&connInit); err != nil {
		stream.Reset()
		return fmt.Errorf("rd read msg error: %w", err)
	}

	if _, exists := n.network[groupID]; !exists {
		n.network[groupID] = make(map[peer.ID]*Connect)
	}

	if _, exists := n.network[groupID][peerID]; exists {
		stream.Reset()
		return fmt.Errorf("peer connect is exists")
	}

	peerConn := Connect{
		PeerID: peerID,
		sendCh: make(chan ConnectPair),
		doneCh: make(chan struct{}),
		reader: rd,
		writer: wt,
	}

	peerBootTs := connInit.BootTs
	peerConnTimes := connInit.ConnTimes

	fmt.Println("2222")

	// 触发已连接事件
	err = n.triggerConnected(groupID, peerID, peerBootTs, peerConnTimes, hostConnTimes, &peerConn)
	if err != nil {
		stream.Reset()
		return fmt.Errorf("trigger connected error: %w", err)
	}

	go func() {
		var wg sync.WaitGroup

		wg.Add(2)
		go n.readStream(&wg, groupID, peerID, stream, peerConn.reader, peerConn.doneCh, peerBootTs, peerConnTimes)
		go n.writeStream(&wg, groupID, peerID, peerConn.writer, peerConn.sendCh, peerConn.doneCh)
		wg.Wait()

		stream.Close()
		n.triggerDisconnected(groupID, peerID, peerBootTs, peerConnTimes)
	}()

	return nil
}

// handleConnectRead 读取连接信息
func (n *NetworkProto) readStream(wg *sync.WaitGroup, groupID GroupID, fromPeerID peer.ID, stream network.Stream, reader pbio.ReadCloser, doneCh <-chan struct{}, peerBootTs, peerConnTimes uint64) {

	defer wg.Done()

	var err error
	var connMaintain pb.ConnectMaintain

	for {
		select {
		case <-doneCh:
			return
		default:
			connMaintain.Reset()
			if err = reader.ReadMsg(&connMaintain); err != nil {
				log.Errorf("read msg error: %v", err)
				return
			}

			if connMaintain.Type == pb.ConnectMaintain_Heartbeat || len(connMaintain.Payload) == 0 {
				continue
			}

			// 转发连接消息
			var conn ConnectPair
			if err = json.Unmarshal(connMaintain.Payload, &conn); err != nil {
				log.Errorf("json unmarshal error: %v", err)
				return
			}

			n.forwardConnectChanged(groupID, fromPeerID, conn)
		}
	}
}

// handleConnectWrite 处理连接写信息
func (n *NetworkProto) writeStream(wg *sync.WaitGroup, groupID string, toPeerID peer.ID, writer pbio.WriteCloser, sendCh <-chan ConnectPair, doneCh <-chan struct{}) {
	defer wg.Done()

	var err error
	for {
		select {
		case conn := <-sendCh:
			bs, err := json.Marshal(conn)
			if err != nil {
				log.Errorf("json marshal error: %v", err)
				return
			}

			err = writer.WriteMsg(&pb.ConnectMaintain{Type: pb.ConnectMaintain_ConnectChange, Payload: bs})
			if err != nil {
				log.Errorf("write msg error: %v", err)
				return
			}

		case <-time.After(HeartbeatInterval):
			if n.host.ID().String() < toPeerID.String() { // 单边发送心跳
				err = writer.WriteMsg(&pb.ConnectMaintain{Type: pb.ConnectMaintain_Heartbeat})
				if err != nil {
					log.Errorf("write heartbeat msg error: %v", err)
					return
				}
			}

		case <-doneCh:
			return
		}
	}
}

func (n *NetworkProto) disconnect(groupID GroupID, peerID peer.ID) error {
	if _, exists := n.network[groupID][peerID]; !exists {
		return fmt.Errorf("conn not exists")
	}

	// disconnect signal
	close(n.network[groupID][peerID].doneCh)

	return nil
}

func (n *NetworkProto) triggerConnected(groupID GroupID, peerID peer.ID, peerBootTs uint64, peerConnTimes uint64, hostConnTimes uint64, conn *Connect) error {
	n.network[groupID][peerID] = conn

	if err := n.emitters.evtGroupConnectChange.Emit(gevent.EvtGroupConnectChange{
		GroupID:     groupID,
		PeerID:      peerID,
		IsConnected: true,
	}); err != nil {
		return err
	}

	return n.forwardConnectChanged(groupID, peerID,
		n.getConnectPair(groupID, n.host.ID(), peerID, n.bootTs, peerBootTs, hostConnTimes, peerConnTimes, StateConnected))
}

func (n *NetworkProto) triggerDisconnected(groupID GroupID, peerID peer.ID, peerBootTs uint64, peerConnTimes uint64) error {
	delete(n.network[groupID], peerID)

	if err := n.emitters.evtGroupConnectChange.Emit(gevent.EvtGroupConnectChange{
		GroupID:     groupID,
		PeerID:      peerID,
		IsConnected: false,
	}); err != nil {
		return err
	}

	return n.forwardConnectChanged(groupID, peerID,
		n.getConnectPair(groupID, n.host.ID(), peerID, n.bootTs, peerBootTs, n.incrConnTimes(), peerConnTimes, StateDisconnected))
}

func (n *NetworkProto) forwardConnectChanged(groupID GroupID, peerID peer.ID, conn ConnectPair) error {
	fmt.Println("forward connect changed")
	// 更新路由表
	if isUpdated := n.updateRoutingTable(groupID, conn); !isUpdated {
		return nil
	}

	fmt.Println("routing updated")

	// 转发路由连接
	if err := n.handleForwardConnect(groupID, peerID, conn); err != nil {
		return fmt.Errorf("forward connect error: %w", err)
	}

	fmt.Println("forward connect")

	return nil
}

// forwardConnectMaintain 转发更新连接
func (n *NetworkProto) handleForwardConnect(groupID GroupID, excludePeerID peer.ID, conn0 ConnectPair) error {
	for peerID, conn := range n.network[groupID] {
		if peerID == excludePeerID {
			continue
		}

		conn.sendCh <- conn0
	}

	return nil
}

func (n *NetworkProto) getConnectPair(groupID GroupID, hostID, peerID peer.ID, hostBootTs, peerBootTs, hostConnTimes, peerConnTimes uint64, state ConnState) ConnectPair {

	if hostID.String() < peerID.String() {
		return ConnectPair{
			GroupID:    groupID,
			PeerID0:    hostID,
			PeerID1:    peerID,
			BootTs0:    hostBootTs,
			BootTs1:    peerBootTs,
			ConnTimes0: hostConnTimes,
			ConnTimes1: peerConnTimes,
			State:      state,
		}
	}

	return ConnectPair{
		GroupID:    groupID,
		PeerID0:    peerID,
		PeerID1:    hostID,
		BootTs0:    peerBootTs,
		BootTs1:    hostBootTs,
		ConnTimes0: peerConnTimes,
		ConnTimes1: hostConnTimes,
		State:      state,
	}
}
