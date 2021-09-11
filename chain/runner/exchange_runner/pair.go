package exchange_runner

import (
	bytes2 "bytes"
	"errors"
	"fmt"
	"github.com/aiot-network/aiotchain/chain/common/kit"
	"github.com/aiot-network/aiotchain/chain/runner/library"
	"github.com/aiot-network/aiotchain/chain/types"
	"github.com/aiot-network/aiotchain/chain/types/functionbody/exchange_func"
	"github.com/aiot-network/aiotchain/chain/types/status"
	"github.com/aiot-network/aiotchain/chain/types/status/exchange"
	"github.com/aiot-network/aiotchain/common/config"
	"github.com/aiot-network/aiotchain/tools/amount"
	"github.com/aiot-network/aiotchain/tools/arry"
	"github.com/aiot-network/aiotchain/tools/crypto/base58"
	"github.com/aiot-network/aiotchain/tools/math"
	chaintypes "github.com/aiot-network/aiotchain/types"

	"math/big"
)

type PairState struct {
	library    *library.RunnerLibrary
	exHeader   *status.Contract
	exBody     *exchange.Exchange
	pairHeader *status.Contract
	pairBody   *exchange.Pair
}

func NewPairState(runnerLibrary *library.RunnerLibrary, pairAddress string) (*PairState, error) {
	pairHeader := runnerLibrary.GetContract(pairAddress)
	if pairHeader == nil {
		return nil, fmt.Errorf("pair %s already exist", pairAddress)
	}
	pairBody, _ := pairHeader.Body.(*exchange.Pair)
	exHeader := runnerLibrary.GetContract(pairBody.Exchange.String())
	if exHeader == nil {
		return nil, fmt.Errorf("pair %s already exist", pairAddress)
	}
	exBody, _ := exHeader.Body.(*exchange.Exchange)
	return &PairState{
		library:    runnerLibrary,
		exHeader:   exHeader,
		exBody:     exBody,
		pairHeader: pairHeader,
		pairBody:   pairBody,
	}, nil
}

func (ps *PairState) Methods() map[string]*MethodInfo {
	return pairMethods
}

func (ps *PairState) MethodExist(method string) bool {
	_, exist := pairMethods[method]
	return exist
}

func (ps *PairState) QuoteAmountB(tokenAStr string, amountA float64) (float64, error) {
	tokenA := arry.StringToAddress(tokenAStr)
	tokenB := ps.pairBody.Token0
	if ps.pairBody.Token0.IsEqual(tokenA) {
		tokenB = ps.pairBody.Token1
	}
	reservesA, reservesB := ps.library.GetReservesByPair(ps.pairBody, tokenA, tokenB)
	iAmountA, _ := amount.NewAmount(amountA)
	amountB, err := ps.quote(iAmountA, reservesA, reservesB)
	if err != nil {
		return 0, err
	} else {
		return amount.Amount(amountB).ToCoin(), nil
	}
}

func (ps *PairState) optimalAmount(reserveA, reserveB, amountADesired, amountBDesired, amountAMin, amountBMin uint64) (uint64, uint64, error) {
	if reserveA == 0 && reserveB == 0 {
		return amountADesired, amountBDesired, nil
	} else {
		// 最优数量B = 期望数量A * 储备B / 储备A
		amountBOptimal, err := ps.quote(amountADesired, reserveA, reserveB)
		if err != nil {
			return 0, 0, err
		}
		// 如果最优数量B < B的期望数量
		if amountBOptimal <= amountBDesired {
			if amountBOptimal < amountBMin {
				return 0, 0, errors.New("insufficient amountB")
			}
			return amountADesired, amountBOptimal, nil
		} else {
			// 则计算 最优数量A = 期望数量B * 储备A / 储备B
			amountAOptimal, err := ps.quote(amountBDesired, reserveB, reserveA)
			if err != nil {
				return 0, 0, err
			}
			if amountAOptimal < amountAMin {
				return 0, 0, errors.New("insufficient amountA")
			}
			return amountAOptimal, amountBDesired, nil
		}
	}
}

// Quote given some amount of an asset and pair reserves, returns an equivalent amount of the other asset
func (ps *PairState) quote(amountA, reserveA, reserveB uint64) (uint64, error) {
	if amountA <= 0 {
		return 0, errors.New("insufficient_amount")
	}
	if reserveA <= 0 || reserveB <= 0 {
		return 0, errors.New("insufficient_liquidity")
	}
	amountB := big.NewInt(0).Div(big.NewInt(0).Mul(big.NewInt(int64(amountA)), big.NewInt(int64(reserveB))), big.NewInt(int64(reserveA)))
	return amountB.Uint64(), nil
}

type TotalValue struct {
	Token0  string  `json:"token0"`
	Symbol0 string  `json:"symbol0"`
	Value0  float64 `json:"value0"`
	Token1  string  `json:"token1"`
	Symbol1 string  `json:"symbol1"`
	Value1  float64 `json:"value1"`
}

func (ps *PairState) TotalValue(liquidity float64) (*TotalValue, error) {
	iLiquidity, _ := amount.NewAmount(liquidity)
	amount0, amount1, err := ps.totalValue(ps.pairBody.Token0, ps.pairBody.Token1, iLiquidity)
	if err != nil {
		return nil, err
	}
	return &TotalValue{
		Token0:  ps.pairBody.Token0.String(),
		Symbol0: ps.pairBody.Symbol0,
		Value0:  amount.Amount(amount0).ToCoin(),
		Token1:  ps.pairBody.Token1.String(),
		Symbol1: ps.pairBody.Symbol1,
		Value1:  amount.Amount(amount1).ToCoin(),
	}, nil
}

func (ps *PairState) totalValue(tokenA, tokenB arry.Address, liquidity uint64) (uint64, uint64, error) {
	token0, token1 := library.SortToken(tokenA, tokenB)
	_reserve0, _reserve1 := ps.library.GetReservesByPair(ps.pairBody, token0, token1)

	_totalSupply := ps.pairBody.TotalSupply
	if _totalSupply < liquidity {
		return 0, 0, fmt.Errorf("exceeding the maximum liquidity value %.8f", amount.Amount(_totalSupply).ToCoin())
	}

	amount0 := new(big.Int).Div(new(big.Int).Mul(big.NewInt(int64(liquidity)), big.NewInt(int64(_reserve0))), big.NewInt(int64(_totalSupply))).Uint64()
	amount1 := new(big.Int).Div(new(big.Int).Mul(big.NewInt(int64(liquidity)), big.NewInt(int64(_reserve1))), big.NewInt(int64(_totalSupply))).Uint64()
	if tokenA.IsEqual(token0) {
		return amount0, amount1, nil
	}
	return amount1, amount0, nil
}

// if fee is on, mint liquidity equivalent to 1/6th of the growth in sqrt(k)
func (ps *PairState) mintFee(_reserve0, _reserve1 uint64) (bool, uint64, error) {
	var feeLiquidity uint64
	feeTo := ps.exBody.FeeTo
	// 收费地址被设置，则收费开
	feeOn := !feeTo.IsEqual(arry.Address{})
	_kLast := ps.pairBody.KLast // gas savings
	if feeOn {
		if _kLast.Cmp(big.NewInt(0)) != 0 {
			// rootK = Sqrt(_reserve0 * _reserve1)
			rootK := big.NewInt(0).Sqrt(big.NewInt(0).Mul(big.NewInt(int64(_reserve0)), big.NewInt(int64(_reserve1))))
			// rootKLast = Sqrt(_kLast)
			rootKLast := big.NewInt(0).Sqrt(_kLast)
			if rootK.Cmp(rootKLast) > 0 {
				// numerator = (rootK-rootKLast)*TotalSupply
				numerator := big.NewInt(0).Mul(big.NewInt(0).Sub(rootK, rootKLast), big.NewInt(int64(ps.pairBody.TotalSupply)))
				// denominator =  * 5 + rootKLast
				denominator := big.NewInt(0).Add(big.NewInt(0).Mul(rootK, big.NewInt(5)), rootKLast)
				// liquidity = numerator / denominator
				liquidityBig := big.NewInt(0).Div(numerator, denominator)
				if liquidityBig.Cmp(big.NewInt(0)) > 0 {
					feeLiquidity = liquidityBig.Uint64()
				}
			}
		}
	} else if _kLast.Cmp(big.NewInt(0)) != 0 {
		ps.pairBody.KLast = big.NewInt(0)
	}
	return feeOn, feeLiquidity, nil
}

type PairRunner struct {
	pairState    *PairState
	contractBody *types.ContractBody
	addBody      *exchange_func.ExchangeAddLiquidity
	removeBody   *exchange_func.ExchangeRemoveLiquidity
	address      arry.Address
	tx           chaintypes.IMessage
	txBody       chaintypes.IMessageBody
	sender       arry.Address
	state        *types.ContractStatus
	events       []*types.Event
	height       uint64
	blockTime    uint64
	isCreate     bool
}

func NewPairRunner(lib *library.RunnerLibrary, tx chaintypes.IMessage, height, blockTime uint64) *PairRunner {
	var exBody *exchange.Exchange
	var pairBody *exchange.Pair
	var exchangeAddr string
	var addBody *exchange_func.ExchangeAddLiquidity
	var removeBody *exchange_func.ExchangeRemoveLiquidity
	txBody := tx.MsgBody()
	contractBody, _ := txBody.(*types.ContractBody)
	address := contractBody.Contract

	switch contractBody.FunctionType {
	case status.Pair_AddLiquidity:
		addBody, _ = contractBody.Function.(*exchange_func.ExchangeAddLiquidity)
		exchangeAddr = addBody.Exchange.String()
	case status.Pair_RemoveLiquidity:
		removeBody, _ = contractBody.Function.(*exchange_func.ExchangeRemoveLiquidity)
		exchangeAddr = removeBody.Exchange.String()
	}

	exHeader := lib.GetContract(exchangeAddr)
	if exHeader != nil {
		exBody, _ = exHeader.Body.(*exchange.Exchange)
	}

	pairHeader := lib.GetContract(address.String())
	if pairHeader != nil {
		pairBody, _ = pairHeader.Body.(*exchange.Pair)
	}
	state := &types.ContractStatus{State: types.Status_Success}
	return &PairRunner{
		pairState: &PairState{
			library:    lib,
			exHeader:   exHeader,
			exBody:     exBody,
			pairHeader: pairHeader,
			pairBody:   pairBody,
		},
		contractBody: contractBody,
		addBody:      addBody,
		removeBody:   removeBody,
		address:      address,
		state:        state,
		tx:           tx,
		height:       height,
		sender:       tx.From(),
		blockTime:    blockTime,
		events:       make([]*types.Event, 0),
	}
}

func (p *PairRunner) PreAddLiquidityVerify() error {
	if p.addBody.Deadline != 0 && p.addBody.Deadline < p.height {
		return fmt.Errorf("past the deadline")
	}
	if p.pairState.exHeader == nil {
		return fmt.Errorf("exchange %s is not exist", p.addBody.Exchange.String())
	}
	/*	if !p.sender.IsEqual(p.exchange.Admin) {
			return errors.New("forbidden")
		}
	*/
	noMainTokenCount := 0
	if !p.addBody.TokenA.IsEqual(config.Param.MainToken) {
		noMainTokenCount++
		if contract := p.pairState.library.GetToken(p.addBody.TokenA.String()); contract == nil {
			return fmt.Errorf("tokenA %s is not exist", p.addBody.TokenA.String())
		}
	}

	if !p.addBody.TokenB.IsEqual(config.Param.MainToken) {
		noMainTokenCount++
		if contract := p.pairState.library.GetToken(p.addBody.TokenB.String()); contract == nil {
			return fmt.Errorf("tokenB %s is not exist", p.addBody.TokenB.String())
		}
	}

	address, err := PairAddress(config.Param.Name, p.addBody.TokenA, p.addBody.TokenB, p.pairState.exHeader.Address)
	if err != nil {
		return fmt.Errorf("pair address error")
	}
	if address != p.address.String() {
		return fmt.Errorf("wrong pair contract address")
	}
	if err := p.preBalanceVerify(address); err != nil {
		return err
	}
	_, err = p.pairState.exBody.LegalPair(p.addBody.TokenA.String(), p.addBody.TokenB.String())
	return err
}

func (p *PairRunner) preBalanceVerify(pairAddr string) error {
	var amountA, amountB uint64
	var err error
	if p.pairState.pairBody == nil {
		amountA = p.addBody.AmountADesired
		amountB = p.addBody.AmountADesired
	} else {
		pairContract := p.pairState.library.GetContract(pairAddr)
		if pairContract == nil {
			return fmt.Errorf("the pair %s is not exist", pairAddr)
		}
		pair := pairContract.Body.(*exchange.Pair)
		reserveA, reserveB := p.pairState.library.GetReservesByPair(pair, p.addBody.TokenA, p.addBody.TokenB)
		amountA, amountB, err = p.pairState.optimalAmount(reserveA, reserveB, p.addBody.AmountADesired, p.addBody.AmountBDesired, p.addBody.AmountAMin, p.addBody.AmountBMin)
		if err != nil {
			return err
		}
	}

	balanceA := p.pairState.library.GetBalance(p.sender, p.addBody.TokenA)
	if balanceA < amountA {
		return fmt.Errorf("insufficient balance %s", p.sender.String())
	}
	balanceB := p.pairState.library.GetBalance(p.sender, p.addBody.TokenB)
	if balanceB < amountB {
		return fmt.Errorf("insufficient balance %s", p.sender.String())
	}
	return nil
}

func (p *PairRunner) PreRemoveLiquidityVerify(lastHeight uint64) error {
	if p.removeBody.Deadline != 0 && p.removeBody.Deadline < lastHeight {
		return fmt.Errorf("past the deadline")
	}
	if p.pairState.exHeader == nil {
		return fmt.Errorf("exchange %s is not exist", p.removeBody.Exchange.String())
	}
	/*	if !p.sender.IsEqual(p.exchange.Admin) {
			return errors.New("forbidden")
		}
	*/
	if p.removeBody.Liquidity == 0 {
		return fmt.Errorf("invalid liquidity")
	}
	if !p.removeBody.TokenA.IsEqual(config.Param.MainToken) {
		if contract := p.pairState.library.GetToken(p.removeBody.TokenA.String()); contract == nil {
			return fmt.Errorf("tokenA %s is not exist", p.removeBody.TokenA.String())
		}
	}
	if !p.removeBody.TokenB.IsEqual(config.Param.MainToken) {
		if contract := p.pairState.library.GetToken(p.removeBody.TokenB.String()); contract == nil {
			return fmt.Errorf("tokenB %s is not exist", p.removeBody.TokenB.String())
		}
	}

	address, err := PairAddress(config.Param.Name, p.removeBody.TokenA, p.removeBody.TokenB, p.pairState.exHeader.Address)
	if err != nil {
		return fmt.Errorf("pair address error")
	}
	if address != p.address.String() {
		return fmt.Errorf("wrong pair contract address")
	}
	if p.pairState.pairBody == nil {
		return fmt.Errorf("pair is not exist")
	}
	balance := p.pairState.library.GetBalance(p.sender, p.address)
	if balance < p.removeBody.Liquidity {
		return fmt.Errorf("%s's liquidity token is insufficient", p.sender.String())
	}
	token0, token1 := library.SortToken(p.removeBody.TokenA, p.removeBody.TokenB)
	_reserve0, _reserve1 := p.pairState.library.GetReservesByPair(p.pairState.pairBody, token0, token1)

	_liquidity := p.removeBody.Liquidity
	if balance < _liquidity {
		return fmt.Errorf("%s's liquidity token is insufficient", p.sender.String())
	}
	_totalSupply := p.pairState.pairBody.TotalSupply
	if _totalSupply < p.removeBody.Liquidity {
		return fmt.Errorf("%s's liquidity token is insufficient", p.address.String())
	}

	amount0 := new(big.Int).Div(new(big.Int).Mul(big.NewInt(int64(_liquidity)), big.NewInt(int64(_reserve0))), big.NewInt(int64(_totalSupply))).Uint64()
	amount1 := new(big.Int).Div(new(big.Int).Mul(big.NewInt(int64(_liquidity)), big.NewInt(int64(_reserve1))), big.NewInt(int64(_totalSupply))).Uint64()
	if token0.IsEqual(p.removeBody.TokenA) {
		if amount0 < p.removeBody.AmountAMin || amount1 < p.removeBody.AmountBMin {
			return fmt.Errorf("not meet expectations")
		}
	} else {
		if amount0 < p.removeBody.AmountBMin || amount1 < p.removeBody.AmountAMin {
			return fmt.Errorf("not meet expectations")
		}
	}
	return nil
}

func (p *PairRunner) AddLiquidity() {
	var ERR error
	var err error
	var feeLiquidity uint64
	var feeOn bool
	defer func() {
		if ERR != nil {
			p.state.State = types.Status_Failed
			p.state.Error = ERR.Error()
		} else {
			p.state.Event = p.events
		}
		p.pairState.library.SetContractState(p.tx.Hash().String(), p.state)
	}()
	if p.addBody.Deadline != 0 && p.addBody.Deadline < p.height {
		ERR = fmt.Errorf("past the deadline")
		return
	}

	if p.pairState.exHeader == nil {
		ERR = fmt.Errorf("exchange %s is not exist", p.addBody.Exchange.String())
		return
	}

	if p.pairState.pairBody == nil {
		p.createPair()
	}

	_reserveA, _reserveB := p.pairState.library.GetReservesByPair(p.pairState.pairBody, p.addBody.TokenA, p.addBody.TokenB)
	_reserve0, _reserve1 := p.pairState.pairBody.Reserve0, p.pairState.pairBody.Reserve1

	amountA, amountB, err := p.pairState.optimalAmount(_reserveA, _reserveB, p.addBody.AmountADesired, p.addBody.AmountBDesired, p.addBody.AmountAMin, p.addBody.AmountBMin)
	if err != nil {
		ERR = err
		return
	}

	liquidity, feeLiquidity, feeOn, err := p.mintLiquidityValue(_reserveA, _reserveB, amountA, amountB)
	if err != nil {
		ERR = err
		return
	}
	if p.addBody.TokenA.IsEqual(p.pairState.pairBody.Token0) {
		p.pairState.pairBody.UpdatePair(_reserve0+amountA, _reserve1+amountB, _reserve0, _reserve1, p.height, feeOn)
	} else {
		p.pairState.pairBody.UpdatePair(_reserve0+amountB, _reserve1+amountA, _reserve0, _reserve1, p.height, feeOn)
	}

	p.transferEvent(p.sender, p.address, p.addBody.TokenA, amountA)
	p.transferEvent(p.sender, p.address, p.addBody.TokenB, amountB)
	p.mintEvent(p.addBody.To, p.address, liquidity)
	if feeOn {
		p.mintEvent(p.pairState.exBody.FeeTo, p.address, feeLiquidity)
	}

	if err = p.runEvents(); err != nil {
		ERR = err
		return
	}
	p.update()
}

func (p *PairRunner) createPair() {
	token0, token1 := library.SortToken(p.addBody.TokenA, p.addBody.TokenB)
	symbol0, _ := p.pairState.library.ContractSymbol(token0)
	symbol1, _ := p.pairState.library.ContractSymbol(token1)
	p.pairState.pairBody = exchange.NewPair(p.addBody.Exchange, token0, token1, symbol0, symbol1, p.pairState.exBody.Symbol)

	p.pairState.pairHeader = &status.Contract{
		Address:    p.address,
		CreateHash: p.tx.Hash(),
		Type:       status.Pair_,
		Body:       p.pairState.pairBody,
	}
	p.pairState.exBody.AddPair(token0, token1, p.address, symbol0, symbol1)
	p.isCreate = true
}

func (p *PairRunner) RemoveLiquidity() {
	var ERR error
	defer func() {
		if ERR != nil {
			p.state.State = types.Status_Failed
			p.state.Error = ERR.Error()
		} else {
			p.state.Event = p.events
		}
		p.pairState.library.SetContractState(p.tx.Hash().String(), p.state)
	}()
	if p.removeBody.Deadline != 0 && p.removeBody.Deadline < p.height {
		ERR = fmt.Errorf("past the deadline")
		return
	}
	if p.pairState.exHeader == nil {
		ERR = fmt.Errorf("exchange %s is not exist", p.addBody.Exchange.String())
		return
	}

	if p.pairState.pairBody == nil {
		ERR = errors.New("pair is not exist")
		return
	}

	token0, token1 := library.SortToken(p.removeBody.TokenA, p.removeBody.TokenB)
	_reserve0, _reserve1 := p.pairState.library.GetReservesByPair(p.pairState.pairBody, token0, token1)
	feeOn, feeLiquidity, err := p.pairState.mintFee(_reserve0, _reserve1)
	if err != nil {
		ERR = err
		return
	}
	balance := p.pairState.library.GetBalance(p.sender, p.address)
	_liquidity := p.removeBody.Liquidity
	if balance < _liquidity {
		ERR = fmt.Errorf("%s's liquidity token is insufficient", p.sender.String())
		return
	}
	_totalSupply := p.pairState.pairBody.TotalSupply
	if _totalSupply < p.removeBody.Liquidity {
		ERR = fmt.Errorf("%s's liquidity token is insufficient", p.address.String())
		return
	}

	amount0 := new(big.Int).Div(new(big.Int).Mul(big.NewInt(int64(_liquidity)), big.NewInt(int64(_reserve0))), big.NewInt(int64(_totalSupply))).Uint64()
	amount1 := new(big.Int).Div(new(big.Int).Mul(big.NewInt(int64(_liquidity)), big.NewInt(int64(_reserve1))), big.NewInt(int64(_totalSupply))).Uint64()
	if token0.IsEqual(p.removeBody.TokenA) {
		if amount0 < p.removeBody.AmountAMin || amount1 < p.removeBody.AmountBMin {
			ERR = fmt.Errorf("not meet expectations")
			return
		}
	} else {
		if amount0 < p.removeBody.AmountBMin || amount1 < p.removeBody.AmountAMin {
			ERR = fmt.Errorf("not meet expectations")
			return
		}
	}
	p.pairState.pairBody.UpdatePair(_reserve0-amount0, _reserve1-amount1, _reserve0, _reserve1, p.height, feeOn)

	if feeOn {
		p.mintEvent(p.pairState.exBody.FeeTo, p.address, feeLiquidity)
	}
	p.burnEvent(p.sender, p.address, p.removeBody.Liquidity)
	p.transferEvent(p.address, p.removeBody.To, token0, amount0)
	p.transferEvent(p.address, p.removeBody.To, token1, amount1)

	if err = p.runEvents(); err != nil {
		ERR = err
		return
	}
	p.update()
}

type mint struct {
	Address arry.Address
	Amount  uint64
}

func (p *PairRunner) mintLiquidityValue(_reserve0, _reserve1, amount0, amount1 uint64) (uint64, uint64, bool, error) {
	// must be defined here since totalSupply can update in mintFee
	_totalSupply := p.pairState.pairBody.TotalSupply
	// 返回铸造币的手续费开关
	feeOn, feeLiquidity, err := p.pairState.mintFee(_reserve0, _reserve1)
	if err != nil {
		return 0, 0, false, err
	}
	var liquidityValue uint64

	if _totalSupply == 0 {
		// sqrt(amount0 * amount1)
		liquidityBig := big.NewInt(0).Sqrt(big.NewInt(0).Mul(big.NewInt(int64(amount0)), big.NewInt(int64(amount1))))
		liquidityValue = liquidityBig.Uint64()
	} else {
		// valiquidityValue1 = amount0 / _reserve0 * _totalSupply
		// valiquidityValue2 = amount1 / _reserve1 * _totalSupply
		value0 := big.NewInt(0).Mul(big.NewInt(int64(amount0)), big.NewInt(int64(_totalSupply)))
		value1 := big.NewInt(0).Mul(big.NewInt(int64(amount1)), big.NewInt(int64(_totalSupply)))
		liquidityValue = math.Min(big.NewInt(0).Div(value0, big.NewInt(int64(_reserve0))).Uint64(), big.NewInt(0).Div(value1, big.NewInt(int64(_reserve1))).Uint64())
		if liquidityValue <= 0 {
			return 0, 0, false, errors.New("insufficient liquidity minted")
		}
	}

	return liquidityValue, feeLiquidity, feeOn, nil
}

func (p *PairRunner) update() {
	p.pairState.exHeader.Body = p.pairState.exBody
	p.pairState.pairHeader.Body = p.pairState.pairBody
	p.pairState.library.SetContract(p.pairState.exHeader)
	p.pairState.library.SetContract(p.pairState.pairHeader)
	if p.isCreate {
		p.pairState.library.SetSymbol(p.pairState.pairBody.Symbol, p.address.String())
	}
}

func (p *PairRunner) transferEvent(from, to, token arry.Address, amount uint64) {
	p.events = append(p.events, &types.Event{
		EventType: types.Event_Transfer,
		From:      from,
		To:        to,
		Token:     token,
		Amount:    amount,
		Height:    p.height,
	})
}

func (p *PairRunner) mintEvent(to, token arry.Address, amount uint64) {
	p.pairState.pairBody.Mint(amount)
	p.events = append(p.events, &types.Event{
		EventType: types.Event_Mint,
		From:      arry.StringToAddress("mint"),
		To:        to,
		Token:     token,
		Amount:    amount,
		Height:    p.height,
	})
}

func (p *PairRunner) burnEvent(from, token arry.Address, amount uint64) {
	p.pairState.pairBody.Burn(amount)
	p.events = append(p.events, &types.Event{
		EventType: types.Event_Burn,
		From:      from,
		To:        arry.StringToAddress("burn"),
		Token:     token,
		Amount:    amount,
		Height:    p.height,
	})
}

func (p *PairRunner) runEvents() error {
	for _, event := range p.events {
		if err := p.pairState.library.PreRunEvent(event); err != nil {
			return err
		}
	}
	for _, event := range p.events {
		p.pairState.library.RunEvent(event)
	}
	return nil
}

func PairAddress(net string, tokenA, tokenB arry.Address, exchange arry.Address) (string, error) {
	token0, token1 := library.SortToken(tokenA, tokenB)
	bytes := bytes2.Join([][]byte{base58.Decode(token0.String()), base58.Decode(token1.String()),
		base58.Decode(exchange.String())}, []byte{})
	return kit.GenerateContractAddress(net, bytes)
}
