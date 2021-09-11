package message

import (
	"github.com/aiot-network/aiotchain/chain/runner/exchange_runner"
	"github.com/aiot-network/aiotchain/chain/types"
	"github.com/aiot-network/aiotchain/chain/types/functionbody/exchange_func"
	"github.com/aiot-network/aiotchain/chain/types/status"
	"github.com/aiot-network/aiotchain/tools/arry"
	"time"
)

func NewExchange(net, from, admin, feeTo, symbol string, nonce uint64) (*types.Message, error) {
	contract, err := exchange_runner.ExchangeAddress(net, from, nonce)
	if err != nil {
		return nil, err
	}

	msg := &types.Message{
		Header: &types.MsgHeader{
			Type:  types.Contract,
			Hash:  arry.Hash{},
			From:  arry.StringToAddress(from),
			Nonce: nonce,
			Fee:   types.ContractFees,
			Time:  uint64(time.Now().Unix()),
			Signature: &types.Signature{
				Bytes:  nil,
				PubKey: nil,
			},
		},
		Body: &types.ContractBody{
			Contract:     arry.StringToAddress(contract),
			Type:         status.Exchange_,
			FunctionType: status.Exchange_Init,
			Function: &exchange_func.ExchangeInitBody{
				Admin:  arry.StringToAddress(admin),
				FeeTo:  arry.StringToAddress(feeTo),
				Symbol: symbol,
			},
		},
	}

	msg.SetHash()
	return msg, nil
}

func NewSetAdmin(from, exchange, admin string, nonce uint64) (*types.Message, error) {
	tx := &types.Message{
		Header: &types.MsgHeader{
			Type:  types.Contract,
			Hash:  arry.Hash{},
			From:  arry.StringToAddress(from),
			Nonce: nonce,
			Fee:   types.ContractFees,
			Time:  uint64(time.Now().Unix()),
			Signature: &types.Signature{
				Bytes:  nil,
				PubKey: nil,
			},
		},
		Body: &types.ContractBody{
			Contract:     arry.StringToAddress(exchange),
			Type:         status.Exchange_,
			FunctionType: status.Exchange_SetAdmin,
			Function: &exchange_func.ExchangeAdmin{
				Address: arry.StringToAddress(admin),
			},
		},
	}
	tx.SetHash()
	return tx, nil
}

func NewSetFeeTo(from, exchange, feeTo string, nonce uint64) (*types.Message, error) {
	tx := &types.Message{
		Header: &types.MsgHeader{
			Type:  types.Contract,
			Hash:  arry.Hash{},
			From:  arry.StringToAddress(from),
			Nonce: nonce,
			Fee:   types.ContractFees,
			Time:  uint64(time.Now().Unix()),
			Signature: &types.Signature{
				Bytes:  nil,
				PubKey: nil,
			},
		},
		Body: &types.ContractBody{
			Contract:     arry.StringToAddress(exchange),
			Type:         status.Exchange_,
			FunctionType: status.Exchange_SetFeeTo,
			Function: &exchange_func.ExchangeFeeTo{
				Address: arry.StringToAddress(feeTo),
			},
		},
	}
	tx.SetHash()
	return tx, nil
}

func NewPairAddLiquidity(net, from, to, exchange, tokenA, tokenB string, amountADesired, amountBDesired, amountAMin, amountBMin, deadline, nonce uint64) (*types.Message, error) {
	contract, err := exchange_runner.PairAddress(net, arry.StringToAddress(tokenA), arry.StringToAddress(tokenB), arry.StringToAddress(exchange))
	if err != nil {
		return nil, err
	}
	tx := &types.Message{
		Header: &types.MsgHeader{
			Type:  types.Contract,
			Hash:  arry.Hash{},
			From:  arry.StringToAddress(from),
			Nonce: nonce,
			Fee:   types.ContractFees,
			Time:  uint64(time.Now().Unix()),
			Signature: &types.Signature{
				Bytes:  nil,
				PubKey: nil,
			},
		},
		Body: &types.ContractBody{
			Contract:     arry.StringToAddress(contract),
			Type:         status.Pair_,
			FunctionType: status.Pair_AddLiquidity,
			Function: &exchange_func.ExchangeAddLiquidity{
				Exchange:       arry.StringToAddress(exchange),
				TokenA:         arry.StringToAddress(tokenA),
				TokenB:         arry.StringToAddress(tokenB),
				To:             arry.StringToAddress(to),
				AmountADesired: amountADesired,
				AmountBDesired: amountBDesired,
				AmountAMin:     amountAMin,
				AmountBMin:     amountBMin,
				Deadline:       deadline,
			},
		},
	}
	tx.SetHash()
	return tx, nil
}

func NewPairRemoveLiquidity(net, from, to, exchange, tokenA, tokenB string, amountAMin, amountBMin, liquidity, deadline, nonce uint64) (*types.Message, error) {
	contract, err := exchange_runner.PairAddress(net, arry.StringToAddress(tokenA), arry.StringToAddress(tokenB), arry.StringToAddress(exchange))
	if err != nil {
		return nil, err
	}
	tx := &types.Message{
		Header: &types.MsgHeader{
			Type:  types.Contract,
			Hash:  arry.Hash{},
			From:  arry.StringToAddress(from),
			Nonce: nonce,
			Fee:   types.ContractFees,
			Time:  uint64(time.Now().Unix()),
			Signature: &types.Signature{
				Bytes:  nil,
				PubKey: nil,
			},
		},
		Body: &types.ContractBody{
			Contract:     arry.StringToAddress(contract),
			Type:         status.Pair_,
			FunctionType: status.Pair_RemoveLiquidity,
			Function: &exchange_func.ExchangeRemoveLiquidity{
				Exchange:   arry.StringToAddress(exchange),
				TokenA:     arry.StringToAddress(tokenA),
				TokenB:     arry.StringToAddress(tokenB),
				To:         arry.StringToAddress(to),
				Liquidity:  liquidity,
				AmountAMin: amountAMin,
				AmountBMin: amountBMin,
				Deadline:   deadline,
			},
		},
	}
	tx.SetHash()
	return tx, nil
}

func NewSwapExactIn(from, to, exchange string, amountIn, amountOutMin uint64, path []string, deadline, nonce uint64) (*types.Message, error) {
	address := make([]arry.Address, 0)
	for _, addr := range path {
		address = append(address, arry.StringToAddress(addr))
	}
	tx := &types.Message{
		Header: &types.MsgHeader{
			Type:  types.Contract,
			Hash:  arry.Hash{},
			From:  arry.StringToAddress(from),
			Nonce: nonce,
			Fee:   types.ContractFees,
			Time:  uint64(time.Now().Unix()),
			Signature: &types.Signature{
				Bytes:  nil,
				PubKey: nil,
			},
		},
		Body: &types.ContractBody{
			Contract:     arry.StringToAddress(exchange),
			Type:         status.Exchange_,
			FunctionType: status.Exchange_ExactIn,
			Function: &exchange_func.ExactIn{
				AmountIn:     amountIn,
				AmountOutMin: amountOutMin,
				Path:         address,
				To:           arry.StringToAddress(to),
				Deadline:     deadline,
			},
		},
	}
	tx.SetHash()
	return tx, nil
}

func NewSwapExactOut(from, to, exchange string, amountOut, amountInMax uint64, path []string, deadline, nonce uint64) (*types.Message, error) {
	address := make([]arry.Address, 0)
	for _, addr := range path {
		address = append(address, arry.StringToAddress(addr))
	}
	tx := &types.Message{
		Header: &types.MsgHeader{
			Type:  types.Contract,
			Hash:  arry.Hash{},
			From:  arry.StringToAddress(from),
			Nonce: nonce,
			Fee:   types.ContractFees,
			Time:  uint64(time.Now().Unix()),
			Signature: &types.Signature{
				Bytes:  nil,
				PubKey: nil,
			},
		},
		Body: &types.ContractBody{
			Contract:     arry.StringToAddress(exchange),
			Type:         status.Exchange_,
			FunctionType: status.Exchange_ExactOut,
			Function: &exchange_func.ExactOut{
				AmountOut:   amountOut,
				AmountInMax: amountInMax,
				Path:        address,
				To:          arry.StringToAddress(to),
				Deadline:    deadline,
			},
		},
	}
	tx.SetHash()
	return tx, nil
}
