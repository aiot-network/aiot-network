package command

import (
	"context"
	"errors"
	"fmt"
	"github.com/aiot-network/aiotchain/chain/common/kit"
	"github.com/aiot-network/aiotchain/chain/common/kit/message"
	"github.com/aiot-network/aiotchain/chain/rpc"
	"github.com/aiot-network/aiotchain/chain/types"
	amount2 "github.com/aiot-network/aiotchain/tools/amount"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"strconv"
	"time"
)

func init() {
	contractCmds := []*cobra.Command{
		TokenCmd,
		SendCreateTokenCmd,
	}
	RootCmd.AddCommand(contractCmds...)
	RootSubCmdGroups["token"] = contractCmds

}

var SendCreateTokenCmd = &cobra.Command{
	Use:     "SendCreateToken {from} {to} {name} {shorthand} {allow increase} {amount} {fees} {password} {nonce}; Send and create token;",
	Aliases: []string{"sendcreatetoken", "sct"},
	Short:   "SendCreateToken {from} {to} {name} {shorthand} {allow increase} {amount} {fees} {password} {nonce}; Send and create token;",
	Example: `
	SendCreateToken 3ajDJUnMYDyzXLwefRfNp7yLcdmg3ULb9ndQ 3ajNkh7yVYkETL9JKvGx3aL2YVNrqksjCUUE "M token" MT false 1000 0.1
		OR
	SendCreateToken 3ajDJUnMYDyzXLwefRfNp7yLcdmg3ULb9ndQ 3ajNkh7yVYkETL9JKvGx3aL2YVNrqksjCUUE "M token" MT false 1000 0.1 123456
		OR
	SendCreateToken 3ajDJUnMYDyzXLwefRfNp7yLcdmg3ULb9ndQ 3ajNkh7yVYkETL9JKvGx3aL2YVNrqksjCUUE "M token" MT false 1000 0.1 123456 0
	`,
	Args: cobra.MinimumNArgs(7),
	Run:  SendCreateToken,
}

func SendCreateToken(cmd *cobra.Command, args []string) {
	var passwd []byte
	var err error
	if len(args) > 7 {
		passwd = []byte(args[7])
	} else {
		fmt.Println("please input password：")
		passwd, err = readPassWd()
		if err != nil {
			log.Error(cmd.Use+" err: ", fmt.Errorf("read password failed! %s", err.Error()))
			return
		}
	}
	privKey, err := loadPrivate(getAddJsonPath(args[0]), passwd)
	if err != nil {
		log.Error(cmd.Use+" err: ", fmt.Errorf("wrong password"))
		return
	}

	tokenMsg, err := parseToken(args)
	if err != nil {
		log.Error(cmd.Use+" err: ", err)
		return
	}
	account, err := AccountByRpc(tokenMsg.From().String())
	if err != nil {
		log.Error(cmd.Use+" err: ", err)
		return
	}
	if tokenMsg.Header.Nonce == 0 {
		tokenMsg.Header.Nonce = account.Nonce + 1
	}
	if err := signMsg(tokenMsg, privKey.Private); err != nil {
		log.Error(cmd.Use+" err: ", errors.New("signature failure"))
		return
	}

	rs, err := sendMsg(tokenMsg)
	if err != nil {
		log.Error(cmd.Use+" err: ", err)
	} else if rs.Code != 0 {
		log.Errorf(cmd.Use+" err: code %d, message: %s", rs.Code, rs.Err)
	} else {
		fmt.Println()
		fmt.Println(string(rs.Result))
	}
}

func parseToken(args []string) (*types.Message, error) {
	var err error
	var from, to, tokenAddr string
	var amount, fee, nonce uint64
	var name, shorthand string
	var allowIncrease bool
	from = args[0]
	to = args[1]
	name = args[2]
	shorthand = args[3]
	allowIncrease, err = strconv.ParseBool(args[4])
	if err != nil {
		return nil, errors.New("[allow increase] wrong, the correct is true or false")
	}
	if fAmount, err := strconv.ParseFloat(args[5], 64); err != nil {
		return nil, errors.New("[amount] wrong")
	} else {
		if fAmount < 0 {
			return nil, errors.New("[amount] wrong")
		}
		if amount, err = amount2.NewAmount(fAmount); err != nil {
			return nil, errors.New("[amount] wrong")
		}
	}
	tokenAddr, err = kit.GenerateTokenAddress(Net, from, shorthand)
	if err != nil {
		return nil, err
	}
	fmt.Println(kit.CheckTokenAddress(Net, tokenAddr))
	fmt.Println("token address is ", tokenAddr)

	if fFees, err := strconv.ParseFloat(args[6], 64); err != nil {
		return nil, errors.New("[fees] wrong")
	} else {
		if fFees < 0 {
			return nil, errors.New("[fees] wrong")
		}
		if fee, err = amount2.NewAmount(fFees); err != nil {
			return nil, errors.New("[fees] wrong")
		}
	}
	if len(args) > 8 {
		nonce, err = strconv.ParseUint(args[8], 10, 64)
		if err != nil {
			return nil, errors.New("[nonce] wrong")
		}
	}
	tokenMsg := message.NewToken(from, to, tokenAddr, amount, fee, nonce, uint64(time.Now().Unix()), name, shorthand, allowIncrease)
	return tokenMsg, nil
}

var TokenCmd = &cobra.Command{
	Use:     "Token {token address}; Get a token records;",
	Aliases: []string{"token", "T", "t"},
	Short:   "Token {token address}; Get a token records;",
	Example: `
	Token Tfb792w8YrJxqgWxBV8iqpHq5ntwDePkcbQ
	`,
	Args: cobra.MinimumNArgs(1),
	Run:  Token,
}

func Token(cmd *cobra.Command, args []string) {
	resp, err := GetTokenByRpc(args[0])
	if err != nil {
		log.Error(cmd.Use+" err: ", err)
		return
	}
	if resp.Code == 0 {
		output(string(resp.Result))
		return
	} else {
		outputRespError(cmd.Use, resp)
	}
}

func GetTokenByRpc(tokenAddr string) (*rpc.Response, error) {
	client, err := NewRpcClient()
	if err != nil {
		return nil, err
	}
	defer client.Close()

	re := &rpc.TokenAddressReq{Token: tokenAddr}
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*20)
	defer cancel()
	return client.Gc.Token(ctx, re)

}
