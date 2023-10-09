package core

// Config is used in NewNode.
type NodeConfig struct {
	dataDir          string
	resourceDir      string
	fileDir          string
	tmpDir           string
	netDriver        NativeNetDriver
	mdnsLockerDriver NativeMDNSLockerDriver
}

func NewNodeConfig(dataDir string, resourceDir string, fileDir string, tmpDir string) *NodeConfig {
	return &NodeConfig{
		dataDir:     dataDir,
		resourceDir: resourceDir,
		fileDir:     fileDir,
		tmpDir:      tmpDir,
	}
}

func (c *NodeConfig) SetNetDriver(driver NativeNetDriver)         { c.netDriver = driver }
func (c *NodeConfig) SetMDNSLocker(driver NativeMDNSLockerDriver) { c.mdnsLockerDriver = driver }
