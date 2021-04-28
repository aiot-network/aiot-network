package config

import (
	"github.com/aiot-network/aiotchain/common/config"
)

type Config struct {
	ConfigFile  string
	Format      bool
	TestNet     bool
	KeystoreDir string
	config.RpcConfig
}
