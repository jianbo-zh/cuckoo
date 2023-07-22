package network

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	gevent "github.com/jianbo-zh/dchat/event"
	"github.com/jianbo-zh/dchat/service/group/protocol/network/pb"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-msgio/pbio"
)

func (n *NetworkService) connect(groupID GroupID, peerID peer.ID) error {

	ctx := context.Background()
	stream, err := n.host.NewStream(ctx, peerID, CONN_ID)
	if err != nil {
		return err
	}
	defer stream.Close()

	hostBootTs := n.bootTs
	hostConnTimes := n.incrConnTimes() // 连接或断开一次则1

	// 发送组网请求，同步Lamport时钟
	wt := pbio.NewDelimitedWriter(stream)
	if err := wt.WriteMsg(&pb.ConnectInit{GroupId: groupID, BootTs: hostBootTs, ConnTimes: hostConnTimes}); err != nil {
		return err
	}

	rd := pbio.NewDelimitedReader(stream, maxMsgSize)

	var connInit pb.ConnectInit
	if err := rd.ReadMsg(&connInit); err != nil {
		return err
	}

	if _, exists := n.network[groupID]; !exists {
		n.network[groupID] = make(map[peer.ID]*Connect)
	}

	if _, exists := n.network[groupID][peerID]; exists {
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

	// 触发已连接事件
	n.triggerConnected(groupID, peerID, peerBootTs, peerConnTimes, hostConnTimes, &peerConn)

	go func() {
		var wg sync.WaitGroup

		wg.Add(2)
		go n.readStream(&wg, groupID, peerID, peerConn.reader, peerConn.doneCh, peerBootTs, peerConnTimes)
		go n.writeStream(&wg, groupID, peerID, peerConn.writer, peerConn.sendCh, peerConn.doneCh)
		wg.Wait()

		defer n.triggerDisconnected(groupID, peerID, peerBootTs, peerConnTimes)
	}()

	return nil
}

// handleConnectRead 读取连接信息
func (n *NetworkService) readStream(wg *sync.WaitGroup, groupID GroupID, fromPeerID peer.ID, reader pbio.ReadCloser, doneCh <-chan struct{}, peerBootTs, peerConnTimes uint64) {

	defer wg.Done()
	defer reader.Close()

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

			n.triggerConnectChanged(groupID, fromPeerID, conn)
		}
	}
}

// handleConnectWrite 处理连接写信息
func (n *NetworkService) writeStream(wg *sync.WaitGroup, groupID string, toPeerID peer.ID, writer pbio.WriteCloser, sendCh <-chan ConnectPair, doneCh <-chan struct{}) {
	defer wg.Done()
	defer writer.Close()

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

		case <-time.After(HeartbeatTimeout):
			if n.host.ID().String() < toPeerID.String() { // 单边发送心跳
				err = writer.WriteMsg(&pb.ConnectMaintain{Type: pb.ConnectMaintain_Heartbeat})
				if err != nil {
					log.Errorf("write msg error: %v", err)
					return
				}
			}

		case <-doneCh:
			return
		}
	}
}

func (n *NetworkService) disconnect(groupID GroupID, peerID peer.ID) error {
	if _, exists := n.network[groupID][peerID]; !exists {
		return fmt.Errorf("conn not exists")
	}

	// disconnect signal
	close(n.network[groupID][peerID].doneCh)

	return nil
}

func (n *NetworkService) triggerConnected(groupID GroupID, peerID peer.ID, peerBootTs uint64, peerConnTimes uint64, hostConnTimes uint64, conn *Connect) error {
	n.network[groupID][peerID] = conn

	if err := n.emitters.evtGroupConnectChange.Emit(gevent.EvtGroupConnectChange{
		GroupID:     groupID,
		PeerID:      peerID,
		IsConnected: true,
	}); err != nil {
		return err
	}

	return n.triggerConnectChanged(groupID, peerID,
		n.getConnectPair(groupID, n.host.ID(), peerID, n.bootTs, peerBootTs, hostConnTimes, peerConnTimes, StateConnected))
}

func (n *NetworkService) triggerDisconnected(groupID GroupID, peerID peer.ID, peerBootTs uint64, peerConnTimes uint64) error {
	delete(n.network[groupID], peerID)

	if err := n.emitters.evtGroupConnectChange.Emit(gevent.EvtGroupConnectChange{
		GroupID:     groupID,
		PeerID:      peerID,
		IsConnected: false,
	}); err != nil {
		return err
	}

	return n.triggerConnectChanged(groupID, peerID,
		n.getConnectPair(groupID, n.host.ID(), peerID, n.bootTs, peerBootTs, n.incrConnTimes(), peerConnTimes, StateDisconnected))
}

func (n *NetworkService) triggerConnectChanged(groupID GroupID, peerID peer.ID, conn ConnectPair) error {
	// 更新路由表
	if isUpdated := n.updateRoutingTable(groupID, conn); !isUpdated {
		return nil
	}

	// 转发路由连接
	if err := n.handleForwardConnect(groupID, peerID, conn); err != nil {
		return err
	}

	return nil
}

// forwardConnectMaintain 转发更新连接
func (n *NetworkService) handleForwardConnect(groupID GroupID, excludePeerID peer.ID, conn0 ConnectPair) error {
	for peerID, conn := range n.network[groupID] {
		if peerID == excludePeerID {
			continue
		}

		conn.sendCh <- conn0
	}

	return nil
}

func (n *NetworkService) getConnectPair(groupID GroupID, hostID, peerID peer.ID, hostBootTs, peerBootTs, hostConnTimes, peerConnTimes uint64, state ConnState) ConnectPair {

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
