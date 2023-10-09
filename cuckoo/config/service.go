package config

// 寄存消息服务配置
type DepositServiceConfig struct {
	EnableDepositService bool
}

// 文件服务配置
type FileServiceConfig struct {
	ResourceDir string // 资源目录
	FileDir     string // 上传目录
	TmpDir      string // 上传目录
	DownloadDir string // 下载目录
}
