package exchange

import (
	"fmt"
	"github.com/aiot-network/aiotchain/tools/arry"
	"github.com/aiot-network/aiotchain/tools/rlp"
	"math/big"
)

const MINIMUM_LIQUIDITY = 1e3

type Pair struct {
	Exchange             arry.Address
	Token0               arry.Address
	Token1               arry.Address
	Reserve0             uint64
	Reserve1             uint64
	BlockTimestampLast   uint32
	Price0CumulativeLast uint64
	Price1CumulativeLast uint64
	KLast                *big.Int
	TotalSupply          uint64
	Symbol               string
	Symbol0              string
	Symbol1              string
}

func NewPair(exchange, token0, token1 arry.Address, symbol0, symbol1, exchangeSymbol string) *Pair {
	return &Pair{
		Exchange:             exchange,
		Token0:               token0,
		Token1:               token1,
		Reserve0:             0,
		Reserve1:             0,
		BlockTimestampLast:   0,
		Price0CumulativeLast: 0,
		Price1CumulativeLast: 0,
		KLast:                big.NewInt(0),
		TotalSupply:          0,
		Symbol:               lpSymbol(symbol0, symbol1, exchangeSymbol),
		Symbol0:              symbol0,
		Symbol1:              symbol1,
	}
}

func (p *Pair) Bytes() []byte {
	bytes, _ := rlp.EncodeToBytes(p)
	return bytes
}

func (p *Pair) Verify() error {
	return nil
}

func (p *Pair) GetReserves() (reserve0 uint64, reserve1 uint64, blockTimestampLast uint32) {
	return p.Reserve0, p.Reserve1, p.BlockTimestampLast
}

func (p *Pair) Mint(number uint64) {
	p.TotalSupply = p.TotalSupply + number
}

func (p *Pair) Burn(number uint64) {
	p.TotalSupply = p.TotalSupply - number
}

func (p *Pair) UpdatePair(balance0, balance1, _reserve0, _reserve1, blockTime uint64, feeOn bool) {
	p.UpdateReserve(balance0, balance1, _reserve0, _reserve1, blockTime)
	if feeOn {
		p.KLast = big.NewInt(0).Mul(big.NewInt(int64(p.Reserve0)), big.NewInt(int64(p.Reserve1)))
	}
}

func (p *Pair) UpdateReserve(balance0, balance1, _reserve0, _reserve1, blockTime uint64) {
	blockTimestamp := uint32(blockTime%2 ^ 32)
	timeElapsed := blockTimestamp - p.BlockTimestampLast // overflow is desired
	if timeElapsed > 0 && _reserve0 != 0 && _reserve1 != 0 {
		// * never overflows, and + overflow is desired
		// 这两个值用于价格预言机
		p.Price0CumulativeLast += _reserve1 / _reserve0 * uint64(timeElapsed)
		p.Price1CumulativeLast += _reserve0 / _reserve1 * uint64(timeElapsed)
	}
	p.BlockTimestampLast = blockTimestamp

	p.Reserve0 = balance0
	p.Reserve1 = balance1
}

func (p *Pair) UpdateKLast() {

}

func DecodeToPair(bytes []byte) (*Pair, error) {
	var pair *Pair
	err := rlp.DecodeBytes(bytes, &pair)
	return pair, err
}

func lpSymbol(symbol0, symbol1, exchangeSymbol string) string {
	return fmt.Sprintf("%sLP-%s-%s", exchangeSymbol, symbol0, symbol1)
}
