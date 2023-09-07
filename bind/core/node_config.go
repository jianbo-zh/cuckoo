package core

// Config is used in NewNode.
type NodeConfig struct {
	avatarDir        string
	storageDir       string
	netDriver        NativeNetDriver
	mdnsLockerDriver NativeMDNSLockerDriver
}

func NewNodeConfig(storageDir string, avatarDir string) *NodeConfig {
	return &NodeConfig{
		avatarDir:  avatarDir,
		storageDir: storageDir,
	}
}

func (c *NodeConfig) SetNetDriver(driver NativeNetDriver)         { c.netDriver = driver }
func (c *NodeConfig) SetMDNSLocker(driver NativeMDNSLockerDriver) { c.mdnsLockerDriver = driver }
