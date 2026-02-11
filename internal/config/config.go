package config

import (
	"fmt"
	"os"
	"sync"

	"github.com/goccy/go-yaml"
)

type Config struct {
	data  *ConfigData
	path  string
	mutex sync.Mutex
}

type Group struct {
	Target string `yaml:"target"`
	Scheme string `yaml:"scheme"`
}

type ConfigData struct {
	ServerGroups []*Group `yaml:"server_groups"`
}

var GlobalConfig *Config

func LoadConfig(path string) (*Config, error) {
	if path == "" {
		return nil, fmt.Errorf("config path cannot be empty")
	}

	configData, err := readConfigData(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	config := &Config{
		data:  configData,
		path:  path,
		mutex: sync.Mutex{},
	}

	return config, nil
}

func (c *Config) Reload() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	configData, err := readConfigData(c.path)
	if err != nil {
		return fmt.Errorf("failed to reload config: %w", err)
	}
	c.data = configData

	return nil
}

func (c *Config) GetData() ConfigData {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return *c.data
}

func readConfigData(path string) (*ConfigData, error) {
	rawConfig, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	config, err := parseConfig(rawConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return config, nil
}

func parseConfig(content []byte) (*ConfigData, error) {
	serverGroups := new(ConfigData)
	err := yaml.Unmarshal(content, serverGroups)
	return serverGroups, err
}
