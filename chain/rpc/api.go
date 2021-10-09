package rpc

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/aiot-network/aiotchain/chain/common/kit"
	"github.com/aiot-network/aiotchain/chain/common/kit/message"
	rpctypes "github.com/aiot-network/aiotchain/chain/rpc/types"
	"github.com/aiot-network/aiotchain/chain/runner"
	chaintypes "github.com/aiot-network/aiotchain/chain/types"
	chainstatus "github.com/aiot-network/aiotchain/chain/types/status"
	"github.com/aiot-network/aiotchain/common/blockchain"
	"github.com/aiot-network/aiotchain/common/status"
	"github.com/aiot-network/aiotchain/service/peers"
	"github.com/aiot-network/aiotchain/service/pool"
	"github.com/aiot-network/aiotchain/tools/arry"
	"github.com/aiot-network/aiotchain/types"
)

type Api struct {
	status   status.IStatus
	msgPool  *pool.Pool
	chain    blockchain.IChain
	peers    *peers.Peers
	runner   *runner.ContractRunner
	getLocal func() *types.Local
}

func NewApi(status status.IStatus,
	msgPool *pool.Pool,
	chain blockchain.IChain,
	peers *peers.Peers,
	runner *runner.ContractRunner) *Api {
	return &Api{
		status:  status,
		chain:   chain,
		peers:   peers,
		runner:  runner,
		msgPool: msgPool,
	}
}

type ApiResponse struct {
}

func (a *Api) GetAccount(address string) (*rpctypes.Account, error) {
	arryAddr := arry.StringToAddress(address)
	/*if !kit.CheckAddress(config.Param.Name, arryAddr.String()) {
		return nil, fmt.Errorf("%s address check failed", address)
	}*/
	account := a.status.Account(arryAddr)
	return rpctypes.ToRpcAccount(account.(*chaintypes.Account)), nil
}

func (a *Api) SendMessageRaw(msgRaw []byte) (string, error) {
	var rpcMsg *chaintypes.RpcMessage
	if err := json.Unmarshal(msgRaw, &rpcMsg); err != nil {
		return "", err
	}
	tx, err := chaintypes.RpcMsgToMsg(rpcMsg)
	if err != nil {
		return "", err
	}
	if err := a.msgPool.Put(tx, false); err != nil {
		return "", err
	}
	return tx.Hash().String(), nil
}

func (a *Api) GetMessage(hash string) (*chaintypes.RpcMessageWithHeight, error) {
	var msg types.IMessage
	var exist bool
	var err error
	var height uint64
	var confirmed bool
	hashArry, err := arry.StringToHash(hash)
	if err != nil {
		return nil, fmt.Errorf("wrong hash " + err.Error())
	}
	msg, err = a.chain.GetMessage(hashArry)
	if err != nil {
		msg, exist = a.msgPool.GetMessage(hashArry)
		if !exist {
			return nil, fmt.Errorf("message hash %s does not exist", hash)
		}
		height = 0
		confirmed = false
	} else {
		index, err := a.chain.GetMessageIndex(hashArry)
		if err != nil {
			return nil, fmt.Errorf("%s is not exist", hash)
		}
		confirmedHeight := a.chain.LastConfirmed()
		height = index.GetHeight()
		confirmed = confirmedHeight >= height
	}
	status := a.status.ContractState(msg.Hash())

	rpcMsg, _ := chaintypes.MsgToRpcMsgWithState(msg.(*chaintypes.Message), status.(*chaintypes.ContractStatus))
	rsMsg := &chaintypes.RpcMessageWithHeight{
		MsgHeader: rpcMsg.MsgHeader,
		MsgBody:   rpcMsg.MsgBody,
		Height:    height,
		Confirmed: confirmed,
	}

	return rsMsg, nil
}

func (a *Api) GetBlockHash(hash string) (*chaintypes.RpcBlock, error) {
	hashArry, err := arry.StringToHash(hash)
	if err != nil {
		return nil, fmt.Errorf("wrong hash")
	}
	block, err := a.chain.GetBlockHash(hashArry)
	if err != nil {
		return nil, err
	}
	return chaintypes.BlockToRpcBlock(block.(*chaintypes.Block), a.chain.LastConfirmed(), a.status.ContractState)
}

func (a *Api) GetBlockHeight(height uint64) (*chaintypes.RpcBlock, error) {
	block, err := a.chain.GetBlockHeight(height)
	if err != nil {
		return nil, err
	}
	return chaintypes.BlockToRpcBlock(block.(*chaintypes.Block), a.chain.LastConfirmed(), a.status.ContractState)
}

func (a *Api) LastHeight() uint64 {
	return a.chain.LastHeight()
}

func (a *Api) Confirmed() uint64 {
	return a.chain.LastConfirmed()
}

func (a *Api) GetMsgPool() *rpctypes.TxPool {
	preparedTxs, futureTxs := a.msgPool.All()
	return rpctypes.MsgsToRpcMsgsPool(preparedTxs, futureTxs)
}

func (a *Api) Candidates() (*rpctypes.RpcCandidates, error) {
	candidates := a.status.Candidates()
	if candidates == nil || candidates.Len() == 0 {
		return nil, fmt.Errorf("no candidates")
	}
	cas := candidates.(*chaintypes.Candidates)
	for i, can := range cas.Members {
		for _, v := range can.Voters {
			cas.Members[i].Weight += a.chain.Vote(v)
		}
	}
	return rpctypes.CandidatesToRpcCandidates(cas), nil
}

func (a *Api) GetCycleSupers(cycle uint64) (*rpctypes.RpcCandidates, error) {
	supers := a.status.CycleSupers(cycle)
	if supers == nil {
		return nil, fmt.Errorf("no supers")
	}
	return rpctypes.SupersToRpcCandidates(supers), nil
}

func (a *Api) GetSupersReward(cycle uint64) ([]rpctypes.RpcReword, error) {
	reword := a.status.CycleReword(cycle)
	if reword == nil {
		return nil, fmt.Errorf("no reword")
	}
	return rpctypes.ToRpcReword(reword), nil
}

func (a *Api) Token(address string) (*rpctypes.RpcToken, error) {
	iToken, err := a.status.Token(arry.StringToAddress(address))
	if err != nil {
		return nil, err
	}
	return rpctypes.TokenToRpcToken(iToken.(*chaintypes.TokenRecord)), nil

}

func (a *Api) GetContract(address string) (interface{}, error) {
	contract, err := a.status.Contract(arry.StringToAddress(address))
	if err != nil {
		return a.Token(address)
	}
	return rpctypes.TranslateContractToRpcContract(contract.(*chainstatus.Contract)), nil
}

func (a *Api) PeersInfo() []*types.Local {
	return a.peers.PeersInfo()
}

func (a *Api) LocalInfo() (*types.Local, error) {
	if a.getLocal != nil {
		return a.getLocal(), nil
	}
	return nil, fmt.Errorf("no local info")
}

func (a *Api) GenerateAddress(network, public string) (string, error) {
	return kit.GenerateAddress(network, public)
}

func (a *Api) GenerateTokenAddress(network, abbr string) (string, error) {
	return kit.GenerateTokenAddress(network, abbr)
}

func (a *Api) CreateTransaction(from, to, token string, amount, fees, nonce, timestamp uint64) *chaintypes.Message {
	toMap := []map[string]uint64{
		map[string]uint64{to: amount},
	}

	return message.NewTransaction(from, token, toMap, fees, nonce, timestamp)
}

func (a *Api) SendTransaction(from, to, token string, amount, fees, nonce, timestamp uint64, signature, public string) (string, error) {
	toMap := []map[string]uint64{
		map[string]uint64{to: amount},
	}
	message := message.NewTransaction(from, token, toMap, fees, nonce, timestamp)

	signatureBytes, err := hex.DecodeString(signature)
	if err != nil {
		return "", err
	}
	pubKey, err := hex.DecodeString(public)
	if err != nil {
		return "", err
	}
	message.Header.Signature.Bytes = signatureBytes
	message.Header.Signature.PubKey = pubKey

	if err = a.msgPool.Put(message, false); err != nil {
		return "", nil
	}
	return message.Hash().String(), nil
}

func (a *Api) SendToken(from, receiver, token string, amount, fess, nonce, timestamp uint64, name, abbr string, increase bool, signature, public string) (string, error) {
	message := message.NewToken(from, receiver, token, amount, fess, nonce, timestamp, name, abbr, increase)

	signatureBytes, err := hex.DecodeString(signature)
	if err != nil {
		return "", err
	}
	pubKey, err := hex.DecodeString(public)
	if err != nil {
		return "", err
	}
	message.Header.Signature.Bytes = signatureBytes
	message.Header.Signature.PubKey = pubKey

	if err = a.msgPool.Put(message, false); err != nil {
		return "", err
	}
	return message.Hash().String(), nil
}

func (a *Api) CreateToken(from, receiver, token string, amount, fess, nonce, timestamp uint64, name, abbr string, increase bool) *chaintypes.Message {
	return message.NewToken(from, receiver, token, amount, fess, nonce, timestamp, name, abbr, increase)
}

func (a *Api) ContractMethod(contract, function string, params []string) (interface{}, error) {
	return a.runner.ReadMethod(contract, function, params)
}

func (a *Api) GetContractBySymbol(symbol string) (string, error) {
	address, exist := a.status.SymbolContract(symbol)
	if !exist {
		return "", fmt.Errorf("%s does no exist", symbol)
	}
	return address.String(), nil
}

type Token struct {
	Symbol   string `json:"symbol"`
	Contract string `json:"contract"`
}

func (a *Api) TokenList() []*Token {
	tokens := []*Token{}
	listMap := a.status.TokenList()
	for _, token := range listMap {
		for key, value := range token {
			tokens = append(tokens, &Token{
				Symbol:   key,
				Contract: value,
			})
		}
	}
	return tokens
}
