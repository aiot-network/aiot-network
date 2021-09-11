package types

import (
	"github.com/aiot-network/aiotchain/chain/types/status"
)

type RpcContractV2TransactionBody struct {
	Contract     string              `json:"contract"`
	Type         status.ContractType `json:"type"`
	FunctionType status.FunctionType `json:"functiontype"`
	Function     IRCFunction         `json:"function"`
}

type RpcContractV2BodyWithState struct {
	Contract     string              `json:"contract"`
	Type         status.ContractType `json:"type"`
	FunctionType status.FunctionType `json:"functiontype"`
	Function     IRCFunction         `json:"function"`
	State        *RpcContractState   `json:"state"`
}

type RpcContractState struct {
	StateCode ContractStatus `json:"statecode"`
	Events    []*RpcEvent    `json:"event"`
	Error     string         `json:"error"`
}

type RpcEvent struct {
	EventType int    `json:"eventtype"`
	From      string `json:"from"`
	To        string `json:"to"`
	Token     string `json:"token"`
	Amount    uint64 `json:"amount"`
	Height    uint64 `json:"height"`
}

type IRCFunction interface {
}

type RpcExchangeInitBody struct {
	Symbol string `json:"symbol"`
	Admin  string `json:"admin"`
	FeeTo  string `json:"feeto"`
}

type RpcExchangeSetAdminBody struct {
	Address string `json:"address"`
}

type RpcExchangeSetFeeToBody struct {
	Address string `json:"address"`
}

type RpcExchangeExactInBody struct {
	AmountIn     uint64   `json:"amountin"`
	AmountOutMin uint64   `json:"amountoutmin"`
	Path         []string `json:"path"`
	To           string   `json:"to"`
	Deadline     uint64   `json:"deadline"`
}

type RpcExchangeExactOutBody struct {
	AmountOut   uint64   `json:"amountout"`
	AmountInMax uint64   `json:"amountinmax"`
	Path        []string `json:"path"`
	To          string   `json:"to"`
	Deadline    uint64   `json:"deadline"`
}

type RpcExchangeAddLiquidity struct {
	Exchange       string `json:"exchange"`
	TokenA         string `json:"tokenA"`
	TokenB         string `json:"tokenB"`
	To             string `json:"to"`
	AmountADesired uint64 `json:"amountadesired"`
	AmountBDesired uint64 `json:"amountbdesired"`
	AmountAMin     uint64 `json:"amountamin"`
	AmountBMin     uint64 `json:"amountbmin"`
	Deadline       uint64 `json:"deadline"`
}

type RpcExchangeRemoveLiquidity struct {
	Exchange   string `json:"exchange"`
	TokenA     string `json:"tokenA"`
	TokenB     string `json:"tokenB"`
	To         string `json:"to"`
	Liquidity  uint64 `json:"liquidity"`
	AmountAMin uint64 `json:"amountamin"`
	AmountBMin uint64 `json:"amountbmin"`
	Deadline   uint64 `json:"deadline"`
}
