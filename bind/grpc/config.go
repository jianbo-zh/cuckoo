package service

import (
	"context"
	"fmt"

	"github.com/jianbo-zh/dchat/bind/grpc/proto"
	"github.com/jianbo-zh/dchat/cuckoo"
	"github.com/jianbo-zh/dchat/service/configsvc"
)

var _ proto.ConfigSvcServer = (*ConfigSvc)(nil)

type ConfigSvc struct {
	getter cuckoo.CuckooGetter
	proto.UnimplementedConfigSvcServer
}

func NewConfigSvc(getter cuckoo.CuckooGetter) *ConfigSvc {
	return &ConfigSvc{
		getter: getter,
	}
}

func (c *ConfigSvc) getConfigSvc() (configsvc.ConfigServiceIface, error) {
	cuckoo, err := c.getter.GetCuckoo()
	if err != nil {
		return nil, fmt.Errorf("get cuckoo error: %s", err.Error())
	}

	configSvc, err := cuckoo.GetConfigSvc()
	if err != nil {
		return nil, fmt.Errorf("cuckoo get deposit svc error: %s", err.Error())
	}

	return configSvc, nil
}

func (c *ConfigSvc) GetConfig(ctx context.Context, request *proto.GetConfigRequest) (reply *proto.GetConfigReply, err error) {

	configSvc, err := c.getConfigSvc()
	if err != nil {
		return nil, fmt.Errorf("get config svc error: %w", err)
	}

	config, err := configSvc.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("svc get config error: %w", err)
	}

	var depositServiceAddress string
	if config.DepositService.EnableDepositService {
		depositServiceAddress = config.Identity.PeerID
	}

	reply = &proto.GetConfigReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		Config: &proto.Config{
			EnableDepositService:  config.DepositService.EnableDepositService,
			DepositServiceAddress: depositServiceAddress,
		},
	}
	return reply, nil
}

func (c *ConfigSvc) SetBootstraps(ctx context.Context, request *proto.SetBootstrapsRequest) (reply *proto.SetBootstrapsReply, err error) {
	configSvc, err := c.getConfigSvc()
	if err != nil {
		return nil, fmt.Errorf("get config svc error: %w", err)
	}

	if err = configSvc.SetBootstraps(request.Bootstraps); err != nil {
		return nil, fmt.Errorf("svc set bootstraps error: %w", err)
	}

	reply = &proto.SetBootstrapsReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
	}
	return reply, nil
}

func (c *ConfigSvc) SetPeeringPeers(ctx context.Context, request *proto.SetPeeringPeersRequest) (reply *proto.SetPeeringPeersReply, err error) {
	configSvc, err := c.getConfigSvc()
	if err != nil {
		return nil, fmt.Errorf("get config svc error: %w", err)
	}

	if err = configSvc.SetPeeringPeers(request.PeeringPeers); err != nil {
		return nil, fmt.Errorf("svc set peering peers error: %w", err)
	}

	reply = &proto.SetPeeringPeersReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
	}
	return reply, nil
}

func (c *ConfigSvc) SetEnableMDNS(ctx context.Context, request *proto.SetEnableMDNSRequest) (reply *proto.SetEnableMDNSReply, err error) {
	configSvc, err := c.getConfigSvc()
	if err != nil {
		return nil, fmt.Errorf("get config svc error: %w", err)
	}

	if err = configSvc.SetEnableMDNS(request.Enable); err != nil {
		return nil, fmt.Errorf("svc set enable mdns error: %w", err)
	}
	reply = &proto.SetEnableMDNSReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
	}
	return reply, nil
}

func (c *ConfigSvc) SetEnableDepositService(ctx context.Context, request *proto.SetEnableDepositServiceRequest) (reply *proto.SetEnableDepositServiceReply, err error) {

	log.Infoln("SetEnableDepositService request: ", request.String())
	defer func() {
		if e := recover(); e != nil {
			log.Panicln("SetEnableDepositService panic: ", e)
		} else if err != nil {
			log.Errorln("SetEnableDepositService error: ", err.Error())
		} else {
			log.Infoln("SetEnableDepositService reply: ", reply.String())
		}
	}()

	configSvc, err := c.getConfigSvc()
	if err != nil {
		return nil, fmt.Errorf("get config svc error: %w", err)
	}

	if err = configSvc.SetEnableDepositService(request.Enable); err != nil {
		return nil, fmt.Errorf("svc set enable deposit service error: %w", err)
	}

	config, err := configSvc.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("svc get config error: %w", err)
	}

	var depositServiceAddress string
	if config.DepositService.EnableDepositService {
		depositServiceAddress = config.Identity.PeerID
	}

	reply = &proto.SetEnableDepositServiceReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		EnableDepositService:  config.DepositService.EnableDepositService,
		DepositServiceAddress: depositServiceAddress,
	}
	return reply, nil
}

func (c *ConfigSvc) SetDownloadDir(ctx context.Context, request *proto.SetDownloadDirRequest) (reply *proto.SetDownloadDirReply, err error) {
	configSvc, err := c.getConfigSvc()
	if err != nil {
		return nil, fmt.Errorf("get config svc error: %w", err)
	}

	if err = configSvc.SetFileDownloadDir(request.DownloadDir); err != nil {
		return nil, fmt.Errorf("svc set download dir error: %w", err)
	}

	reply = &proto.SetDownloadDirReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
	}
	return reply, nil
}
