package cuckoo

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jianbo-zh/dchat/cuckoo/config"
)

func LoadConfig(storageDir string, avatarDir string) (*config.Config, error) {

	cfile := filepath.Join(storageDir, config.DefaultConfigFile)

	var conf *config.Config

	bs, err := os.ReadFile(cfile)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("os.ReadFile conf error: %s", err.Error())
		}

		// config file not exists
		// use default config
		conf, err = config.DefaultConfig()
		if err != nil {
			return nil, fmt.Errorf("config.DefaultConfig error: %s", err.Error())
		}
		conf.StorageDir = storageDir
		conf.AvatarDir = avatarDir

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
