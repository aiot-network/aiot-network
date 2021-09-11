package kit

import (
	"bytes"
	"errors"
	"github.com/aiot-network/aiotchain/common/param"
	"github.com/aiot-network/aiotchain/tools/arry"
	"github.com/aiot-network/aiotchain/tools/crypto/base58"
	"github.com/aiot-network/aiotchain/tools/crypto/hash"
	"math/big"
	"unicode"
)

func CalConsumption(amount uint64, proportion uint64) uint64 {
	if float64(amount)/float64(proportion) < 1 {
		return 1
	} else if amount%proportion != 0 {
		return amount/proportion + 1
	}
	return amount / proportion
}

func CalCoinBase(net string, allWorks, works uint64) uint64 {
	params := param.MainNetParam
	switch net {
	case param.MainNet:
		params = param.MainNetParam
	case param.TestNet:
		params = param.TestNetParam
	}
	if allWorks == 0 {
		return 0
	}
	times := params.CycleInterval / params.BlockInterval / uint64(params.CoinBaseAddressList.Len())
	x := big.NewInt(0).Mul(big.NewInt(int64(params.CoinBaseOneDay)), big.NewInt(int64(works)))
	y := big.NewInt(0).Div(x, big.NewInt(int64(allWorks)))
	coinbase := big.NewInt(0).Div(y, big.NewInt(int64(times))).Uint64()
	// params.CoinBaseOneDay * works / allWorks / times
	return coinbase
}

func CalCountByTimes(coinbase *float64, times *uint64, coefficient float64) float64 {
	if *times == 1 {
		return *coinbase
	}
	if *times > 10 {
		return 0
	}
	*coinbase += *coinbase * coefficient
	*times--
	return CalCountByTimes(coinbase, times, coefficient)
}

func GenerateTokenAddress(net string, shorthand string) (string, error) {
	ver := []byte{}
	switch net {
	case param.MainNet:
		ver = append(ver, param.MainNetParam.PubKeyHashTokenID[0:]...)
	case param.TestNet:
		ver = append(ver, param.TestNetParam.PubKeyHashTokenID[0:]...)
	default:
		return "", errors.New("wrong network")
	}
	if err := CheckSymbol(shorthand); err != nil {
		return "", err
	}
	buffBytes := []byte(shorthand)
	hashed := hash.Hash(buffBytes)
	hash160, err := hash.Hash160(hashed.Bytes())
	if err != nil {
		return "", err
	}

	addNet := append(ver, hash160...)
	hashed1 := hash.Hash(addNet)
	hashed2 := hash.Hash(hashed1.Bytes())
	checkSum := hashed2[0:4]
	hashedCheck1 := append(addNet, checkSum...)
	code58 := base58.Encode(hashedCheck1)
	return arry.StringToAddress(code58).String(), nil
}

func CheckContractAddress(net string, address string) bool {

	ver := []byte{}
	switch net {
	case param.MainNet:
		ver = append(ver, param.MainNetParam.PubKeyHashTokenID[0:]...)
		if address == param.MainNetParam.MainToken.String() {
			return true
		}
	case param.TestNet:
		ver = append(ver, param.TestNetParam.PubKeyHashTokenID[0:]...)
		if address == param.TestNetParam.MainToken.String() {
			return true
		}
	default:
		return false
	}
	addr := address
	if len(addr) != addressLength {
		return false
	}
	addrBytes := base58.Decode(addr)
	if len(addrBytes) != addressBytesLength {
		return false
	}
	checkSum := addrBytes[len(addrBytes)-4:]
	checkBytes := addrBytes[0 : len(addrBytes)-4]
	checkBytesHashed1 := hash.Hash(checkBytes)
	checkBytesHashed2 := hash.Hash(checkBytesHashed1.Bytes())
	netBytes := checkBytes[0:2]
	if bytes.Compare(ver, netBytes) != 0 {
		return false
	}
	return bytes.Compare(checkSum, checkBytesHashed2[0:4]) == 0
}

// Check the secondary account name, it must be letters,
// all uppercase or all lowercase, no more than 10
// characters and no less than 2.
func CheckSymbol(shorthand string) error {
	if len(shorthand) < 2 || len(shorthand) > 10 {
		return errors.New("the shorthand length must be in the range of 2 and 10")
	}
	for _, c := range shorthand {
		if !unicode.IsLetter(c) {
			return errors.New("shorthand must be letter")
		}
		if !unicode.IsUpper(c) {
			return errors.New("shorthand must be upper")
		}
	}
	return nil
}

func GenerateContractAddress(net string, bytes []byte) (string, error) {
	ver := []byte{}
	switch net {
	case param.MainNet:
		ver = append(ver, param.MainNetParam.PubKeyHashTokenID[0:]...)
	case param.TestNet:
		ver = append(ver, param.TestNetParam.PubKeyHashTokenID[0:]...)
	default:
		return "", errors.New("wrong network")
	}

	buffBytes := bytes
	hashed := hash.Hash(buffBytes)
	hash160, err := hash.Hash160(hashed.Bytes())
	if err != nil {
		return "", err
	}

	addNet := append(ver, hash160...)
	hashed1 := hash.Hash(addNet)
	hashed2 := hash.Hash(hashed1.Bytes())
	checkSum := hashed2[0:4]
	hashedCheck1 := append(addNet, checkSum...)
	code58 := base58.Encode(hashedCheck1)
	return arry.StringToAddress(code58).String(), nil
}
