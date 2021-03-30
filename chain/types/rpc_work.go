package types

type RpcWorkBody struct {
	StartTime uint64           `json:"start"`
	EndTime   uint64           `json:"end"`
	List      []RpcAddressWork `json:"list"`
}

type RpcAddressWork struct {
	Address  string `json:"address"`
	Workload uint64 `json:"workload"`
	EndTime  uint64 `json:"end"`
}
