package rpc

import (
	"context"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aiot-network/aiotchain/chain/common/kit"
	"github.com/aiot-network/aiotchain/chain/common/kit/message"
	rpctypes "github.com/aiot-network/aiotchain/chain/rpc/types"
	chaintypes "github.com/aiot-network/aiotchain/chain/types"
	"github.com/aiot-network/aiotchain/common/blockchain"
	"github.com/aiot-network/aiotchain/common/config"
	"github.com/aiot-network/aiotchain/common/param"
	"github.com/aiot-network/aiotchain/common/status"
	"github.com/aiot-network/aiotchain/service/peers"
	"github.com/aiot-network/aiotchain/service/pool"
	"github.com/aiot-network/aiotchain/tools/arry"
	"github.com/aiot-network/aiotchain/tools/crypto/certgen"
	log "github.com/aiot-network/aiotchain/tools/log/log15"
	"github.com/aiot-network/aiotchain/tools/utils"
	"github.com/aiot-network/aiotchain/types"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"golang.org/x/net/http2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const module = "rpc"

type Rpc struct {
	grpcServer *grpc.Server
	httpServer *http.Server
	status     status.IStatus
	msgPool    *pool.Pool
	chain      blockchain.IChain
	peers      *peers.Peers
	getLocal   func() *types.Local
}

func NewRpc(status status.IStatus, msgPool *pool.Pool, chain blockchain.IChain, peers *peers.Peers) *Rpc {
	return &Rpc{status: status, msgPool: msgPool, chain: chain, peers: peers}
}

func (r *Rpc) Name() string {
	return module
}

func (r *Rpc) Start() error {
	var err error
	var tlsConfig *tls.Config
	httpListen, err := net.Listen("tcp", ":"+config.Param.HttpPort)
	if err != nil {
		return err
	}
	endPoint := "127.0.0.1:" + config.Param.RpcPort
	lis, err := net.Listen("tcp", ":"+config.Param.RpcPort)
	if err != nil {
		return err
	}
	r.grpcServer, err = r.NewGRpcServer()
	if err != nil {
		return err
	}
	go func() {
		if err := r.grpcServer.Serve(lis); err != nil {
			log.Info("Rpc startup failed!", "module", module, "err", err)
			os.Exit(1)
			return
		}
	}()

	getWay, err := r.NewGetWay(endPoint)
	if err != nil {
		return err
	}

	mux := http.NewServeMux()
	mux.Handle("/", getWay)
	if config.Param.RpcTLS {
		tlsConfig, err = getTLSConfig(config.Param.RpcCert, config.Param.RpcCertKey)
		if err != nil {
			return err
		}
		httpListen = tls.NewListener(httpListen, tlsConfig)
	}

	r.httpServer = &http.Server{
		Addr:      ":" + config.Param.HttpPort,
		Handler:   GrpcHandlerFunc(r.grpcServer, mux),
		TLSConfig: tlsConfig,
	}

	go func() {
		if err := r.httpServer.Serve(httpListen); err != nil {
			log.Info("Rpc startup failed!", "module", module, "err", err)
			os.Exit(1)
			return
		}
		log.Info("http startup", "module", module, "port", config.Param.HttpPort)
	}()
	if config.Param.RpcTLS {
		log.Info("Rpc startup", "module", module, "port", config.Param.RpcPort, "pem", config.Param.RpcCert)
	} else {
		log.Info("Rpc startup", "module", module, "port", config.Param.RpcPort)
	}
	return nil
}

func getTLSConfig(certPemPath, certKeyPath string) (*tls.Config, error) {
	var certKeyPair *tls.Certificate
	cert, _ := ioutil.ReadFile(certPemPath)
	key, _ := ioutil.ReadFile(certKeyPath)

	pair, err := tls.X509KeyPair(cert, key)
	if err != nil {
		return nil, err
	}

	certKeyPair = &pair

	return &tls.Config{
		Certificates: []tls.Certificate{*certKeyPair},
		NextProtos:   []string{http2.NextProtoTLS},
	}, nil
}

func GrpcHandlerFunc(grpcServer *grpc.Server, otherHandler http.Handler) http.Handler {
	if otherHandler == nil {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			grpcServer.ServeHTTP(w, r)
		})
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, _ := r.BasicAuth()
		if user != config.Param.RpcUser || pass != config.Param.RpcPass {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(fmt.Sprintf("the token authentication information is invalid: username=%s, password=%s\n", user, pass)))
			return
		}
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
		} else {
			otherHandler.ServeHTTP(w, r)
		}
	})
}

func (r *Rpc) Stop() error {
	r.grpcServer.Stop()
	log.Info("Rpc was stopped", "module", module)
	return nil
}

func (r *Rpc) Info() map[string]interface{} {
	return make(map[string]interface{}, 0)
}

func (r *Rpc) NewGRpcServer() (*grpc.Server, error) {
	var opts []grpc.ServerOption
	opts = append(opts, grpc.UnaryInterceptor(r.interceptor))

	// If tls is configured, generate tls certificate
	if config.Param.RpcTLS {
		if err := r.certFile(); err != nil {
			return nil, err
		}
		transportCredentials, err := credentials.NewServerTLSFromFile(config.Param.RpcCert, config.Param.RpcCertKey)
		if err != nil {
			return nil, err
		}
		opts = append(opts, grpc.Creds(transportCredentials))

	}

	// Set the maximum number of bytes received and sent
	opts = append(opts, grpc.MaxRecvMsgSize(param.MaxReqBytes))
	opts = append(opts, grpc.MaxSendMsgSize(param.MaxReqBytes))
	server := grpc.NewServer(opts...)
	RegisterGreeterServer(server, r)
	reflection.Register(server)
	return server, nil
}

func (r *Rpc) NewGetWay(endPoint string) (*runtime.ServeMux, error) {
	dopts := []grpc.DialOption{}
	dopts = append(dopts, grpc.WithPerRPCCredentials(&customCredential{
		Username: config.Param.RpcUser,
		Password: config.Param.RpcPass,
		OpenTLS:  config.Param.RpcTLS,
	}))
	ctx := context.Background()
	if config.Param.RpcTLS {
		creds, err := credentials.NewClientTLSFromFile(config.Param.RpcCert, "")
		if err != nil {
			return nil, err
		}
		dopts = append(dopts, grpc.WithTransportCredentials(creds))
	} else {
		dopts = append(dopts, grpc.WithInsecure())
	}

	gwmux := runtime.NewServeMux()
	if err := RegisterGreeterHandlerFromEndpoint(ctx, gwmux, endPoint, dopts); err != nil {
		return nil, err
	}
	gwmux.GetForwardResponseOptions()
	return gwmux, nil
}

func (r *Rpc) RegisterLocalInfo(f func() *types.Local) {
	r.getLocal = f
}

func (r *Rpc) GetAccount(_ context.Context, address *AddressReq) (*Response, error) {
	arryAddr := arry.StringToAddress(address.Address)
	if !kit.CheckAddress(config.Param.Name, arryAddr.String()) {
		return NewResponse(Err_Params, nil, fmt.Sprintf("%s address check failed", address.Address)), nil
	}
	account := r.status.Account(arryAddr)

	bytes, _ := json.Marshal(rpctypes.ToRpcAccount(account.(*chaintypes.Account)))
	return NewResponse(Success, bytes, ""), nil
}

func (r *Rpc) SendMessageRaw(ctx context.Context, code *SendMessageCodeReq) (*Response, error) {
	var rpcMsg *chaintypes.RpcMessage
	if err := json.Unmarshal(code.Code, &rpcMsg); err != nil {
		return NewResponse(Err_Params, nil, err.Error()), nil
	}
	tx, err := chaintypes.RpcMsgToMsg(rpcMsg)
	if err != nil {
		return NewResponse(Err_Params, nil, err.Error()), nil
	}
	if err := r.msgPool.Put(tx, false); err != nil {
		return NewResponse(Err_MsgPool, nil, err.Error()), nil
	}
	return NewResponse(Success, []byte(fmt.Sprintf("send message raw %s success", tx.Hash().String())), ""), nil
}

func (r *Rpc) GetMessage(ctx context.Context, hash *HashReq) (*Response, error) {
	var msg types.IMessage
	var exist bool
	var err error
	var height uint64
	var confirmed bool
	hashArry, err := arry.StringToHash(hash.Hash)
	if err != nil {
		return NewResponse(Err_Params, nil, "wrong hash "+err.Error()), nil
	}
	msg, err = r.chain.GetMessage(hashArry)
	if err != nil {
		msg, exist = r.msgPool.GetMessage(hashArry)
		if !exist {
			return NewResponse(Err_Chain, nil, fmt.Sprintf("Message hash %s does not exist", hash.Hash)), nil
		}
		height = 0
		confirmed = false
	} else {
		index, err := r.chain.GetMessageIndex(hashArry)
		if err != nil {
			return NewResponse(Err_Chain, nil, fmt.Sprintf("%s is not exist", hash.Hash)), nil
		}
		confirmedHeight := r.chain.LastConfirmed()
		height = index.GetHeight()
		confirmed = confirmedHeight >= height
	}
	rpcMsg, _ := chaintypes.MsgToRpcMsg(msg.(*chaintypes.Message))
	rsMsg := &chaintypes.RpcMessageWithHeight{
		MsgHeader: rpcMsg.MsgHeader,
		MsgBody:   rpcMsg.MsgBody,
		Height:    height,
		Confirmed: confirmed,
	}

	bytes, _ := json.Marshal(rsMsg)

	return NewResponse(Success, bytes, ""), nil
}

func (r *Rpc) GetBlockHash(ctx context.Context, hash *HashReq) (*Response, error) {
	hashArry, err := arry.StringToHash(hash.Hash)
	if err != nil {
		return NewResponse(Err_Params, nil, "wrong hash"), nil
	}
	block, err := r.chain.GetBlockHash(hashArry)
	if err != nil {
		return NewResponse(Err_Chain, nil, err.Error()), nil
	}
	rpcBlock, err := rpctypes.BlockToRpcBlock(block.(*chaintypes.Block), r.chain.LastConfirmed())
	if err != nil {
		return NewResponse(Err_Chain, nil, err.Error()), nil
	}
	bytes, _ := json.Marshal(rpcBlock)
	return NewResponse(Success, bytes, ""), nil
}

func (r *Rpc) GetBlockHeight(ctx context.Context, height *HeightReq) (*Response, error) {
	block, err := r.chain.GetBlockHeight(height.Height)
	if err != nil {
		return NewResponse(Err_Chain, nil, err.Error()), nil
	}
	rpcBlock, err := rpctypes.BlockToRpcBlock(block.(*chaintypes.Block), r.chain.LastConfirmed())
	if err != nil {
		return NewResponse(Err_Chain, nil, err.Error()), nil
	}
	bytes, _ := json.Marshal(rpcBlock)
	return NewResponse(Success, bytes, ""), nil
}

func (r *Rpc) LastHeight(context.Context, *NullReq) (*Response, error) {
	height := r.chain.LastHeight()
	sHeight := strconv.FormatUint(height, 10)
	return NewResponse(Success, []byte(sHeight), ""), nil
}
func (r *Rpc) Confirmed(context.Context, *NullReq) (*Response, error) {
	height := r.chain.LastConfirmed()
	sHeight := strconv.FormatUint(height, 10)
	return NewResponse(Success, []byte(sHeight), ""), nil
}

func (r *Rpc) GetMsgPool(context.Context, *NullReq) (*Response, error) {
	preparedTxs, futureTxs := r.msgPool.All()
	txPoolTxs := rpctypes.MsgsToRpcMsgsPool(preparedTxs, futureTxs)
	bytes, _ := json.Marshal(txPoolTxs)
	return NewResponse(Success, bytes, ""), nil
}

func (r *Rpc) Candidates(context.Context, *NullReq) (*Response, error) {
	candidates := r.status.Candidates()
	if candidates == nil || candidates.Len() == 0 {
		return NewResponse(Err_DPos, nil, "no candidates"), nil
	}
	cas := candidates.(*chaintypes.Candidates)
	for i, can := range cas.Members {
		for _, v := range can.Voters {
			cas.Members[i].Weight += r.chain.Vote(v)
		}
	}
	bytes, _ := json.Marshal(rpctypes.CandidatesToRpcCandidates(cas))
	return NewResponse(Success, bytes, ""), nil
}

func (r *Rpc) GetCycleSupers(ctx context.Context, cycle *CycleReq) (*Response, error) {
	supers := r.status.CycleSupers(cycle.Cycle)
	if supers == nil {
		return NewResponse(Err_DPos, nil, "no supers"), nil
	}
	bytes, _ := json.Marshal(rpctypes.SupersToRpcCandidates(supers))

	return NewResponse(Success, bytes, ""), nil
}

func (r *Rpc) GetSupersReward(ctx context.Context, in *CycleReq) (*Response, error) {
	reword := r.status.CycleReword(in.Cycle)
	if reword == nil {
		return NewResponse(Err_DPos, nil, "no reword"), nil
	}
	bytes, _ := json.Marshal(rpctypes.ToRpcReword(reword))

	return NewResponse(Success, bytes, ""), nil
}

func (r *Rpc) Token(ctx context.Context, token *TokenAddressReq) (*Response, error) {
	iToken, err := r.status.Token(arry.StringToAddress(token.Token))
	if err != nil {
		return NewResponse(Err_Token, nil, fmt.Sprintf("token address %s is not exist", token.Token)), nil
	}
	bytes, _ := json.Marshal(rpctypes.TokenToRpcToken(iToken.(*chaintypes.TokenRecord)))
	return NewResponse(Success, bytes, ""), nil
}

func (r *Rpc) PeersInfo(context.Context, *NullReq) (*Response, error) {
	peersInfo := r.peers.PeersInfo()
	bytes, _ := json.Marshal(peersInfo)
	return NewResponse(Success, bytes, ""), nil
}
func (r *Rpc) LocalInfo(context.Context, *NullReq) (*Response, error) {
	if r.getLocal != nil {
		local := r.getLocal()
		bytes, _ := json.Marshal(local)
		return NewResponse(Success, bytes, ""), nil
	}
	return NewResponse(Err_Local, nil, "no local info"), nil
}

func (r *Rpc) GenerateAddress(ctx context.Context, req *GenerateReq) (*Response, error) {
	address, err := kit.GenerateAddress(req.Network, req.Publickey)
	if err != nil {
		return NewResponse(Err_Unknown, nil, err.Error()), nil
	}
	return NewResponse(Success, []byte(address), ""), nil
}

func (r *Rpc) GenerateTokenAddress(ctx context.Context, req *GenerateTokenReq) (*Response, error) {
	address, err := kit.GenerateTokenAddress(req.Network, req.Address, req.Abbr)
	if err != nil {
		return NewResponse(Err_Unknown, nil, err.Error()), nil
	}
	return NewResponse(Success, []byte(address), ""), nil
}

func (r *Rpc) CreateTransaction(ctx context.Context, req *TransactionReq) (*Response, error) {
	to := []map[string]uint64{
		map[string]uint64{req.To: req.Amount},
	}

	message := message.NewTransaction(req.From, req.Token, to, req.Fees, req.Nonce, req.Timestamp)
	bytes, _ := json.Marshal(message)
	return NewResponse(Success, bytes, ""), nil
}

func (r *Rpc) CreateToken(ctx context.Context, req *TokenReq) (*Response, error) {
	message := message.NewToken(req.From, req.Receiver, req.Token, req.Amount, req.Fees, req.Nonce, req.Timestamp, req.Name, req.Abbr, req.Increase)
	bytes, _ := json.Marshal(message)
	return NewResponse(Success, bytes, ""), nil
}

func (r *Rpc) CreateCandidate(ctx context.Context, req *CandidateReq) (*Response, error) {
	message := message.NewCandidate(req.From, req.P2Pid, req.Fees, req.Nonce, req.Timestamp)
	bytes, _ := json.Marshal(message)
	return NewResponse(Success, bytes, ""), nil
}

func (r *Rpc) CreateCancel(ctx context.Context, req *CancelReq) (*Response, error) {
	message := message.NewCancel(req.From, req.Fees, req.Nonce, req.Timestamp)
	bytes, _ := json.Marshal(message)
	return NewResponse(Success, bytes, ""), nil
}

func (r *Rpc) CreateVote(ctx context.Context, req *VoteReq) (*Response, error) {
	message := message.NewVote(req.From, req.To, req.Fees, req.Nonce, req.Timestamp)
	bytes, _ := json.Marshal(message)
	return NewResponse(Success, bytes, ""), nil
}

func (r *Rpc) SendTransaction(ctx context.Context, req *TransactionReq) (*Response, error) {
	to := []map[string]uint64{
		map[string]uint64{req.To: req.Amount},
	}
	message := message.NewTransaction(req.From, req.Token, to, req.Fees, req.Nonce, req.Timestamp)

	signature, err := hex.DecodeString(req.Signature)
	if err != nil {
		return NewResponse(Err_Params, nil, err.Error()), nil
	}
	pubKey, err := hex.DecodeString(req.Publickey)
	if err != nil {
		return NewResponse(Err_Params, nil, err.Error()), nil
	}
	message.Header.Signature.Bytes = signature
	message.Header.Signature.PubKey = pubKey

	if err := r.msgPool.Put(message, false); err != nil {
		return NewResponse(Err_MsgPool, nil, err.Error()), nil
	}
	return NewResponse(Success, []byte(message.Hash().String()), ""), nil
}

func (r *Rpc) SendToken(ctx context.Context, req *TokenReq) (*Response, error) {
	message := message.NewToken(req.From, req.Receiver, req.Token, req.Amount, req.Fees, req.Nonce, req.Timestamp, req.Name, req.Abbr, req.Increase)

	fmt.Println()
	x, _ := json.Marshal(message)
	fmt.Println(string(x))
	signature, err := hex.DecodeString(req.Signature)
	if err != nil {
		return NewResponse(Err_Params, nil, err.Error()), nil
	}
	pubKey, err := hex.DecodeString(req.Publickey)
	if err != nil {
		return NewResponse(Err_Params, nil, err.Error()), nil
	}
	message.Header.Signature.Bytes = signature
	message.Header.Signature.PubKey = pubKey

	if err := r.msgPool.Put(message, false); err != nil {
		return NewResponse(Err_MsgPool, nil, err.Error()), nil
	}
	return NewResponse(Success, []byte(message.Hash().String()), ""), nil
}

func (r *Rpc) SendCandidate(ctx context.Context, req *CandidateReq) (*Response, error) {
	message := message.NewCandidate(req.From, req.P2Pid, req.Fees, req.Nonce, req.Timestamp)
	signature, err := hex.DecodeString(req.Signature)
	if err != nil {
		return NewResponse(Err_Params, nil, err.Error()), nil
	}
	pubKey, err := hex.DecodeString(req.Publickey)
	if err != nil {
		return NewResponse(Err_Params, nil, err.Error()), nil
	}
	message.Header.Signature.Bytes = signature
	message.Header.Signature.PubKey = pubKey

	if err := r.msgPool.Put(message, false); err != nil {
		return NewResponse(Err_MsgPool, nil, err.Error()), nil
	}
	return NewResponse(Success, []byte(message.Hash().String()), ""), nil
}

func (r *Rpc) SendCancel(ctx context.Context, req *CancelReq) (*Response, error) {
	message := message.NewCancel(req.From, req.Fees, req.Nonce, req.Timestamp)
	signature, err := hex.DecodeString(req.Signature)
	if err != nil {
		return NewResponse(Err_Params, nil, err.Error()), nil
	}
	pubKey, err := hex.DecodeString(req.Publickey)
	if err != nil {
		return NewResponse(Err_Params, nil, err.Error()), nil
	}
	message.Header.Signature.Bytes = signature
	message.Header.Signature.PubKey = pubKey

	if err := r.msgPool.Put(message, false); err != nil {
		return NewResponse(Err_MsgPool, nil, err.Error()), nil
	}
	return NewResponse(Success, []byte(message.Hash().String()), ""), nil
}

func (r *Rpc) SendVote(ctx context.Context, req *VoteReq) (*Response, error) {
	message := message.NewVote(req.From, req.To, req.Fees, req.Nonce, req.Timestamp)
	signature, err := hex.DecodeString(req.Signature)
	if err != nil {
		return NewResponse(Err_Params, nil, err.Error()), nil
	}
	pubKey, err := hex.DecodeString(req.Publickey)
	if err != nil {
		return NewResponse(Err_Params, nil, err.Error()), nil
	}
	message.Header.Signature.Bytes = signature
	message.Header.Signature.PubKey = pubKey

	if err := r.msgPool.Put(message, false); err != nil {
		return NewResponse(Err_MsgPool, nil, err.Error()), nil
	}
	return NewResponse(Success, []byte(message.Hash().String()), ""), nil
}

func NewResponse(code int32, result []byte, err string) *Response {
	return &Response{Code: code, Result: result, Err: err}
}

// Authenticate rpc users
func (r *Rpc) auth(ctx context.Context) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return errors.New("no token authentication information")
	}
	var (
		password string
		username string
	)

	if val, ok := md["username"]; ok {
		username = val[0]
	}

	if val, ok := md["password"]; ok {
		password = val[0]
	}

	if username != config.Param.RpcUser {
		return fmt.Errorf("the token authentication information is invalid: username=%s, password=%s", username, password)
	}
	if password != config.Param.RpcPass {
		return fmt.Errorf("the token authentication information is invalid: username=%s, password=%s", username, password)
	}
	return nil
}

func (r *Rpc) interceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	err = r.auth(ctx)
	if err != nil {
		return
	}
	return handler(ctx, req)
}

func (r *Rpc) certFile() error {
	if config.Param.RpcCert == "" {
		config.Param.RpcCert = config.Param.Data + "/server.pem"
	}
	if config.Param.RpcCertKey == "" {
		config.Param.RpcCertKey = config.Param.Data + "/server.key"
	}
	if !utils.Exist(config.Param.RpcCert) || !utils.Exist(config.Param.RpcCertKey) {
		return certgen.GenCertPair(config.Param.RpcCert, config.Param.RpcCertKey)
	}
	return nil
}
