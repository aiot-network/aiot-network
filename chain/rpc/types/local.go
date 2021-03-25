package types

type Local struct {
	// Node version
	Version string `json:"version"`
	// Node network
	Network string `json:"network"`
	// Node p2p id
	Peer string `json:"peer"`
	// Node p2p address
	Address string `json:"address"`
	// Linked node
	Connections uint32 `json:"connections"`
	// Pool message
	Messages uint32 `json:"messages"`
	// Current block height
	Height uint64 `json:"height"`
	// Current effective block height
	Confirmed uint64 `json:"confirmed"`
}
