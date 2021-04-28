package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/aiot-network/aiotchain/chain/common/private"
	"github.com/aiot-network/aiotchain/service/p2p"
	"os"
	"sync"
)

func main() {
	var (
		port     = flag.String("port", "19563", "the port of start a bootstrap")
		keyFile  = flag.String("k", "", "bootstrap node key file")
		password = flag.String("p", "", "the decryption password for key file")
	)
	flag.Parse()
	StartBootStrap(*port, *keyFile, *password)
}

func StartBootStrap(port, keyFile, password string) {
	wg := sync.WaitGroup{}
	wg.Add(1)

	if keyFile == "" {
		flag.PrintDefaults()
		return
	}
	if password == "" {
		fmt.Println("please enter the password for the key file:")
		passWd, err := readPassWd()
		if err != nil {
			fmt.Printf("read password failed! %s\n", err.Error())
			return
		}
		password = string(passWd)
	}
	pri := private.NewPrivate(nil)
	err := pri.Load(keyFile, password)
	if err != nil {
		fmt.Printf("failed to load key file %s! %s\n", keyFile, err.Error())
		return
	}

	server, err := p2p.NewBoot(port, "0.0.0.0", pri.PrivateKey())

	if err := server.StartBoot(); err != nil {
		fmt.Printf("start p2p server failed! %v\n", err)
		return
	}
	wg.Wait()
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
