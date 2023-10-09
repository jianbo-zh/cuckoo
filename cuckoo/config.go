package cuckoo

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jianbo-zh/dchat/cuckoo/config"
)

func LoadConfig(dataDir string, resourceDir string, fileDir string, tmpDir string) (*config.Config, error) {

	cfile := filepath.Join(dataDir, config.DefaultConfigFile)

	var conf *config.Config

	bs, err := os.ReadFile(cfile)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("os.ReadFile error: %s", err.Error())
		}

		if dd, err := os.Stat(dataDir); os.IsNotExist(err) {
			if err = os.Mkdir(dataDir, 0755); err != nil {
				return nil, fmt.Errorf("os.Mkdir data error: %w", err)
			}
		} else if !dd.IsDir() {
			return nil, fmt.Errorf("data dir is not dir")
		}

		if dd, err := os.Stat(resourceDir); os.IsNotExist(err) {
			if err = os.Mkdir(resourceDir, 0755); err != nil {
				return nil, fmt.Errorf("os.Mkdir resource dir error: %w", err)
			}

		} else if !dd.IsDir() {
			return nil, fmt.Errorf("resource dir is not dir")
		}

		if dd, err := os.Stat(fileDir); os.IsNotExist(err) {
			if err = os.Mkdir(fileDir, 0755); err != nil {
				return nil, fmt.Errorf("os.Mkdir upload dir error: %w", err)
			}

		} else if !dd.IsDir() {
			return nil, fmt.Errorf("upload dir is not dir")
		}

		// config file not exists
		// use default config
		conf, err = config.DefaultConfig()
		if err != nil {
			return nil, fmt.Errorf("config.DefaultConfig error: %s", err.Error())
		}

		conf.DataDir = dataDir
		conf.FileService.ResourceDir = resourceDir // 资源文件保存目录（头像及语音内容）
		conf.FileService.FileDir = fileDir         // 上传文件保存目录（发送文件前需要copy到上传目录，其他人下载是从下载目录上传）
		conf.FileService.TmpDir = tmpDir           // 临时文件目录
		conf.FileService.DownloadDir = ""          // 指定以后下载时，再指定外部目录

		// marshal for write to file
		bs, err = json.MarshalIndent(conf, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("json.Marshal conf error: %s", err.Error())
		}

		// wirte config to file for storage
		if err = os.WriteFile(cfile, bs, 0755); err != nil {
			return nil, fmt.Errorf("os.WriteFile conf error: %s", err.Error())
		}
	}

	// unmarshall config return
	if err = json.Unmarshal(bs, &conf); err != nil {
		return nil, fmt.Errorf("json.Unmarshal error: %s", err.Error())
	}

	return conf, nil
}
