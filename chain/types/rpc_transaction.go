package types

type RpcReceiver struct {
	Address string `json:"address"`
	Amount  uint64 `json:"amount"`
}

type RpcTransactionBody struct {
	Token     string         `json:"token"`
	Receivers []*RpcReceiver `json:"receivers"`
}
