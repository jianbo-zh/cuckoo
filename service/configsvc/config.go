package configsvc

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/jianbo-zh/dchat/cuckoo/config"
	"github.com/jianbo-zh/dchat/internal/myevent"
	logging "github.com/jianbo-zh/go-log"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
)

var log = logging.Logger("configsvc")

var _ ConfigServiceIface = (*ConfigSvc)(nil)

type ConfigSvc struct {
	mutex sync.Mutex
	conf  config.Config

	emitters struct {
		evtConfigBootstrapsChange           event.Emitter
		evtConfigPeeringPeersChange         event.Emitter
		evtConfigEnableMDNSChange           event.Emitter
		evtConfigEnableDepositServiceChange event.Emitter
		evtConfigDownloadDirChange          event.Emitter
	}
}

func NewConfigService(ctx context.Context, conf *config.Config, ebus event.Bus) (*ConfigSvc, error) {
	var err error

	confSvc := ConfigSvc{
		conf: *conf,
	}

	if confSvc.emitters.evtConfigBootstrapsChange, err = ebus.Emitter(&myevent.EvtConfigBootstrapsChange{}); err != nil {
		return nil, fmt.Errorf("set bootstraps config change emitter error: %w", err)
	}

	if confSvc.emitters.evtConfigPeeringPeersChange, err = ebus.Emitter(&myevent.EvtConfigPeeringPeersChange{}); err != nil {
		return nil, fmt.Errorf("set peering peers config change emitter error: %w", err)
	}

	if confSvc.emitters.evtConfigEnableMDNSChange, err = ebus.Emitter(&myevent.EvtConfigEnableMDNSChange{}); err != nil {
		return nil, fmt.Errorf("set enable mdns config change emitter error: %w", err)
	}

	if confSvc.emitters.evtConfigEnableDepositServiceChange, err = ebus.Emitter(&myevent.EvtConfigEnableDepositServiceChange{}); err != nil {
		return nil, fmt.Errorf("set enable deposit service config change emitter error: %w", err)
	}

	if confSvc.emitters.evtConfigDownloadDirChange, err = ebus.Emitter(&myevent.EvtConfigDownloadDirChange{}); err != nil {
		return nil, fmt.Errorf("set download dir config change emitter error: %w", err)
	}

	return &confSvc, nil
}

// GetConfig 获取配置
func (c *ConfigSvc) GetConfig() (config.Config, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.conf, nil
}

// SetBootstraps 设置引导节点
func (c *ConfigSvc) SetBootstraps(bootstraps []string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	var maddrs []string
	for _, addr := range bootstraps {
		maddr, err := ma.NewMultiaddr(addr)
		if err != nil {
			return fmt.Errorf("bootstrap address error: %w", err)
		}

		maddrs = append(maddrs, maddr.String())
	}

	c.conf.Bootstrap = maddrs

	if err := c.syncConfigToFile(); err != nil {
		return fmt.Errorf("sync config to file error: %w", err)
	}

	if err := c.emitters.evtConfigBootstrapsChange.Emit(myevent.EvtConfigBootstrapsChange{
		Bootstraps: maddrs,
	}); err != nil {
		return fmt.Errorf("emit bootstrap config event error: %w", err)
	}

	return nil
}

// SetPeeringPeers 设置中继服务器地址（公网地址）
func (c *ConfigSvc) SetPeeringPeers(peers []string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	var painfos []peer.AddrInfo
	for _, addr := range peers {
		painfo, err := peer.AddrInfoFromString(addr)
		if err != nil {
			return fmt.Errorf("peering peers address error: %w", err)
		}

		painfos = append(painfos, *painfo)
	}

	c.conf.Peering.Peers = painfos

	if err := c.syncConfigToFile(); err != nil {
		return fmt.Errorf("sync config to file error: %w", err)
	}

	if err := c.emitters.evtConfigPeeringPeersChange.Emit(myevent.EvtConfigPeeringPeersChange{
		PeeringPeers: painfos,
	}); err != nil {
		return fmt.Errorf("emit peering peers config event error: %w", err)
	}

	return nil
}

// SetEnableMDNS 设置是否启动 MDNS
func (c *ConfigSvc) SetEnableMDNS(isEnable bool) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.conf.Discovery.MDNS.Enabled == isEnable {
		return nil
	}

	c.conf.Discovery.MDNS.Enabled = isEnable

	if err := c.syncConfigToFile(); err != nil {
		return fmt.Errorf("sync config to file error: %w", err)
	}

	if err := c.emitters.evtConfigEnableMDNSChange.Emit(myevent.EvtConfigEnableMDNSChange{
		Enable: isEnable,
	}); err != nil {
		return fmt.Errorf("emit enable mdns config event error: %w", err)
	}

	return nil
}

// SetEnableDepositService 设置是否启动寄存服务
func (c *ConfigSvc) SetEnableDepositService(isEnable bool) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// 判断更新条件
	if c.conf.DepositService.EnableDepositService == isEnable {
		return nil
	}

	// 更新配置
	c.conf.DepositService.EnableDepositService = isEnable
	if err := c.syncConfigToFile(); err != nil {
		return fmt.Errorf("sync config to file error: %w", err)
	}

	// 触发配置改变事件
	if err := c.emitters.evtConfigEnableDepositServiceChange.Emit(myevent.EvtConfigEnableDepositServiceChange{
		Enable: isEnable,
	}); err != nil {
		return fmt.Errorf("emit enable deposit service config event error: %w", err)
	}

	return nil
}

// SetFileDownloadDir 设置下载目录
func (c *ConfigSvc) SetFileDownloadDir(downloadDir string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.conf.FileService.DownloadDir == downloadDir {
		return nil
	}

	c.conf.FileService.DownloadDir = downloadDir

	if err := c.syncConfigToFile(); err != nil {
		return fmt.Errorf("sync config to file error: %w", err)
	}

	if err := c.emitters.evtConfigDownloadDirChange.Emit(myevent.EvtConfigDownloadDirChange{
		DownloadDir: downloadDir,
	}); err != nil {
		return fmt.Errorf("emit download dir config event error: %w", err)
	}

	return nil
}

// Close 关闭服务
func (c *ConfigSvc) Close() {}

// syncConfig 同步配置
func (c *ConfigSvc) syncConfigToFile() error {
	configPath := filepath.Join(c.conf.DataDir, config.DefaultConfigFile)

	bs, err := json.MarshalIndent(c.conf, "", "  ")
	if err != nil {
		return fmt.Errorf("json marshal conf error: %s", err.Error())
	}

	if err = os.WriteFile(configPath, bs, 0755); err != nil {
		return fmt.Errorf("os write conf file error: %s", err.Error())
	}

	return nil
}
