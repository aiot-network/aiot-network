package types

import (
	"github.com/aiot-network/aiotchain/chain/types"
	"github.com/aiot-network/aiotchain/tools/amount"
	types2 "github.com/aiot-network/aiotchain/types"
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

type RpcReword struct {
	Address string  `json:"address"`
	EndTime uint64  `json:"end"`
	Amount  float64 `json:"amount"`
	Cycle   uint64  `json:"cycle"`
	Blocks  uint64  `json:"blocks"`
	Workload uint64 `json:"workload"`
}

func ToRpcReword(reword []types2.IReword) []RpcReword {
	rpcReword := make([]RpcReword, 0)
	for _, r := range reword {
		rpcReword = append(rpcReword, RpcReword{
			Address: r.GetAddress(),
			EndTime: r.GetEndTime(),
			Amount:  amount.Amount(r.GetReword()).ToCoin(),
			Cycle:   r.GetCycle(),
			Blocks:  r.GetBlocks(),
			Workload: r.GetWorkLoad(),
		})
	}
	return rpcReword
}
