package groupnetworkproto

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/jianbo-zh/dchat/internal/myevent"
	"github.com/jianbo-zh/dchat/internal/mytype"
	pb "github.com/jianbo-zh/dchat/protobuf/pb/grouppb"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-msgio/pbio"
)

func (n *NetworkProto) connect(groupID GroupID, peerID peer.ID) error {
	n.networkMutex.Lock()
	if _, exists := n.network[groupID].Conns[peerID]; exists {
		n.networkMutex.Unlock()
		log.Warnln("connect is exists, do nothing")
		return nil
	}
	n.networkMutex.Unlock()

	ctx := context.Background()
	stream, err := n.host.NewStream(network.WithDialPeerTimeout(ctx, mytype.DialTimeout), peerID, CONN_ID)
	if err != nil {
		return fmt.Errorf("host new stream error: %w", err)
	}

	fmt.Println("connect peer: ", peerID.String())

	hostBootTs := n.bootTs
	hostConnTimes := n.incrConnTimes() // 连接或断开一次则1

	fmt.Println("send conn init")
	// 发送组网请求，同步Lamport时钟
	wt := pbio.NewDelimitedWriter(stream)
	sendConnInit := pb.GroupConnectInit{
		GroupId:   groupID,
		BootTs:    hostBootTs,
		ConnTimes: hostConnTimes,
	}
	if err := wt.WriteMsg(&sendConnInit); err != nil {
		stream.Reset()
		return fmt.Errorf("wt write msg error: %w", err)
	}

	fmt.Println("recv conn init")
	var recvConnInit pb.GroupConnectInit
	rd := pbio.NewDelimitedReader(stream, mytype.PbioReaderMaxSizeNormal)
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

	fmt.Println("update network peerConn")

	// 更新网络状态
	n.networkMutex.Lock()
	n.network[groupID].Conns[peerID] = &peerConn
	n.networkMutex.Unlock()

	peerBootTs := recvConnInit.BootTs
	peerConnTimes := recvConnInit.ConnTimes

	fmt.Println("read write stream")
	go func() {
		var wg sync.WaitGroup

		wg.Add(2)
		go n.readStream(&wg, stream, groupID, peerID, peerConn.reader, peerConn.doneCh)
		go n.writeStream(&wg, stream, groupID, peerID, peerConn.writer, peerConn.sendCh, peerConn.doneCh)
		wg.Wait()

		fmt.Println("read write closed")

		stream.Close()
		if err := n.triggerDisconnected(groupID, peerID, peerBootTs, peerConnTimes); err != nil {
			fmt.Println("triggerDisconnected error: %w", err)
		}
	}()

	// 触发连接改变事件
	if err := n.emitters.evtGroupConnectChange.Emit(myevent.EvtGroupConnectChange{
		GroupID:     groupID,
		PeerID:      peerID,
		IsConnected: true,
	}); err != nil {
		return err
	}

	fmt.Println("update routing and froward")

	// 转发连接改变事件
	n.updateRoutingAndForward(groupID, peerID,
		n.connectPair(groupID, n.host.ID(), n.bootTs, hostConnTimes, peerID, peerBootTs, peerConnTimes, StateConnected))

	return nil
}

// handleConnectRead 读取连接信息
func (n *NetworkProto) readStream(wg *sync.WaitGroup, stream network.Stream, groupID GroupID, peerID peer.ID, reader pbio.ReadCloser, doneCh <-chan struct{}) {

	defer wg.Done()

	var err error
	var connMaintain pb.GroupConnectMaintain

	for {
		select {
		case <-doneCh:
			return
		default:
			connMaintain.Reset()

			//  离线判断（接收方验证）
			stream.SetReadDeadline(time.Now().Add(HeartbeatTimeout))
			if err = reader.ReadMsg(&connMaintain); err != nil {
				log.Errorf("read msg error: %v", err)
				return
			}
			stream.SetReadDeadline(time.Time{})

			if connMaintain.Type == pb.GroupConnectMaintain_Heartbeat || len(connMaintain.Payload) == 0 {
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
			err = writer.WriteMsg(&pb.GroupConnectMaintain{Type: pb.GroupConnectMaintain_ConnectChange, Payload: bs})
			if err != nil {
				log.Errorf("write msg error: %v", err)
				return
			}

		case <-time.After(HeartbeatInterval):
			stream.SetWriteDeadline(time.Now().Add(HeartbeatTimeout))
			err = writer.WriteMsg(&pb.GroupConnectMaintain{Type: pb.GroupConnectMaintain_Heartbeat})
			if err != nil {
				log.Errorf("write heartbeat msg error: %v", err)
				return
			}
			stream.SetWriteDeadline(time.Time{})

		case <-doneCh:
			return
		}
	}
}

func (n *NetworkProto) disconnect(groupID GroupID, peerID peer.ID) error {
	n.networkMutex.Lock()
	defer n.networkMutex.Unlock()

	if _, exists := n.network[groupID].Conns[peerID]; !exists {
		return fmt.Errorf("conn not exists")
	}

	// 发送断开信号
	close(n.network[groupID].Conns[peerID].doneCh)
	return nil
}

func (n *NetworkProto) triggerDisconnected(groupID GroupID, peerID peer.ID, peerBootTs uint64, peerConnTimes uint64) error {
	n.networkMutex.Lock()
	delete(n.network[groupID].Conns, peerID)
	n.networkMutex.Unlock()

	if err := n.emitters.evtGroupConnectChange.Emit(myevent.EvtGroupConnectChange{
		GroupID:     groupID,
		PeerID:      peerID,
		IsConnected: false,
	}); err != nil {
		return fmt.Errorf("emit EvtGroupConnectChange error: %w", err)
	}

	log.Infof("trigger disconnect: %s, %s", groupID, peerID.String())

	return n.updateRoutingAndForward(groupID, peerID,
		n.connectPair(groupID, n.host.ID(), n.bootTs, n.incrConnTimes(), peerID, peerBootTs, peerConnTimes, StateDisconnected))
}

func (n *NetworkProto) updateRoutingAndForward(groupID GroupID, peerID peer.ID, conn ConnectPair) error {
	log.Debugln("updateRoutingAndForward")
	// 更新路由表
	if isUpdated := n.updateRoutingTable(groupID, conn); !isUpdated {
		return nil
	}
	log.Debugln("update group routing table")

	// 转发路由连接
	if err := n.handleForwardConnect(groupID, conn, peerID); err != nil {
		return fmt.Errorf("forward connect error: %w", err)
	}
	log.Debugln("forword peer connect")

	// 触发路由更新
	if err := n.emitters.evtGroupRoutingTableChanged.Emit(myevent.EvtGroupRoutingTableChanged{GroupID: groupID}); err != nil {
		return fmt.Errorf("emit group routing table changed event error: %w", err)
	}

	return nil
}

// forwardConnectMaintain 转发更新连接
func (n *NetworkProto) handleForwardConnect(groupID GroupID, conn0 ConnectPair, excludePeerID peer.ID) error {
	for peerID, conn := range n.network[groupID].Conns {
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
