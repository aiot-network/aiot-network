package rpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aiot-network/aiotchain/chain/rpc/rpchttp"
	"github.com/aiot-network/aiotchain/chain/runner"
	"github.com/aiot-network/aiotchain/common/blockchain"
	"github.com/aiot-network/aiotchain/common/config"
	"github.com/aiot-network/aiotchain/common/param"
	"github.com/aiot-network/aiotchain/common/status"
	"github.com/aiot-network/aiotchain/service/peers"
	"github.com/aiot-network/aiotchain/service/pool"
	"github.com/aiot-network/aiotchain/tools/crypto/certgen"
	log "github.com/aiot-network/aiotchain/tools/log/log15"
	"github.com/aiot-network/aiotchain/tools/utils"
	"github.com/aiot-network/aiotchain/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"net"
	"os"
	"strconv"
)

const module = "rpc"

type Rpc struct {
	grpcServer *grpc.Server
	httpServer *rpchttp.RpcServer
	api        *Api
}

func NewRpc(status status.IStatus, msgPool *pool.Pool, chain blockchain.IChain, peers *peers.Peers, runner *runner.ContractRunner) *Rpc {
	return &Rpc{api: NewApi(status, msgPool, chain, peers, runner)}
}

func (r *Rpc) Name() string {
	return module
}

func (r *Rpc) Start() error {
	var err error
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

	endPoint := "0.0.0.0:" + config.Param.HttpPort
	r.httpServer, err = rpchttp.NewRPCServer(&config.RpcConfig{
		RpcIp:      config.Param.RpcIp,
		RpcPort:    config.Param.HttpPort,
		RpcTLS:     config.Param.RpcTLS,
		RpcCert:    config.Param.RpcCert,
		RpcCertKey: config.Param.RpcCertKey,
		RpcUser:    config.Param.RpcUser,
		RpcPass:    config.Param.RpcPass,
	})

	// Register all the APIs exposed by the services
	for _, api := range r.APIs() {
		if err := r.httpServer.RegisterService("rpc", api.Service); err != nil {
			return err
		}
	}
	if err := r.httpServer.Start([]string{endPoint}); err != nil {
		return err
	}

	if config.Param.RpcTLS {
		log.Info("Rpc startup", "module", module, "port", config.Param.RpcPort, "pem", config.Param.RpcCert)
	} else {
		log.Info("Rpc startup", "module", module, "port", config.Param.RpcPort)
	}
	return nil
}

func (r *Rpc) Stop() error {
	r.grpcServer.Stop()
	log.Info("Rpc was stopped", "module", module)
	return nil
}

func (r *Rpc) APIs() []rpchttp.API {
	return []rpchttp.API{
		{
			NameSpace: "rpc",
			Service:   r.api,
			Public:    true,
		},
	}
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

func (r *Rpc) RegisterLocalInfo(f func() *types.Local) {
	r.api.getLocal = f
}

func (r *Rpc) GetAccount(_ context.Context, req *AddressReq) (*Response, error) {
	account, err := r.api.GetAccount(req.Address)
	if err != nil {
		return NewResponse(Err_Params, nil, err.Error()), nil
	}
	bytes, _ := json.Marshal(account)
	return NewResponse(Success, bytes, ""), nil
}

func (r *Rpc) SendMessageRaw(ctx context.Context, req *SendMessageCodeReq) (*Response, error) {
	hash, err := r.api.SendMessageRaw(req.Code)
	if err != nil {
		return NewResponse(Err_Params, nil, err.Error()), nil
	}
	return NewResponse(Success, []byte(fmt.Sprintf("send message raw %s success", hash)), ""), nil
}

func (r *Rpc) GetMessage(ctx context.Context, req *HashReq) (*Response, error) {
	msg, err := r.api.GetMessage(req.Hash)
	if err != nil {
		return NewResponse(Err_Chain, nil, err.Error()), nil
	}
	bytes, _ := json.Marshal(msg)

	return NewResponse(Success, bytes, ""), nil
}

func (r *Rpc) GetBlockHash(ctx context.Context, req *HashReq) (*Response, error) {
	block, err := r.api.GetBlockHash(req.Hash)
	if err != nil {
		return NewResponse(Err_Chain, nil, err.Error()), nil
	}
	bytes, _ := json.Marshal(block)
	return NewResponse(Success, bytes, ""), nil
}

func (r *Rpc) GetBlockHeight(ctx context.Context, req *HeightReq) (*Response, error) {
	block, err := r.api.GetBlockHeight(req.Height)
	if err != nil {
		return NewResponse(Err_Chain, nil, err.Error()), nil
	}
	bytes, _ := json.Marshal(block)
	return NewResponse(Success, bytes, ""), nil
}

func (r *Rpc) LastHeight(context.Context, *NullReq) (*Response, error) {
	height := r.api.LastHeight()
	sHeight := strconv.FormatUint(height, 10)
	return NewResponse(Success, []byte(sHeight), ""), nil
}

func (r *Rpc) Confirmed(context.Context, *NullReq) (*Response, error) {
	height := r.api.Confirmed()
	sHeight := strconv.FormatUint(height, 10)
	return NewResponse(Success, []byte(sHeight), ""), nil
}

func (r *Rpc) GetMsgPool(context.Context, *NullReq) (*Response, error) {
	bytes, _ := json.Marshal(r.api.msgPool)
	return NewResponse(Success, bytes, ""), nil
}

func (r *Rpc) Candidates(context.Context, *NullReq) (*Response, error) {
	cas, err := r.api.Candidates()
	if err != nil {
		return NewResponse(Err_Chain, nil, err.Error()), nil
	}
	bytes, _ := json.Marshal(cas)
	return NewResponse(Success, bytes, ""), nil
}

func (r *Rpc) GetCycleSupers(ctx context.Context, req *CycleReq) (*Response, error) {
	supers, err := r.api.GetCycleSupers(req.Cycle)
	if err != nil {
		return NewResponse(Err_Chain, nil, err.Error()), nil
	}
	bytes, _ := json.Marshal(supers)

	return NewResponse(Success, bytes, ""), nil
}

func (r *Rpc) GetSupersReward(ctx context.Context, in *CycleReq) (*Response, error) {
	rewords, err := r.api.GetSupersReward(in.Cycle)
	if err != nil {
		return NewResponse(Err_Chain, nil, err.Error()), nil
	}
	bytes, _ := json.Marshal(rewords)

	return NewResponse(Success, bytes, ""), nil
}

func (r *Rpc) Token(ctx context.Context, req *TokenAddressReq) (*Response, error) {
	token, err := r.api.Token(req.Token)
	if err != nil {
		return NewResponse(Err_Chain, nil, err.Error()), nil
	}
	bytes, _ := json.Marshal(token)
	return NewResponse(Success, bytes, ""), nil
}

func (r *Rpc) GetContract(ctx context.Context, in *AddressReq) (*Response, error) {
	contract, err := r.api.Token(in.Address)
	if err != nil {
		return NewResponse(Err_Chain, nil, err.Error()), nil
	}
	bytes, _ := json.Marshal(contract)
	return NewResponse(Success, bytes, ""), nil
}

func (r *Rpc) PeersInfo(context.Context, *NullReq) (*Response, error) {
	peersInfo := r.api.PeersInfo()
	bytes, _ := json.Marshal(peersInfo)
	return NewResponse(Success, bytes, ""), nil
}

func (r *Rpc) LocalInfo(context.Context, *NullReq) (*Response, error) {
	local, err := r.api.LocalInfo()
	if err != nil {
		return NewResponse(Err_Unknown, nil, err.Error()), nil
	}
	bytes, _ := json.Marshal(local)
	return NewResponse(Success, bytes, ""), nil
}

func (r *Rpc) GenerateAddress(ctx context.Context, req *GenerateReq) (*Response, error) {
	address, err := r.api.GenerateAddress(req.Network, req.Publickey)
	if err != nil {
		return NewResponse(Err_Unknown, nil, err.Error()), nil
	}
	return NewResponse(Success, []byte(address), ""), nil
}

func (r *Rpc) GenerateTokenAddress(ctx context.Context, req *GenerateTokenReq) (*Response, error) {
	address, err := r.api.GenerateTokenAddress(req.Network, req.Abbr)
	if err != nil {
		return NewResponse(Err_Unknown, nil, err.Error()), nil
	}
	return NewResponse(Success, []byte(address), ""), nil
}

func (r *Rpc) CreateTransaction(ctx context.Context, req *TransactionReq) (*Response, error) {
	message := r.api.CreateTransaction(req.From, req.To, req.Token, req.Amount, req.Fees, req.Nonce, req.Timestamp)
	bytes, _ := json.Marshal(message)
	return NewResponse(Success, bytes, ""), nil
}

func (r *Rpc) CreateToken(ctx context.Context, req *TokenReq) (*Response, error) {
	message := r.api.CreateToken(req.From, req.Receiver, req.Token, req.Amount, req.Fees, req.Nonce, req.Timestamp, req.Name, req.Abbr, req.Increase)
	bytes, _ := json.Marshal(message)
	return NewResponse(Success, bytes, ""), nil
}

func (r *Rpc) SendTransaction(ctx context.Context, req *TransactionReq) (*Response, error) {
	tx, err := r.api.SendTransaction(req.From, req.To, req.Token, req.Amount, req.Fees, req.Nonce, req.Timestamp, req.Signature, req.Publickey)
	if err != nil {
		return NewResponse(Err_Unknown, nil, err.Error()), nil
	}
	return NewResponse(Success, []byte(tx), ""), nil
}

func (r *Rpc) SendToken(ctx context.Context, req *TokenReq) (*Response, error) {
	tx, err := r.api.SendToken(req.From, req.Receiver, req.Token, req.Amount, req.Fees, req.Nonce, req.Timestamp, req.Name, req.Abbr, req.Increase, req.Signature, req.Publickey)
	if err != nil {
		return NewResponse(Err_Unknown, nil, err.Error()), nil
	}
	return NewResponse(Success, []byte(tx), ""), nil
}

func (r *Rpc) GetContractBySymbol(ctx context.Context, in *SymbolReq) (*Response, error) {
	address, err := r.api.GetContractBySymbol(in.Symbol)
	if err != nil {
		return NewResponse(Err_Unknown, nil, err.Error()), nil
	}
	return NewResponse(Success, []byte(address), ""), nil
}

func (r *Rpc) ContractMethod(ctx context.Context, in *MethodReq) (*Response, error) {
	result, err := r.api.ContractMethod(in.Contract, in.Method, in.Params)
	if err != nil {
		return NewResponse(Err_Unknown, nil, err.Error()), nil
	}
	bytes, _ := json.Marshal(result)
	return NewResponse(Success, bytes, ""), nil
}

func (r *Rpc) TokenList(ctx context.Context, in *NullReq) (*Response, error) {
	list := r.api.TokenList()
	bytes, _ := json.Marshal(list)
	return NewResponse(Success, []byte(bytes), ""), nil
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
