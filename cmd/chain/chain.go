package main

import (
	"fmt"
	"github.com/aiot-network/aiotchain/chain/common/blockchain"
	chaindpos "github.com/aiot-network/aiotchain/chain/common/dpos"
	"github.com/aiot-network/aiotchain/chain/common/msglist"
	"github.com/aiot-network/aiotchain/chain/common/private"
	chainstatus "github.com/aiot-network/aiotchain/chain/common/status"
	"github.com/aiot-network/aiotchain/chain/common/status/act_status"
	"github.com/aiot-network/aiotchain/chain/common/status/dpos_status"
	"github.com/aiot-network/aiotchain/chain/common/status/token_status"
	"github.com/aiot-network/aiotchain/chain/node"
	"github.com/aiot-network/aiotchain/chain/request"
	"github.com/aiot-network/aiotchain/chain/rpc"
	"github.com/aiot-network/aiotchain/common/config"
	"github.com/aiot-network/aiotchain/common/horn"
	"github.com/aiot-network/aiotchain/service/generate"
	"github.com/aiot-network/aiotchain/service/gorutinue"
	"github.com/aiot-network/aiotchain/service/p2p"
	"github.com/aiot-network/aiotchain/service/peers"
	"github.com/aiot-network/aiotchain/service/pool"
	sync_service "github.com/aiot-network/aiotchain/service/sync"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
)

// interruptSignals defines the default signals to catch in order to do a proper
// shutdown.  This may be modified during init depending on the platform.
var interruptSignals = []os.Signal{
	os.Interrupt,
	os.Kill,
	syscall.SIGINT,
	syscall.SIGTERM,
}

func main() {
	// Initialize the goroutine count,  Use all processor cores.
	runtime.GOMAXPROCS(runtime.NumCPU())
	// Work around defer not working after os.Exit()
	if err := ChainMain(); err != nil {
		fmt.Println("Failed to start, ", err)
		os.Exit(1)
	}
}

// main start the node function
func ChainMain() error {
	var node *node.Node
	var err error
	wg := sync.WaitGroup{}
	wg.Add(1)

	if node, err = createNode(); err != nil {
		return err
	}
	if err := node.Start(); err != nil {
		return err
	}
	c := make(chan os.Signal, 1)
	signal.Notify(c, interruptSignals...)

	// Listen for initial shutdown signal and close the returned
	// channel to notify the caller.
	go func() {
		<-c
		node.Stop()
		close(c)
		wg.Done()
	}()
	wg.Wait()
	return nil
}

func createNode() (*node.Node, error) {
	err := config.LoadParam(private.NewPrivate(nil))
	if err != nil {
		return nil, err
	}
	actStatus, err := act_status.NewActStatus()
	if err != nil {
		return nil, err
	}
	dPosStatus, err := dpos_status.NewDPosStatus()
	if err != nil {
		return nil, err
	}
	tokenStatus, err := token_status.NewTokenStatus()
	if err != nil {
		return nil, err
	}

	dPos := chaindpos.NewDPos(dPosStatus)
	status := chainstatus.NewStatus(actStatus, dPosStatus, tokenStatus)
	gPool := gorutinue.NewPool()
	chain, err := blockchain.NewChain(status, dPos)
	if err != nil {
		return nil, err
	}
	if config.Param.RollBack != 0 {
		if err := chain.RollbackTo(config.Param.RollBack); err != nil {
			return nil, err
		}
		os.Exit(0)
	}

	reqHandler := request.NewRequestHandler(chain)
	peersSv := peers.NewPeers(reqHandler)

	p2pSv, err := p2p.NewP2p(peersSv, reqHandler)
	if err != nil {
		return nil, err
	}

	horn := horn.NewHorn(peersSv, gPool, reqHandler)
	msgManage, err := msglist.NewMsgManagement(status, actStatus)
	if err != nil {
		return nil, err
	}
	poolSv := pool.NewPool(horn, msgManage)

	rpcSv := rpc.NewRpc(status, poolSv, chain, peersSv)
	syncSv := sync_service.NewSync(peersSv, dPosStatus, reqHandler, chain)
	generateSv := generate.NewGenerate(chain, dPos, poolSv, horn)
	node := node.NewNode()

	rpcSv.RegisterLocalInfo(node.LocalInfo)
	reqHandler.RegisterLocalInfo(node.LocalInfo)

	chain.RegisterMsgPoolDeleteFunc(poolSv.Delete)

	// Register peer nodes to send blocks and message processing
	reqHandler.RegisterReceiveMessage(poolSv.ReceiveMsgFromPeer)
	reqHandler.RegisterReceiveBlock(syncSv.ReceivedBlockFromPeer)

	node.Register(syncSv)
	node.Register(peersSv)
	node.Register(p2pSv)
	node.Register(rpcSv)
	node.Register(reqHandler)
	node.Register(gPool)
	node.Register(poolSv)
	node.Register(generateSv)
	return node, nil
}
