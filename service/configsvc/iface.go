package configsvc

import (
	"github.com/jianbo-zh/dchat/cuckoo/config"
)

type ConfigServiceIface interface {
	GetConfig() (config.Config, error)
	SetBootstraps(maddr []string) error
	SetPeeringPeers(peers []string) error
	SetEnableMDNS(isEnable bool) error
	SetEnableDepositService(isEnable bool) error
	SetFileDownloadDir(dir string) error
	Close()
}
