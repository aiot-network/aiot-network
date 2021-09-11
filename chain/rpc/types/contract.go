package types

import (
	"github.com/aiot-network/aiotchain/chain/types/status"
	"github.com/aiot-network/aiotchain/chain/types/status/exchange"
	"github.com/aiot-network/aiotchain/tools/amount"
)

type RpcPair struct {
	Address              string  `json:"address"`
	Exchange             string  `json:"exchange"`
	Symbol               string  `json:"symbol"`
	Token0               string  `json:"token0"`
	Token1               string  `json:"token1"`
	Symbol0              string  `json:"symbol0"`
	Symbol1              string  `json:"symbol1"`
	Reserve0             float64 `json:"reserve0"`
	Reserve1             float64 `json:"reserve1"`
	TotalSupply          float64 `json:"totalSupply"`
	BlockTimestampLast   uint32  `json:"blockTimestampLast"`
	Price0CumulativeLast uint64  `json:"price0CumulativeLast"`
	Price1CumulativeLast uint64  `json:"price1CumulativeLast"`
	KLast                string  `json:"kLast"`
}

type RpcExchange struct {
	Address string                       `json:"address"`
	Symbol  string                       `json:"symbol"`
	FeeTo   string                       `json:"feeTo"`
	Admin   string                       `json:"admin"`
	Pair    map[string]map[string]string `json:"pair"`
}

type PairAddress struct {
	Key     string `json:"key"`
	Address string `json:"address"`
}

func TranslateContractToRpcContract(contract *status.Contract) interface{} {
	switch contract.Type {
	case status.Exchange_:
		exchange, _ := contract.Body.(*exchange.Exchange)
		pair := map[string]map[string]string{}
		for token0, token1AndAddr := range exchange.Pair {
			addressMap := map[string]string{}
			for token1, address := range token1AndAddr {
				addressMap[token1.String()] = address.String()
			}
			pair[token0.String()] = addressMap
		}
		return &RpcExchange{
			Address: contract.Address.String(),
			FeeTo:   exchange.FeeTo.String(),
			Admin:   exchange.Admin.String(),
			Symbol:  exchange.Symbol,
			Pair:    pair,
		}
	case status.Pair_:
		pair, _ := contract.Body.(*exchange.Pair)
		return &RpcPair{
			Address:              contract.Address.String(),
			Exchange:             pair.Exchange.String(),
			Symbol:               pair.Symbol,
			Token0:               pair.Token0.String(),
			Token1:               pair.Token1.String(),
			Symbol0:              pair.Symbol0,
			Symbol1:              pair.Symbol1,
			Reserve0:             amount.Amount(pair.Reserve0).ToCoin(),
			Reserve1:             amount.Amount(pair.Reserve1).ToCoin(),
			BlockTimestampLast:   pair.BlockTimestampLast,
			Price0CumulativeLast: pair.Price0CumulativeLast,
			Price1CumulativeLast: pair.Price1CumulativeLast,
			KLast:                pair.KLast.String(),
			TotalSupply:          amount.Amount(pair.TotalSupply).ToCoin(),
		}
	}
	return nil
}
