package command

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aiot-network/aiot-network/chain/common/keystore"
	"github.com/aiot-network/aiot-network/chain/common/kit"
	"github.com/aiot-network/aiot-network/chain/rpc"
	"github.com/aiot-network/aiot-network/chain/rpc/types"
	"github.com/aiot-network/aiot-network/tools/crypto/mnemonic"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"time"
)

func init() {
	accountCmds := []*cobra.Command{
		CreateCmd,
		GetAccountCmd,
		ShowAccountsCmd,
		DecryptPrivateCmd,
		MnemonicToAccountCmd,
	}

	RootCmd.AddCommand(accountCmds...)
	RootSubCmdGroups["account"] = accountCmds
}

var GetAccountCmd = &cobra.Command{
	Use:     "Account {address};Get account status;",
	Aliases: []string{"Account", "A", "a"},
	Short:   "Account {address};Get account status;",
	Example: `
	Account xC8RqvGNhQ8sEpKrBHqnxJQh2rrtiJCXZrH 
	`,
	Args: cobra.MinimumNArgs(1),
	Run:  Account,
}

func Account(cmd *cobra.Command, args []string) {
	account, err := AccountByRpc(args[0])
	if err != nil {
		outputError(cmd.Use, err)
		return
	}
	bytes, _ := json.Marshal(account)
	output(string(bytes))
	return
}

func AccountByRpc(addr string) (*types.Account, error) {
	client, err := NewRpcClient()
	if err != nil {
		return nil, err
	}
	defer client.Close()

	re := &rpc.AddressReq{Address: addr}
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*20)
	defer cancel()
	resp, err := client.Gc.GetAccount(ctx, re)
	if err != nil {
		return nil, err
	}
	if resp.Code == 0 {
		var account *types.Account
		if err := json.Unmarshal(resp.Result, &account); err != nil {
			return nil, err
		}
		if account.Address != addr {
			account.Address = addr
		}
		return account, nil
	} else {
		return nil, fmt.Errorf("err code :%d, message :%s", resp.Code, resp.Err)
	}

}

var CreateCmd = &cobra.Command{
	Use:     "Create {password}",
	Short:   "Create {password}; Create account;",
	Aliases: []string{"create", "C", "c"},
	Example: `
	Create  
		OR
	Create 123456
	`,
	Args: cobra.MinimumNArgs(0),
	Run:  Create,
}

func Create(cmd *cobra.Command, args []string) {
	var passWd []byte
	var err error
	if len(args) == 1 && args[0] != "" {
		passWd = []byte(args[0])
	} else {
		fmt.Println("please set account password, cannot exceed 32 bytes：")
		passWd, err = readPassWd()
		if err != nil {
			log.Error(cmd.Use+" err: ", fmt.Errorf("read password failed! %s", err.Error()))
			return
		}
	}
	if len(passWd) > 32 {
		log.Error(cmd.Use+" err: ", fmt.Errorf("password too long! "))
		return
	}
	entropy, err := kit.Entropy()
	if err != nil {
		log.Error(cmd.Use+" err: ", err)
		return
	}
	mnemonicStr, err := kit.Mnemonic(entropy)
	if err != nil {
		log.Error(cmd.Use+" err: ", err)
		return
	}
	key, err := kit.MnemonicToEc(mnemonicStr)
	if err != nil {
		log.Error(cmd.Use+" err: ", fmt.Errorf("generate secp256k1 key failed! %s", err.Error()))
		return
	}
	p2pId, err := kit.GenerateP2PID(key)
	if err != nil {
		log.Error(cmd.Use+" err: ", fmt.Errorf("generate p2p id failed! %s", err.Error()))
	}
	if j, err := keystore.GenerateKeyJson(Net, Cfg.KeystoreDir, key, mnemonicStr, passWd); err != nil {
		log.Error(cmd.Use+" err: ", fmt.Errorf("generate key failed! %s", err.Error()))
	} else {
		j.P2pId = p2pId.String()
		bytes, _ := json.Marshal(j)
		output(string(bytes))
	}
}

func readPassWd() ([]byte, error) {
	var passWd [33]byte

	n, err := os.Stdin.Read(passWd[:])
	if err != nil {
		return nil, err
	}
	if n <= 1 {
		return nil, errors.New("not read")
	}
	return passWd[:n-1], nil
}

var ShowAccountsCmd = &cobra.Command{
	Use:     "ListAccounts",
	Short:   "ListAccounts; List all account of the wallet;",
	Aliases: []string{"listaccounts", "LA", "la"},
	Example: `
	ListAccounts
	`,
	Args: cobra.MinimumNArgs(0),
	Run:  ListAccount,
}

func ListAccount(cmd *cobra.Command, args []string) {
	if addrList, err := keystore.ReadAllAccount(Cfg.KeystoreDir); err != nil {
		log.Error(cmd.Use+" err: ", fmt.Errorf("read account failed! %s", err.Error()))
	} else {
		bytes, _ := json.Marshal(addrList)
		output(string(bytes))
	}
}

var DecryptPrivateCmd = &cobra.Command{
	Use:     "DecryptPrivate {address} {password} {key file}；Decrypting account json file generates the private key and mnemonic;；",
	Short:   "DecryptPrivate {address} {password} {key file}; Decrypting account json file generates the private key and mnemonic;",
	Aliases: []string{"decryptprivate", "DP", "dp"},

	Example: `
	DecryptPrivate xCHiGPLCzgnrdTqjKABXZteAGVJu3jXLjnQ
		OR
	DecryptPrivate xCHiGPLCzgnrdTqjKABXZteAGVJu3jXLjnQ 123456
		OR
	DecryptPrivate xCHiGPLCzgnrdTqjKABXZteAGVJu3jXLjnQ 123456 xCHiGPLCzgnrdTqjKABXZteAGVJu3jXLjnQ.json
	`,
	Args: cobra.MinimumNArgs(1),
	Run:  DecryptPrivate,
}

func DecryptPrivate(cmd *cobra.Command, args []string) {
	var passWd []byte
	var keyFile string
	var err error
	if len(args) >= 2 && args[1] != "" {
		passWd = []byte(args[1])
	} else {
		fmt.Println("please input password：")
		passWd, err = readPassWd()
		if err != nil {
			log.Error(cmd.Use+" err: ", fmt.Errorf("read password failed! %s", err.Error()))
			return
		}
	}
	if len(args) == 3 && args[2] != "" {
		keyFile = args[2]
	} else {
		keyFile = getAddJsonPath(args[0])
	}

	privKey, err := loadPrivate(keyFile, passWd)
	if err != nil {
		log.Error(cmd.Use+" err: ", fmt.Errorf("wrong password"))
		return
	}

	bytes, _ := json.Marshal(privKey)
	output(string(bytes))
}

var MnemonicToAccountCmd = &cobra.Command{
	Use:     "MnemonicToAccount {mnemonic} {password}；Restore address by mnemonic and set new password;",
	Short:   "MnemonicToAccount {mnemonic} {password}; Restore address by mnemonic and set new password;",
	Aliases: []string{"mnemonictoaccount", "MTA", "mta"},
	Example: `
	MnemonicToAccount "sadness ladder sister camp suspect sting height diagram confirm program twist ostrich blush bronze pass gasp resist random nothing recycle husband install business turtle"
		OR
	MnemonicToAccount "sadness ladder sister camp suspect sting height diagram confirm program twist ostrich blush bronze pass gasp resist random nothing recycle husband install business turtle" 123456
	`,
	Args: cobra.MinimumNArgs(1),
	Run:  MnemonicToAccount,
}

func MnemonicToAccount(cmd *cobra.Command, args []string) {
	var passWd []byte
	var err error
	priv, err := mnemonic.MnemonicToEc(args[0])
	if err != nil {
		log.Error(cmd.Use+" err: ", errors.New("[mnemonic] wrong"))
		return
	}
	if len(args) == 2 && args[1] != "" {
		passWd = []byte(args[1])
	} else {
		fmt.Println("please set address password, cannot exceed 32 bytes：")
		passWd, err = readPassWd()
		if err != nil {
			log.Error(cmd.Use+" err: ", fmt.Errorf("read pass word failed! %s", err.Error()))
			return
		}
	}
	if len(passWd) > 32 {
		log.Error(cmd.Use+" err: ", fmt.Errorf("password too long! "))
		return
	}
	p2pId, err := kit.GenerateP2PID(priv)
	if err != nil {
		log.Error(cmd.Use+" err: ", fmt.Errorf("generate p2p id failed! %s", err.Error()))
	}
	if j, err := keystore.GenerateKeyJson(Net, Cfg.KeystoreDir, priv, args[0], passWd); err != nil {
		log.Error(cmd.Use+" err: ", fmt.Errorf("generate key failed! %s", err.Error()))
	} else {
		j.P2pId = p2pId.String()
		bytes, _ := json.Marshal(j)
		output(string(bytes))
	}
}

func getAddJsonPath(addr string) string {
	return Cfg.KeystoreDir + "/" + addr + ".json"
}

func loadPrivate(jsonFile string, password []byte) (*keystore.Private, error) {
	j, err := keystore.ReadJson(jsonFile)
	if err != nil {
		return nil, err
	}
	return keystore.DecryptPrivate(password, j)
}
