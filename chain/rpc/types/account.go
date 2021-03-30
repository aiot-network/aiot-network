package types

import (
	"github.com/aiot-network/aiot-network/chain/types"
	"github.com/aiot-network/aiot-network/tools/amount"
)

type Account struct {
	Address   string    `json:"address"`
	Nonce     uint64    `json:"nonce"`
	Tokens    Tokens    `json:"tokens"`
	Confirmed uint64    `json:"confirmed"`
	Works     *RpcWorks `json:"work"`
}

type RpcWorks struct {
	Cycle    uint64 `json:"cycle"`
	Workload uint64 `json:"workload"`
	EndTime  uint64 `json:"end"`
}

type TokenAccount struct {
	Address  string  `json:"address"`
	Balance  float64 `json:"balance"`
	LockedIn float64 `json:"locked"`
}

// List of secondary accounts
type Tokens []*TokenAccount

func ToRpcAccount(a *types.Account) *Account {
	tokens := make(Tokens, len(a.Tokens))
	for i, t := range a.Tokens {
		tokens[i] = &TokenAccount{
			Address:  t.Address,
			Balance:  amount.Amount(t.Balance).ToCoin(),
			LockedIn: amount.Amount(t.LockedIn).ToCoin(),
		}
	}
	return &Account{
		Address:   a.Address.String(),
		Nonce:     a.Nonce,
		Tokens:    tokens,
		Confirmed: a.Confirmed,
		Works: &RpcWorks{
			Cycle:    a.Works.Cycle,
			Workload: a.Works.WorkLoad,
			EndTime:  a.Works.EndTime,
		},
	}
}
