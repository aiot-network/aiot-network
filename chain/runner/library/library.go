package library

import (
	"errors"
	"fmt"
	"github.com/aiot-network/aiotchain/chain/common/status/act_status"
	"github.com/aiot-network/aiotchain/chain/common/status/token_status"
	"github.com/aiot-network/aiotchain/chain/types"
	"github.com/aiot-network/aiotchain/chain/types/status"
	"github.com/aiot-network/aiotchain/chain/types/status/exchange"
	"github.com/aiot-network/aiotchain/common/config"
	arry "github.com/aiot-network/aiotchain/tools/arry"

	"strings"
)

type RunnerLibrary struct {
	aState *act_status.ActStatus
	cState *token_status.TokenStatus
}

func NewRunnerLibrary(aState *act_status.ActStatus, cState *token_status.TokenStatus) *RunnerLibrary {
	return &RunnerLibrary{aState: aState, cState: cState}
}

func (r *RunnerLibrary) ContractSymbol(address arry.Address) (string, error) {
	if address.IsEqual(config.Param.MainToken) {
		return config.Param.MainToken.String(), nil
	}
	contract, _ := r.cState.Token(address)
	if contract == nil {
		return "", fmt.Errorf("%s is not exist", address.String())
	}
	return contract.Symbol(), nil
}

func (r *RunnerLibrary) GetToken(contractAddr string) *types.TokenRecord {
	token, _ := r.cState.Token(arry.StringToAddress(contractAddr))
	return token.(*types.TokenRecord)
}

func (r *RunnerLibrary) SetToken(token *types.TokenRecord) {
	r.cState.SetToken(token)
}

func (r *RunnerLibrary) GetContract(address string) *status.Contract {
	contract, err := r.cState.Contract(arry.StringToAddress(address))
	if err != nil {
		return nil
	}
	return contract.(*status.Contract)
}

func (r *RunnerLibrary) SetContract(contract *status.Contract) {
	r.cState.SetContract(contract)
}

func (r RunnerLibrary) SetContractState(msgHash string, state *types.ContractStatus) {
	hash, _ := arry.StringToHash(msgHash)
	r.cState.SetContractState(hash, state)
}

func (r RunnerLibrary) GetBalance(address arry.Address, token arry.Address) uint64 {
	account := r.aState.Account(address)
	return account.GetBalance(token)
}

func (r *RunnerLibrary) PreRunEvent(event *types.Event) error {
	switch event.EventType {
	case types.Event_Transfer:
		return r.aState.PreTransfer(event.From, event.To, event.Token, event.Amount, event.Height)
	case types.Event_Mint:
		return nil
	case types.Event_Burn:
		return r.aState.PreBurn(event.From, event.Token, event.Amount, event.Height)
	}
	return fmt.Errorf("invalid event type")
}

func (r *RunnerLibrary) RunEvent(event *types.Event) {
	switch event.EventType {
	case types.Event_Transfer:
		r.aState.Transfer(event.From, event.To, event.Token, event.Amount, event.Height)
	case types.Event_Mint:
		r.aState.Mint(event.To, event.Token, event.Amount, event.Height)
	case types.Event_Burn:
		r.aState.Burn(event.From, event.Token, event.Amount, event.Height)
	}
}

func (r *RunnerLibrary) GetSymbolContract(symbol string) (arry.Address, error) {
	contract, exist := r.cState.SymbolContract(symbol)
	if exist {
		return arry.Address{}, fmt.Errorf("%s already exist", symbol)
	}
	return contract, nil
}

func (r *RunnerLibrary) SetSymbol(symbol string, contract string) {
	r.cState.SetSymbolContract(symbol, arry.StringToAddress(contract))
}

func (r *RunnerLibrary) GetPair(pairAddress arry.Address) (*exchange.Pair, error) {
	pairContract := r.GetContract(pairAddress.String())
	if pairContract == nil {
		return nil, errors.New("%s pair does not exist")
	}
	return pairContract.Body.(*exchange.Pair), nil
}

func (r *RunnerLibrary) GetExchange(exchangeAddress arry.Address) (*exchange.Exchange, error) {
	exContract := r.GetContract(exchangeAddress.String())
	if exContract != nil {
		return nil, errors.New("%s exchange does not exist")
	}
	return exContract.Body.(*exchange.Exchange), nil
}

func (r *RunnerLibrary) GetReservesByPairAddress(pairAddress, tokenA, tokenB arry.Address) (uint64, uint64, error) {
	pairContract := r.GetContract(pairAddress.String())
	if pairContract == nil {
		return 0, 0, fmt.Errorf("pair %s  dose not exist", pairAddress.String())
	}
	pair := pairContract.Body.(*exchange.Pair)
	reserves0, reserves1 := r.GetReservesByPair(pair, tokenA, tokenB)
	return reserves0, reserves1, nil
}

func (r *RunnerLibrary) GetReservesByPair(pair *exchange.Pair, tokenA, tokenB arry.Address) (uint64, uint64) {
	reserve0, reserve1, _ := pair.GetReserves()
	token0, _ := SortToken(tokenA, tokenB)
	if tokenA.IsEqual(token0) {
		return reserve0, reserve1
	} else {
		return reserve1, reserve0
	}
}

func SortToken(tokenA, tokenB arry.Address) (arry.Address, arry.Address) {
	if strings.Compare(tokenA.String(), tokenB.String()) > 0 {
		return tokenA, tokenB
	} else {
		return tokenB, tokenA
	}
}
