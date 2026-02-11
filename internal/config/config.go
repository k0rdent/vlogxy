package config

import (
	"fmt"
	"os"
	"sync"

	"github.com/goccy/go-yaml"
	"github.com/k0rdent/victorialogs-aggregator/internal/interfaces"
)

// Config implements ConfigProvider interface
type Config struct {
	data  *ConfigData
	path  string
	mutex sync.RWMutex
}

type Group struct {
	Target string `yaml:"target"`
	Scheme string `yaml:"scheme"`
}

type ConfigData struct {
	ServerGroups []*Group `yaml:"server_groups"`
}

// LoadConfig loads configuration from the specified path
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
		mutex: sync.RWMutex{},
	}

	return config, nil
}

// Reload reloads configuration from the source file
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

// GetServerGroups returns the list of configured server groups
func (c *Config) GetServerGroups() []interfaces.ServerGroup {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	groups := make([]interfaces.ServerGroup, 0, len(c.data.ServerGroups))
	for _, g := range c.data.ServerGroups {
		groups = append(groups, interfaces.ServerGroup{
			Target: g.Target,
			Scheme: g.Scheme,
		})
	}
	return groups
}

// GetData returns a copy of the configuration data (for backward compatibility)
func (c *Config) GetData() ConfigData {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
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
