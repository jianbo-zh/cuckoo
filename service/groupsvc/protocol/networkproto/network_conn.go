package network

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/jianbo-zh/dchat/internal/myevent"
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
	sendConnInit := pb.ConnectInit{
		GroupId:   groupID,
		BootTs:    hostBootTs,
		ConnTimes: hostConnTimes,
	}
	if err := wt.WriteMsg(&sendConnInit); err != nil {
		stream.Reset()
		return fmt.Errorf("wt write msg error: %w", err)
	}

	var recvConnInit pb.ConnectInit
	rd := pbio.NewDelimitedReader(stream, maxMsgSize)
	if err := rd.ReadMsg(&recvConnInit); err != nil {
		stream.Reset()
		return fmt.Errorf("rd read msg error: %w", err)
	}

	peerConn := Connect{
		PeerID: peerID,
		sendCh: make(chan ConnectPair),
		doneCh: make(chan struct{}),
		reader: rd,
		writer: wt,
	}

	// 更新网络状态
	n.networkMutex.Lock()
	if _, exists := n.network[groupID]; !exists {
		n.network[groupID] = make(map[peer.ID]*Connect)
	}

	if _, exists := n.network[groupID][peerID]; exists {
		n.networkMutex.Unlock()
		stream.Reset()
		return fmt.Errorf("peer connect is exists")
	}

	n.network[groupID][peerID] = &peerConn
	n.networkMutex.Unlock()

	peerBootTs := recvConnInit.BootTs
	peerConnTimes := recvConnInit.ConnTimes

	go func() {
		var wg sync.WaitGroup

		wg.Add(2)
		go n.readStream(&wg, stream, groupID, peerID, peerConn.reader, peerConn.doneCh)
		go n.writeStream(&wg, stream, groupID, peerID, peerConn.writer, peerConn.sendCh, peerConn.doneCh)
		wg.Wait()

		stream.Close()
		n.triggerDisconnected(groupID, peerID, peerBootTs, peerConnTimes)
	}()

	// 触发连接改变事件
	if err := n.emitters.evtGroupConnectChange.Emit(myevent.EvtGroupConnectChange{
		GroupID:     groupID,
		PeerID:      peerID,
		IsConnected: true,
	}); err != nil {
		return err
	}

	// 转发连接改变事件
	n.updateRoutingAndForward(groupID, peerID,
		n.connectPair(groupID, n.host.ID(), n.bootTs, hostConnTimes, peerID, peerBootTs, peerConnTimes, StateConnected))

	return nil
}

// handleConnectRead 读取连接信息
func (n *NetworkProto) readStream(wg *sync.WaitGroup, stream network.Stream, groupID GroupID, peerID peer.ID, reader pbio.ReadCloser, doneCh <-chan struct{}) {

	defer wg.Done()

	var err error
	var connMaintain pb.ConnectMaintain

	for {
		select {
		case <-doneCh:
			return
		default:
			connMaintain.Reset()

			//  离线判断（接收方验证）
			if !n.isHeartbeatSender(n.host.ID(), peerID) {
				stream.SetReadDeadline(time.Now().Add(HeartbeatTimeout))
				if err = reader.ReadMsg(&connMaintain); err != nil {
					log.Errorf("read msg error: %v", err)
					return
				}
				stream.SetReadDeadline(time.Time{})

			} else {
				if err = reader.ReadMsg(&connMaintain); err != nil {
					log.Errorf("read msg error: %v", err)
					return
				}
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

			n.updateRoutingAndForward(groupID, peerID, conn)
		}
	}
}

// handleConnectWrite 处理连接写信息
func (n *NetworkProto) writeStream(wg *sync.WaitGroup, stream network.Stream, groupID string, peerID peer.ID, writer pbio.WriteCloser, sendCh <-chan ConnectPair, doneCh <-chan struct{}) {
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
			//  离线判断（发送方验证）
			if n.isHeartbeatSender(n.host.ID(), peerID) {
				stream.SetWriteDeadline(time.Now().Add(HeartbeatTimeout))
				err = writer.WriteMsg(&pb.ConnectMaintain{Type: pb.ConnectMaintain_Heartbeat})
				if err != nil {
					log.Errorf("write heartbeat msg error: %v", err)
					return
				}
				stream.SetWriteDeadline(time.Time{})

				log.Debugln("send heartbeat")
			}

		case <-doneCh:
			return
		}
	}
}

func (n *NetworkProto) disconnect(groupID GroupID, peerID peer.ID) error {
	n.networkMutex.Lock()
	defer n.networkMutex.Unlock()

	if _, exists := n.network[groupID][peerID]; !exists {
		return fmt.Errorf("conn not exists")
	}

	// disconnect signal
	close(n.network[groupID][peerID].doneCh)
	return nil
}

func (n *NetworkProto) triggerDisconnected(groupID GroupID, peerID peer.ID, peerBootTs uint64, peerConnTimes uint64) error {
	n.networkMutex.Lock()
	if _, exists := n.network[groupID]; exists {
		delete(n.network[groupID], peerID)
	}
	n.networkMutex.Unlock()

	if err := n.emitters.evtGroupConnectChange.Emit(myevent.EvtGroupConnectChange{
		GroupID:     groupID,
		PeerID:      peerID,
		IsConnected: false,
	}); err != nil {
		return err
	}

	log.Infof("trigger disconnect: %s, %s", groupID, peerID.String())

	return n.updateRoutingAndForward(groupID, peerID,
		n.connectPair(groupID, n.host.ID(), n.bootTs, n.incrConnTimes(), peerID, peerBootTs, peerConnTimes, StateDisconnected))
}

func (n *NetworkProto) updateRoutingAndForward(groupID GroupID, peerID peer.ID, conn ConnectPair) error {
	// 更新路由表
	if isUpdated := n.updateRoutingTable(groupID, conn); !isUpdated {
		return nil
	}
	log.Debugln("update group routing table")

	// 转发路由连接
	if err := n.handleForwardConnect(groupID, peerID, conn); err != nil {
		return fmt.Errorf("forward connect error: %w", err)
	}
	log.Debugln("forword peer connect")

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

func (n *NetworkProto) connectPair(groupID GroupID, hostID peer.ID, hostBootTs uint64, hostConnTimes uint64,
	peerID peer.ID, peerBootTs uint64, peerConnTimes uint64, state ConnState) ConnectPair {

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

func (n *NetworkProto) isHeartbeatSender(hostID peer.ID, peerID peer.ID) bool {
	return n.host.ID().String() < peerID.String()
}
