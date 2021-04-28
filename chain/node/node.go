package node

import (
	"encoding/json"
	"github.com/aiot-network/aiotchain/common/config"
	"github.com/aiot-network/aiotchain/server"
	log "github.com/aiot-network/aiotchain/tools/log/log15"
	"github.com/aiot-network/aiotchain/types"
	"github.com/aiot-network/aiotchain/version"
)

const module = "aiot_node"

type Node struct {
	services []server.IService
}

func NewNode() *Node {
	return &Node{
		services: make([]server.IService, 0),
	}
}

func (n *Node) Start() error {
	if err := n.startServices(); err != nil {
		return err
	}
	return nil
}

func (n *Node) Stop() error {
	for _, s := range n.services {
		if err := s.Stop(); err != nil {
			log.Error("Service failed to stop", "module", module, "service", s.Name(), "error", err.Error())
		}
	}

	return nil
}

func (n *Node) Register(s server.IService) {
	n.services = append(n.services, s)
}

func (n *Node) LocalInfo() *types.Local {
	all := make(map[string]interface{})
	for _, s := range n.services {
		infoMap := s.Info()
		for name, value := range infoMap {
			all[name] = value
		}
	}
	all["version"] = version.StringifySingleLine(config.Param.App)
	all["network"] = config.Param.Name
	bytes, err := json.Marshal(all)
	if err != nil {
		return &types.Local{}
	}
	var rs *types.Local
	err = json.Unmarshal(bytes, &rs)
	if err != nil {
		return &types.Local{}
	}
	return rs
}

func (n *Node) startServices() error {
	for _, s := range n.services {
		if err := s.Start(); err != nil {
			log.Error("Service failed to start", "module", module, "service", s.Name(), "error", err.Error())
		}
	}
	return nil
}
