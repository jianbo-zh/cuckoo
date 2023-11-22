package core

import (
	"context"
	"fmt"
	"sync"

	logging2 "github.com/ipfs/go-log/v2"
	"github.com/jianbo-zh/dchat/bind/utils"
	"github.com/jianbo-zh/dchat/cuckoo"
	cuckooConfig "github.com/jianbo-zh/dchat/cuckoo/config"
	"github.com/jianbo-zh/go-anet"
	logging "github.com/jianbo-zh/go-log"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	"go.uber.org/zap"
)

var (
	node   *Node
	locker sync.Mutex
)

type Node struct {
	cuckoo       *cuckoo.Cuckoo
	cuckooConfig *cuckooConfig.Config
	nodeConfig   *NodeConfig
	mdnsService  mdns.Service
}

func (n *Node) GetCuckoo() (*cuckoo.Cuckoo, error) {
	if n.cuckoo == nil {
		return nil, fmt.Errorf("cuckoo not start")
	}
	return n.cuckoo, nil
}

func StartNode(config *NodeConfig) error {
	locker.Lock()
	defer locker.Unlock()

	logging.SetupLogging(logging.Config{
		Level:  logging.LevelDebug,
		Format: logging.FormatPlaintextOutput,
		Stdout: true,
	})

	logging2.SetLogLevel("*", "ERROR")

	if node != nil {
		return fmt.Errorf("node is started")
	}

	cuckooConf, err := cuckoo.LoadConfig(config.dataDir, config.resourceDir, config.fileDir, config.tmpDir)
	if err != nil {
		return fmt.Errorf("cuckoo.LoadConfig: %s", err.Error())
	}

	// Set up netdriver.
	if config.netDriver != nil {
		logger, _ := zap.NewDevelopment()
		inet := &inet{
			net:    config.netDriver,
			logger: logger,
		}
		utils.SetNetDriver(inet)
		anet.SetNetDriver(inet)
	}

	mdnsLocked := false
	if cuckooConf.Discovery.MDNS.Enabled && config.mdnsLockerDriver != nil {
		config.mdnsLockerDriver.Lock()
		mdnsLocked = true
		cuckooConf.Discovery.MDNS.Enabled = false
	}

	ctx := context.Background()

	cuckooNode, err := cuckoo.NewCuckoo(ctx, cuckooConf)
	if err != nil {
		return fmt.Errorf("cuckoo.NewCuckoo error: %s", err.Error())
	}

	var mdnsService mdns.Service
	if mdnsLocked {
		cuckooConf.Discovery.MDNS.Enabled = true

		mdnslogger, _ := zap.NewDevelopment()
		host, _ := cuckooNode.GetHost()
		dh := utils.DiscoveryHandler(ctx, mdnslogger, host)
		mdnsService = utils.NewMdnsService(mdnslogger, host, utils.MDNSServiceName, dh)

		// Start the mDNS service.
		// Get multicast interfaces.
		ifaces, err := utils.GetMulticastInterfaces()
		if err != nil {
			if mdnsLocked {
				config.mdnsLockerDriver.Unlock()
			}
			return fmt.Errorf("utils.GetMulticastInterfaces error: %s", err.Error())
		}

		// If multicast interfaces are found, start the mDNS service.
		if len(ifaces) > 0 {
			mdnslogger.Info("starting mdns")
			if err := mdnsService.Start(); err != nil {
				if mdnsLocked {
					config.mdnsLockerDriver.Unlock()
				}
				return fmt.Errorf("mdnsService.Start error: %s", err.Error())
			}
		} else {
			mdnslogger.Error("unable to start mdns service, no multicast interfaces found")
		}
	}

	// 全局变量
	node = &Node{
		cuckoo:       cuckooNode,
		cuckooConfig: cuckooConf,
		nodeConfig:   config,
		mdnsService:  mdnsService,
	}

	return nil
}

func CloseNode() {
	if node != nil {
		if node.mdnsService != nil {
			node.mdnsService.Close()

			if node.nodeConfig.mdnsLockerDriver != nil {
				node.nodeConfig.mdnsLockerDriver.Unlock()
			}
		}
		node.mdnsService = nil

		if node.cuckoo != nil {
			node.cuckoo.Close()
		}
		node.cuckoo = nil

		node = nil
	}
}

func PrintAddr() (string, error) {

	var addr string
	if node != nil && node.cuckoo != nil {
		host, err := node.cuckoo.GetHost()
		if err == nil {
			addr += host.ID().String() + "\n"
			for _, v := range host.Addrs() {
				addr += "\n" + v.String()
			}
		}
	}

	return addr, nil
}
