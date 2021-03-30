package config

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/aiot-network/aiot-network/common/param"
	"github.com/aiot-network/aiot-network/common/private"
	log2 "github.com/aiot-network/aiot-network/tools/log"
	log "github.com/aiot-network/aiot-network/tools/log/log15"
	"github.com/aiot-network/aiot-network/tools/utils"
	"github.com/aiot-network/aiot-network/version"
	"github.com/jessevdk/go-flags"
	"os"
	"path/filepath"
	"strings"
)

var DefaultHomeDir string
var Param *param.Param

// Config is the node startup parameter
type Config struct {
	ConfigFile string `long:"config" description:"Start with a configuration file"`
	Data       string `long:"data" description:"Path to application data directory"`
	Logging    bool   `long:"logging" description:"Logging switch"`
	ExternalIp string `long:"externalip" description:"External network IP address"`
	Boot       string `long:"boot" description:"Custom boot"`
	P2PPort    string `long:"p2pport" description:"Add an interface/port to listen for connections"`
	RpcPort    string `long:"rpcport" description:"Add an interface/port to listen for RPC connections"`
	RpcTLS     bool   `long:"rpctls" description:"Open TLS for the RPC server -- NOTE: This is only allowed if the RPC server is bound to localhost"`
	RpcCert    string `long:"rpccert" description:"File containing the certificate file"`
	RpcKey     string `long:"rpckey" description:"File containing the certificate key"`
	RpcPass    string `long:"rpcpass" description:"Password for RPC connections"`
	TestNet    bool   `long:"testnet" description:"Use the test network"`
	KeyFile    string `long:"keyfile" description:"If you participate in mining, you need to configure the mining address key file"`
	KeyPass    string `long:"keypass" description:"The decryption password for key file"`
	RollBack   uint64 `long:"rollback" description:"Roll back to the previous height"`
	Version    bool   `long:"version" description:"View Version number"`
	Private    private.IPrivate
}

// LoadConfig load the parse node startup parameter
func LoadParam(private private.IPrivate) error {
	cfg := &Config{}
	appName := filepath.Base(os.Args[0])
	appName = strings.TrimSuffix(appName, filepath.Ext(appName))
	preParser := newConfigParser(cfg, flags.HelpFlag)
	_, err := preParser.Parse()
	if err != nil {
		if e, ok := err.(*flags.Error); ok && e.Type != flags.ErrHelp {
			return err
		} else if ok && e.Type == flags.ErrHelp {
			return err
		}
	}

	if cfg.ConfigFile != "" {
		_, err = toml.DecodeFile(cfg.ConfigFile, cfg)
		if err != nil {
			return err
		}
	}

	if cfg.TestNet {
		Param = param.TestNetParam
	} else {
		Param = param.MainNetParam
	}

	if cfg.Version {
		fmt.Println(version.StringifySingleLine(Param.App))
		os.Exit(0)
	}

	if cfg.Data != "" {
		Param.Data = cfg.Data
	} else {
		Param.Data = utils.AppDataDir(Param.App, false)
	}
	// p2p service listening port, if not, use the default port
	if cfg.P2PPort != "" {
		Param.P2pParam.P2pPort = cfg.P2PPort
	}
	if cfg.Boot != "" {
		Param.P2pParam.CustomBoot = cfg.Boot
	}

	// Set the default external IP. If the external IP is not set,
	// other nodes can only know you but cannot send messages to you.
	if cfg.ExternalIp != "" {
		Param.P2pParam.ExternalIp = cfg.ExternalIp
	}

	if cfg.RpcPort != "" {
		Param.RpcParam.RpcPort = cfg.RpcPort
	}
	if cfg.RpcPass != "" {
		Param.RpcParam.RpcPass = cfg.RpcPass
	}
	if cfg.RpcCert != "" {
		Param.RpcParam.RpcCert = cfg.RpcCert
	}
	if cfg.RpcTLS {
		Param.RpcParam.RpcTLS = cfg.RpcTLS
	}
	if cfg.RpcCert != "" {
		Param.RpcParam.RpcCertKey = cfg.RpcCert
	}
	if cfg.KeyFile != "" {
		Param.PrivateFile = cfg.KeyFile
	}
	if cfg.KeyPass != "" {
		Param.PrivatePass = cfg.KeyPass
	}
	if cfg.RollBack != 0 {
		Param.RollBack = cfg.RollBack
	}

	if !utils.Exist(Param.Data) {
		if err := os.Mkdir(Param.Data, os.ModePerm); err != nil {
			return err
		}
	}
	Param.Data = Param.Data + "/" + Param.P2pParam.P2pPort
	if !utils.Exist(Param.Data) {
		if err := os.Mkdir(Param.Data, os.ModePerm); err != nil {
			return err
		}
	}

	// Each node requires a secp256k1 private key, which is used as the p2p id
	// generation and signature of the node that generates the block.
	// If this parameter is not configured in the startup parameter,
	// the node will be automatically generated and loaded automatically at startup
	Param.IPrivate = private
	if cfg.KeyFile == "" {
		Param.PrivateFile = Param.Data + "/" + Param.PrivateFile
		if err := Param.IPrivate.Load(Param.PrivateFile, Param.PrivatePass); err != nil {
			if err = Param.IPrivate.Create(Param.Name, Param.PrivateFile, Param.PrivatePass); err != nil {
				return fmt.Errorf("create default priavte failed! %s", err.Error())
			}
		}
	} else {
		if Param.PrivatePass == "" {
			fmt.Println("Please enter the password for the key file:")
			passWd, err := utils.ReadPassWd()
			if err != nil {
				return fmt.Errorf("read password failed! %s", err.Error())
			}
			Param.PrivatePass = string(passWd)
		}
		if err := Param.IPrivate.Load(Param.PrivateFile, Param.PrivatePass); err != nil {
			return fmt.Errorf("load private failed! %s", err.Error())
		}
	}

	// If this parameter is true, the log is also written to the file
	if cfg.Logging != Param.Logging {
		Param.Logging = cfg.Logging
	}
	if Param.Logging {
		logDir := Param.Data + "/log"
		if !utils.Exist(logDir) {
			if err := os.Mkdir(logDir, os.ModePerm); err != nil {
				return err
			}
		}
		utils.CleanAndExpandPath(logDir)
		logDir = filepath.Join(logDir, Param.Name)
		log2.InitLogRotator(filepath.Join(logDir, "future_mine.log"))
	}
	log.Info("Data storage directory", "module", "config", "path", Param.Data)
	return nil
}

func newConfigParser(cfg *Config, options flags.Options) *flags.Parser {
	parser := flags.NewParser(cfg, options)
	return parser
}
