package config

import (
	"github.com/k0rdent/vlogxy/internal/interfaces"
	servergroup "github.com/k0rdent/vlogxy/internal/server-group"
)

func GetFakeConfig(serverGroups []*servergroup.Server, maxLogsLimit int) interfaces.ConfigProvider {
	return &Config{
		data: &ConfigData{
			ServerGroups: serverGroups,
		},
		maxLogsLimit: maxLogsLimit,
	}
}
