package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mitchellh/go-homedir"
)

const (
	// DefaultPathName is the default config dir name
	DefaultPathName = ".dchat"
	// DefaultPathRoot is the path to the default config dir location.
	DefaultPathRoot = "~/" + DefaultPathName
	// DefaultConfigFile is the filename of the configuration file
	DefaultConfigFile = "config"
	// EnvDir is the environment variable used to change the path root.
	EnvDir = "DCHAT_PATH"
)

type Config struct {
	Addresses Addresses // local node's addresses
	Identity  Identity  // local node's peer identity
	Bootstrap []string  // local nodes's bootstrap peer addresses
	Discovery Discovery // local node's discovery mechanisms
	Ipns      Ipns      // Ipns settings
	Swarm     SwarmConfig
	AutoNAT   AutoNATConfig
	Peering   Peering
	DNS       DNS

	StorageDir string

	AccountService AccountServiceConfig
	ContactService ContactServiceConfig
	GroupService   GroupServiceConfig
	DepositService DepositServiceConfig
	FileService    FileServiceConfig
}

// Clone copies the config. Use when updating.
func (c *Config) Clone() (*Config, error) {
	var newConfig Config
	var buf bytes.Buffer

	if err := json.NewEncoder(&buf).Encode(c); err != nil {
		return nil, fmt.Errorf("failure to encode config: %s", err)
	}

	if err := json.NewDecoder(&buf).Decode(&newConfig); err != nil {
		return nil, fmt.Errorf("failure to decode config: %s", err)
	}

	return &newConfig, nil
}

// PathRoot returns the default configuration root directory
func PathRoot() (string, error) {
	dir := os.Getenv(EnvDir)
	if len(dir) == 0 {
		return homedir.Expand(DefaultPathRoot)
	}
	return homedir.Expand(dir)
}

// Path returns the path `extension` relative to the configuration root. If an
// empty string is provided for `configroot`, the default root is used.
func Path(configroot, extension string) (string, error) {
	if len(configroot) == 0 {
		dir, err := PathRoot()
		if err != nil {
			return "", err
		}
		return filepath.Join(dir, extension), nil
	}

	dir, err := homedir.Expand(configroot)
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, extension), nil
}

// Filename returns the configuration file path given a configuration root
// directory and a user-provided configuration file path argument with the
// following rules:
//   - If the user-provided configuration file path is empty, use the default one.
//   - If the configuration root directory is empty, use the default one.
//   - If the user-provided configuration file path is only a file name, use the
//     configuration root directory, otherwise use only the user-provided path
//     and ignore the configuration root.
func Filename(configroot string, userConfigFile string) (string, error) {
	if userConfigFile == "" {
		return Path(configroot, DefaultConfigFile)
	}

	if filepath.Dir(userConfigFile) == "." {
		return Path(configroot, userConfigFile)
	}

	return userConfigFile, nil
}

// HumanOutput gets a config value ready for printing
func HumanOutput(value interface{}) ([]byte, error) {
	s, ok := value.(string)
	if ok {
		return []byte(strings.Trim(s, "\n")), nil
	}
	return Marshal(value)
}

// Marshal configuration with JSON
func Marshal(value interface{}) ([]byte, error) {
	// need to prettyprint, hence MarshalIndent, instead of Encoder
	return json.MarshalIndent(value, "", "  ")
}

func FromMap(v map[string]interface{}) (*Config, error) {
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(v); err != nil {
		return nil, err
	}
	var conf Config
	if err := json.NewDecoder(buf).Decode(&conf); err != nil {
		return nil, fmt.Errorf("failure to decode config: %s", err)
	}
	return &conf, nil
}

func ToMap(conf *Config) (map[string]interface{}, error) {
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(conf); err != nil {
		return nil, err
	}
	var m map[string]interface{}
	if err := json.NewDecoder(buf).Decode(&m); err != nil {
		return nil, fmt.Errorf("failure to decode config: %s", err)
	}
	return m, nil
}
