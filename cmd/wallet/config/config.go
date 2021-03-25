package config

import (
	"github.com/aiot-network/aiot-network/common/config"
)

type Config struct {
	ConfigFile  string
	Format      bool
	TestNet     bool
	KeystoreDir string
	config.RpcConfig
}
