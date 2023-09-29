package config

// 寄存消息服务配置
type DepositServiceConfig struct {
	EnableDepositService bool
}

// 文件服务配置
type FileServiceConfig struct {
	ResourceDir string
	DownloadDir string
}
