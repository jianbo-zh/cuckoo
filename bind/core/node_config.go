package core

// Config is used in NewNode.
type NodeConfig struct {
	netDriver        NativeNetDriver
	mdnsLockerDriver NativeMDNSLockerDriver
}

func NewNodeConfig() *NodeConfig {
	return &NodeConfig{}
}

func (c *NodeConfig) SetNetDriver(driver NativeNetDriver)         { c.netDriver = driver }
func (c *NodeConfig) SetMDNSLocker(driver NativeMDNSLockerDriver) { c.mdnsLockerDriver = driver }
