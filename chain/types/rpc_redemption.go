package types

type RpcRedemptionBody struct {
	Address    string `json:"address"`
	Amount     uint64 `json:"amount"`
	PledgeRate int    `json:"pledgerate"`
}
