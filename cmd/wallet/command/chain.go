package command

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aiot-network/aiotchain/chain/common/kit"
	"github.com/aiot-network/aiotchain/chain/common/kit/message"
	"github.com/aiot-network/aiotchain/chain/rpc"
	"github.com/aiot-network/aiotchain/chain/types"
	amount2 "github.com/aiot-network/aiotchain/tools/amount"
	"github.com/aiot-network/aiotchain/tools/crypto/ecc/secp256k1"
	"github.com/spf13/cobra"
	"strconv"
	"strings"
	"time"
)

func init() {
	blockCmds := []*cobra.Command{
		LastHeightCmd,
		GetBlockCmd,
		GetMessageCmd,
		SendMessageCmd,
		SendDerivedTransactionCmd,
	}

	RootCmd.AddCommand(blockCmds...)
	RootSubCmdGroups["chain"] = blockCmds
}

var LastHeightCmd = &cobra.Command{
	Use:     "LastHeight",
	Short:   "LastHeight; Get last height of node;",
	Aliases: []string{"lastheight", "LH", "lh"},
	Example: `
	LastHeight 
	`,
	Args: cobra.MinimumNArgs(0),
	Run:  LastHeight,
}

func LastHeight(cmd *cobra.Command, args []string) {
	client, err := NewRpcClient()
	if err != nil {
		outputError(cmd.Use, err)
		return
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*20)
	defer cancel()
	resp, err := client.Gc.LastHeight(ctx, &rpc.NullReq{})
	if err != nil {
		outputError(cmd.Use, err)
		return
	}
	if resp.Code == 0 {
		output(string(resp.Result))
		return
	}
	outputRespError(cmd.Use, resp)
}

var GetBlockCmd = &cobra.Command{
	Use:     "GetBlock {height/hash};",
	Short:   "GetBlock {height/hash}; Get block by height or hash;",
	Aliases: []string{"getblock", "gb", "GB"},
	Example: `
	GetBlock 1 
	GetBlock 0x4e32b712330c0d4ee45f06017390c5d1d3c26d0e6c7be4ea9a5036bdb6c72a07 
	`,
	Args: cobra.MinimumNArgs(1),
	Run:  GetBlock,
}

func GetBlock(cmd *cobra.Command, args []string) {
	height, err := strconv.ParseUint(args[0], 10, 64)
	if err != nil {
		GetBlockByHash(cmd, args)
		return
	}
	client, err := NewRpcClient()
	if err != nil {
		outputError(cmd.Use, err)
		return
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*20)
	defer cancel()

	resp, err := client.Gc.GetBlockHeight(ctx, &rpc.HeightReq{Height: height})
	if err != nil {
		outputError(cmd.Use, err)
		return
	}
	if resp.Code == 0 {
		output(string(resp.Result))
		return
	}
	outputRespError(cmd.Use, resp)

}

func GetBlockByHash(cmd *cobra.Command, args []string) {
	var err error
	client, err := NewRpcClient()
	if err != nil {
		outputError(cmd.Use, err)
		return
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*20)
	defer cancel()

	resp, err := client.Gc.GetBlockHash(ctx, &rpc.HashReq{Hash: args[0]})
	if err != nil {
		outputError(cmd.Use, err)
		return
	}
	if resp.Code == 0 {
		output(string(resp.Result))
		return
	}
	outputRespError(cmd.Use, resp)

}

var SendMessageCmd = &cobra.Command{
	Use:     "SendTransaction {from} {token} {to:amount|{to:amount}} {fees} {password} {nonce}; Send a transaction;",
	Aliases: []string{"sendtransaction", "ST", "st"},
	Short:   "SendTransaction {from} {token} {to:amount|to:amount} {fees} {password} {nonce}; Send a transaction;",
	Example: `
	SendTransaction xCHiGPLCzgnrdTqjKABXZteAGVJu3jXLjnQ FC xCE9boXz2TxSE9srVPDdfszyiXtfT3vduc8:10|xCHiGPLCzgnrdTqjKABXZteAGVJu3jXLjnQ:10 0.1
		OR
	SendTransaction xCHiGPLCzgnrdTqjKABXZteAGVJu3jXLjnQ FC xCE9boXz2TxSE9srVPDdfszyiXtfT3vduc8:10|xCHiGPLCzgnrdTqjKABXZteAGVJu3jXLjnQ:10 0.1 123456
		OR
	SendTransaction xCHiGPLCzgnrdTqjKABXZteAGVJu3jXLjnQ FC xCE9boXz2TxSE9srVPDdfszyiXtfT3vduc8:10 123456 1
	`,
	Args: cobra.MinimumNArgs(5),
	Run:  SendTransaction,
}

func SendTransaction(cmd *cobra.Command, args []string) {
	var passwd []byte
	var err error
	if len(args) > 4 {
		passwd = []byte(args[4])
	} else {
		fmt.Println("please input password：")
		passwd, err = readPassWd()
		if err != nil {
			outputError(cmd.Use, fmt.Errorf("read password failed! %s", err.Error()))
			return
		}
	}
	privKey, err := loadPrivate(getAddJsonPath(args[0]), passwd)
	if err != nil {
		outputError(cmd.Use, fmt.Errorf("wrong password"))
		return
	}

	tx, err := parseTransaction(args)
	if err != nil {
		outputError(cmd.Use, err)
		return
	}
	account, err := AccountByRpc(tx.From().String())
	if err != nil {
		outputError(cmd.Use, err)
		return
	}
	if tx.Header.Nonce == 0 {
		tx.Header.Nonce = account.Nonce + 1
	}
	if err := signMsg(tx, privKey.Private); err != nil {
		outputError(cmd.Use, errors.New("signature failure"))
		return
	}

	rs, err := sendMsg(tx)
	if err != nil {
		outputError(cmd.Use, err)
	} else if rs.Code != 0 {
		outputRespError(cmd.Use, rs)
	} else {
		fmt.Println()
		fmt.Println(string(rs.Result))
	}
}

func parseTransaction(args []string) (*types.Message, error) {
	var err error
	var from, tos, token string
	var fee, nonce uint64
	from = args[0]
	token = args[1]
	tos = args[2]
	if fFees, err := strconv.ParseFloat(args[3], 64); err != nil {
		return nil, errors.New("[fees] wrong")
	} else {
		if fFees < 0 {
			return nil, errors.New("[fees] wrong")
		}
		if fee, err = amount2.NewAmount(fFees); err != nil {
			return nil, errors.New("[fees] wrong")
		}
	}
	if len(args) > 5 {
		nonce, err = strconv.ParseUint(args[5], 10, 64)
		if err != nil {
			return nil, errors.New("[nonce] wrong")
		}
	}
	toList, err := parseReceiver(tos)
	if err != nil {
		return nil, err
	}
	return message.NewTransaction(from, token, toList, fee, nonce, uint64(time.Now().Unix())), nil
}

func parseReceiver(toStr string) ([]map[string]uint64, error) {
	toList := []map[string]uint64{}
	receivers := strings.Split(toStr, "|")
	if len(receivers) == 0 {
		return nil, fmt.Errorf("no receiver")
	}
	for _, receiver := range receivers {
		strs := strings.Split(receiver, ":")
		if len(strs) != 2 {
			return nil, fmt.Errorf("wrong receiver %s", receiver)
		}
		fAmt, err := strconv.ParseFloat(strs[1], 64)
		if err != nil {
			return nil, fmt.Errorf("wrong receiver %s", receiver)
		}
		if amt, err := amount2.NewAmount(fAmt); err != nil {
			return nil, fmt.Errorf("wrong receiver %s", receiver)
		} else {
			toList = append(toList, map[string]uint64{strs[0]: amt})
		}
	}
	return toList, nil
}

var SendDerivedTransactionCmd = &cobra.Command{
	Use:     "SendDerivedTransaction {from} {index} {token} {to:amount|{to:amount}} {fees} {password} {nonce}; Send a transaction;",
	Aliases: []string{"sendderivedtransaction", "SDT", "sdt"},
	Short:   "SendDerivedTransaction {from} {index} {token} {to:amount|to:amount} {fees} {password} {nonce}; Send a transaction;",
	Example: `
	SendDerivedTransaction xCHiGPLCzgnrdTqjKABXZteAGVJu3jXLjnQ 1 FC xCE9boXz2TxSE9srVPDdfszyiXtfT3vduc8:10|xCHiGPLCzgnrdTqjKABXZteAGVJu3jXLjnQ:10 0.1
		OR
	SendDerivedTransaction xCHiGPLCzgnrdTqjKABXZteAGVJu3jXLjnQ 1 FC xCE9boXz2TxSE9srVPDdfszyiXtfT3vduc8:10|xCHiGPLCzgnrdTqjKABXZteAGVJu3jXLjnQ:10 0.1 123456
		OR
	SendDerivedTransaction xCHiGPLCzgnrdTqjKABXZteAGVJu3jXLjnQ 1 FC xCE9boXz2TxSE9srVPDdfszyiXtfT3vduc8:10 123456 1
	`,
	Args: cobra.MinimumNArgs(6),
	Run:  SendDerivedTransaction,
}

func SendDerivedTransaction(cmd *cobra.Command, args []string) {
	var passwd []byte
	var err error
	if len(args) > 5 {
		passwd = []byte(args[5])
	} else {
		fmt.Println("please input password：")
		passwd, err = readPassWd()
		if err != nil {
			outputError(cmd.Use, fmt.Errorf("read password failed! %s", err.Error()))
			return
		}
	}
	m, err := getMnemonic(getAddJsonPath(args[0]), passwd)
	if err != nil {
		outputError(cmd.Use, fmt.Errorf("wrong password"))
		return
	}
	seed := kit.MnemonicToSeed(m)

	tx, privStr, err := parseHDTransaction(args, seed)
	if err != nil {
		outputError(cmd.Use, err)
		return
	}
	fmt.Println(tx.From().String())
	account, err := AccountByRpc(tx.From().String())
	if err != nil {
		outputError(cmd.Use, err)
		return
	}
	if tx.Header.Nonce == 0 {
		tx.Header.Nonce = account.Nonce + 1
	}
	if err := signMsg(tx, privStr); err != nil {
		outputError(cmd.Use, errors.New("signature failure"))
		return
	}

	rs, err := sendMsg(tx)
	if err != nil {
		outputError(cmd.Use, err)
	} else if rs.Code != 0 {
		outputRespError(cmd.Use, rs)
	} else {
		fmt.Println()
		fmt.Println(string(rs.Result))
	}
}

func parseHDTransaction(args []string, entropy string) (*types.Message, string, error) {
	var err error
	var from, tos, token string
	var fee, nonce uint64
	index, err := strconv.Atoi(args[1])
	if err != nil {
		return nil, "", err
	}
	hdPri, err := kit.HdDerive(Net, entropy, uint32(index))
	if err != nil {
		return nil, "", err
	}
	hdPub, err := kit.HdPrivateToPublic(hdPri, Net)
	if err != nil {
		return nil, "", err
	}
	pub, err := kit.HdToEc(hdPub, Net)
	if err != nil {
		return nil, "", err
	}
	priv, err := kit.HdToEc(hdPri, Net)
	if err != nil {
		return nil, "", err
	}
	from, err = kit.GenerateAddress(Net, pub)
	if err != nil {
		return nil, "", err
	}
	token = args[2]
	tos = args[3]
	if fFees, err := strconv.ParseFloat(args[4], 64); err != nil {
		return nil, "", errors.New("[fees] wrong")
	} else {
		if fFees < 0 {
			return nil, "", errors.New("[fees] wrong")
		}
		if fee, err = amount2.NewAmount(fFees); err != nil {
			return nil, "", errors.New("[fees] wrong")
		}
	}
	if len(args) > 6 {
		nonce, err = strconv.ParseUint(args[6], 10, 64)
		if err != nil {
			return nil, "", errors.New("[nonce] wrong")
		}
	}
	toList, err := parseReceiver(tos)
	if err != nil {
		return nil, "", err
	}
	return message.NewTransaction(from, token, toList, fee, nonce, uint64(time.Now().Unix())), priv, nil
}

func signMsg(msg *types.Message, key string) error {
	msg.SetHash()
	priv, err := secp256k1.PrivKeyFromString(key)
	if err != nil {
		return errors.New("[key] wrong")
	}
	if err := msg.SignMessage(priv); err != nil {
		return errors.New("sign failed")
	}
	return nil
}

func sendMsg(msg *types.Message) (*rpc.Response, error) {
	rpcMsg, err := types.MsgToRpcMsg(msg)
	if err != nil {
		return nil, err
	}
	rpcClient, err := NewRpcClient()
	if err != nil {
		return nil, err
	}
	defer rpcClient.Close()

	jsonBytes, err := json.Marshal(rpcMsg)
	if err != nil {
		return nil, err
	}
	re := &rpc.SendMessageCodeReq{Code: jsonBytes}
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*20)
	defer cancel()

	resp, err := rpcClient.Gc.SendMessageRaw(ctx, re)
	if err != nil {
		return nil, err
	}
	return resp, nil

}

var GetMessageCmd = &cobra.Command{
	Use:     "GetMessage {msghash}; Get Message by hash;",
	Aliases: []string{"getmessage", "GM", "gm"},
	Short:   "GetMessage {msghash}; Get Message by hash;",
	Example: `
	GetMessage 0xef7b92e552dca02c97c9d596d1bf69d0044d95dec4cee0e6a20153e62bce893b
	`,
	Args: cobra.MinimumNArgs(1),
	Run:  GetMessage,
}

func GetMessage(cmd *cobra.Command, args []string) {
	resp, err := GetMessageRpc(args[0])
	if err != nil {
		outputError(cmd.Use, err)
		return
	}
	if resp.Code == 0 {
		output(string(resp.Result))
		return
	}
	outputRespError(cmd.Use, resp)
}

func GetMessageRpc(hashStr string) (*rpc.Response, error) {
	client, err := NewRpcClient()
	if err != nil {
		return nil, err
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*20)
	defer cancel()

	resp, err := client.Gc.GetMessage(ctx, &rpc.HashReq{Hash: hashStr})
	return resp, err
}
