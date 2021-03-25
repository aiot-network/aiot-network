package command

import (
	"context"
	"github.com/aiot-network/aiot-network/chain/rpc"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"time"
)

func init() {
	nodeCmds := []*cobra.Command{
		MsgPoolCmd,
		LocalInfoCmd,
		PeerInfoCmd,
	}
	RootCmd.AddCommand(nodeCmds...)
	RootSubCmdGroups["node"] = nodeCmds
}

//GenerateCmd cpu mine block
var MsgPoolCmd = &cobra.Command{
	Use:     "MsgPool",
	Short:   "MsgPool; Get messages in the message pool;",
	Aliases: []string{"msgpool", "MP", "mp"},
	Example: `
	MsgPool 
	`,
	Args: cobra.MinimumNArgs(0),
	Run:  MsgPool,
}

func MsgPool(cmd *cobra.Command, args []string) {

	client, err := NewRpcClient()
	if err != nil {
		log.Error(cmd.Use+" err: ", err)
		return
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*20)
	defer cancel()

	resp, err := client.Gc.GetMsgPool(ctx, &rpc.NullReq{})
	if err != nil {
		log.Error(cmd.Use+" err: ", err)
		return
	}
	if resp.Code == 0 {
		output(string(resp.Result))
		return
	}
	outputRespError(cmd.Use, resp)
}

var PeerInfoCmd = &cobra.Command{
	Use:     "PeerInfo",
	Short:   "PeerInfo; Get peer info;",
	Aliases: []string{"peerinfo", "PI", "pi"},
	Example: `
	PeerInfo 
	`,
	Args: cobra.MinimumNArgs(0),
	Run:  PeerInfo,
}

func PeerInfo(cmd *cobra.Command, args []string) {
	client, err := NewRpcClient()
	if err != nil {
		log.Error(cmd.Use+" err: ", err)
		return
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*20)
	defer cancel()
	resp, err := client.Gc.PeersInfo(ctx, &rpc.NullReq{})
	if err != nil {
		log.Error(cmd.Use+" err: ", err)
		return
	}
	if resp.Code == 0 {
		output(string(resp.Result))
		return
	}
	outputRespError(cmd.Use, resp)
}

var LocalInfoCmd = &cobra.Command{
	Use:     "LocalInfo ;Get the current node information",
	Short:   "LocalInfo ;Get the current node information;",
	Aliases: []string{"localinfo", "LI", "li"},
	Example: `
	LocalInfo
	`,
	Args: cobra.MinimumNArgs(0),
	Run:  LocalInfo,
}

func LocalInfo(cmd *cobra.Command, args []string) {
	client, err := NewRpcClient()
	if err != nil {
		log.Error(cmd.Use+" err: ", err)
		return
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*20)
	defer cancel()
	resp, err := client.Gc.LocalInfo(ctx, &rpc.NullReq{})
	if err != nil {
		log.Error(cmd.Use+" err: ", err)
		return
	}
	if resp.Code == 0 {
		output(string(resp.Result))
		return
	}
	outputRespError(cmd.Use, resp)
}
