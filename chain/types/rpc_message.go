package types

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aiot-network/aiotchain/chain/types/functionbody/exchange_func"
	"github.com/aiot-network/aiotchain/chain/types/status"
	"github.com/aiot-network/aiotchain/tools/arry"
	"github.com/aiot-network/aiotchain/types"
)

type IRpcMessageBody interface {
}

type RpcMessageHeader struct {
	MsgHash   string        `json:"msghash"`
	Type      MessageType   `json:"type"`
	From      string        `json:"from"`
	Nonce     uint64        `json:"nonce"`
	Fee       uint64        `json:"fee"`
	Time      uint64        `json:"time"`
	Signature *RpcSignature `json:"signscript"`
}

type RpcMessage struct {
	MsgHeader *RpcMessageHeader `json:"msgheader"`
	MsgBody   IRpcMessageBody   `json:"msgbody"`
}

type RpcMessageWithHeight struct {
	MsgHeader *RpcMessageHeader `json:"msgheader"`
	MsgBody   IRpcMessageBody   `json:"msgbody"`
	Height    uint64            `json:"height"`
	Confirmed bool              `json:"confirmed"`
}

type RpcSignature struct {
	Signature string `json:"signature"`
	PubKey    string `json:"pubkey"`
}

func RpcMsgToMsg(rpcMsg *RpcMessage) (*Message, error) {
	var err error
	if rpcMsg.MsgHeader == nil {
		return nil, errors.New("message header is nil")
	}
	signScript, err := RpcSignatureToSignature(rpcMsg.MsgHeader.Signature)
	if err != nil {
		return nil, err
	}
	var msgBody types.IMessageBody
	switch rpcMsg.MsgHeader.Type {
	case Transaction:
		body := &RpcTransactionBody{}
		bytes, err := json.Marshal(rpcMsg.MsgBody)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(bytes, body)
		if err != nil {
			return nil, err
		}
		msgBody, err = RpcTransactionBodyToBody(body)
	case Token:
		body := &RpcTokenBody{}
		bytes, err := json.Marshal(rpcMsg.MsgBody)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(bytes, body)
		if err != nil {
			return nil, err
		}
		msgBody, err = RpcTokenBodyToBody(body)
	case TokenV2:
		body := &RpcTokenBody{}
		bytes, err := json.Marshal(rpcMsg.MsgBody)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(bytes, body)
		if err != nil {
			return nil, err
		}
		msgBody, err = RpcTokenBodyToV2Body(body)
	case Redemption:
		body := &RpcRedemptionBody{}
		bytes, err := json.Marshal(rpcMsg.MsgBody)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(bytes, body)
		if err != nil {
			return nil, err
		}
		msgBody, err = RpcRedemptionBodyToBody(body)
	case Contract:
		body := &RpcContractBody{}
		bytes, err := json.Marshal(rpcMsg.MsgBody)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(bytes, body)
		if err != nil {
			return nil, err
		}
		msgBody, err = RpcContractBodyToBody(body)
	case Candidate:
		body := &RpcCandidateBody{}
		bytes, err := json.Marshal(rpcMsg.MsgBody)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(bytes, body)
		if err != nil {
			return nil, err
		}
		msgBody, err = RpcCandidateBodyToBody(body)
	case Cancel:
		msgBody = &CancelBody{}
	case Vote:
		body := &RpcVoteBody{}
		bytes, err := json.Marshal(rpcMsg.MsgBody)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(bytes, body)
		if err != nil {
			return nil, err
		}
		msgBody, err = RpcVoteBodyToBody(body)
	case Work:
		body := &RpcWorkBody{}
		bytes, err := json.Marshal(rpcMsg.MsgBody)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(bytes, body)
		if err != nil {
			return nil, err
		}
		msgBody, err = RpcWorkBodyToBody(body)
	}
	if err != nil {
		return nil, err
	}
	hash, err := arry.StringToHash(rpcMsg.MsgHeader.MsgHash)
	if err != nil {
		return nil, fmt.Errorf("wrong message hash %s", rpcMsg.MsgHeader.MsgHash)
	}
	tx := &Message{
		Header: &MsgHeader{
			Hash:      hash,
			Type:      rpcMsg.MsgHeader.Type,
			From:      arry.StringToAddress(rpcMsg.MsgHeader.From),
			Nonce:     rpcMsg.MsgHeader.Nonce,
			Fee:       rpcMsg.MsgHeader.Fee,
			Time:      rpcMsg.MsgHeader.Time,
			Signature: signScript,
		},
		Body: msgBody,
	}
	return tx, nil
}

type GetContractState func(arry.Hash) types.IStatus

func MsgToRpcMsg(msg types.IMessage) (*RpcMessage, error) {
	rpcMsg := &RpcMessage{
		MsgHeader: &RpcMessageHeader{
			MsgHash: msg.Hash().String(),
			Type:    MessageType(msg.Type()),
			From:    addressToString(msg.From()),
			Nonce:   msg.Nonce(),
			Fee:     msg.Fee(),
			Time:    msg.Time(),
			Signature: &RpcSignature{
				Signature: msg.Signature(),
				PubKey:    msg.PublicKey(),
			}},
		MsgBody: nil,
	}
	switch MessageType(msg.Type()) {
	case Transaction:
		rpcRecis := []RpcReceiver{}
		for _, re := range msg.MsgBody().MsgTo().ReceiverList() {
			rpcRecis = append(rpcRecis, RpcReceiver{
				Address: re.Address.String(),
				Amount:  re.Amount,
			})
		}
		rpcMsg.MsgBody = RpcTransactionBody{
			Token:     msg.MsgBody().MsgContract().String(),
			Receivers: rpcRecis,
		}
	case Token:
		body, ok := msg.MsgBody().(*TokenBody)
		if !ok {
			return nil, errors.New("message type error")
		}

		rpcMsg.MsgBody = &RpcTokenBody{
			Address:        msg.MsgBody().MsgContract().String(),
			Receiver:       msg.MsgBody().MsgTo().ReceiverList()[0].Address.String(),
			Name:           body.Name,
			Shorthand:      body.Shorthand,
			IncreaseIssues: body.IncreaseIssues,
			Amount:         msg.MsgBody().MsgAmount(),
		}
	case TokenV2:
		body, ok := msg.MsgBody().(*TokenV2Body)
		if !ok {
			return nil, errors.New("message type error")
		}

		rpcMsg.MsgBody = &RpcTokenBody{
			Address:        msg.MsgBody().MsgContract().String(),
			Receiver:       msg.MsgBody().MsgTo().ReceiverList()[0].Address.String(),
			Name:           body.Name,
			Shorthand:      body.Shorthand,
			Amount:         msg.MsgBody().MsgAmount(),
			PledgeRate:     int(body.PledgeRate),
			IncreaseIssues: false,
		}
	case Redemption:
		body, ok := msg.MsgBody().(*RedemptionBody)
		if !ok {
			return nil, errors.New("message type error")
		}

		rpcMsg.MsgBody = &RpcRedemptionBody{
			Address:    msg.MsgBody().MsgContract().String(),
			PledgeRate: int(body.PledgeRate),
			Amount:     body.Amount,
		}
	case Contract:
		body, ok := msg.MsgBody().(*ContractBody)
		if !ok {
			return nil, errors.New("message type error")
		}

		msgBody, err := ContractBodyToRpcContract(body)
		if err != nil {
			return nil, errors.New("message body error")
		}
		rpcMsg.MsgBody = msgBody
	case Candidate:
		body, ok := msg.MsgBody().(*CandidateBody)
		if !ok {
			return nil, errors.New("message type error")
		}
		rpcMsg.MsgBody = &RpcCandidateBody{
			PeerId: body.Peer.String(),
		}
	case Cancel:
		rpcMsg.MsgBody = &RpcCancelBody{}
	case Vote:
		rpcMsg.MsgBody = &RpcVoteBody{To: msg.MsgBody().MsgTo().ReceiverList()[0].Address.String()}
	case Work:
		body, ok := msg.MsgBody().(*WorkBody)
		if !ok {
			return nil, errors.New("message type error")
		}
		list := []RpcAddressWork{}
		for _, w := range body.List {
			list = append(list, RpcAddressWork{
				Address:  w.Address.String(),
				Workload: w.Workload,
				EndTime:  w.EndTime,
			})
		}
		rpcMsg.MsgBody = &RpcWorkBody{
			StartTime: body.StartTime,
			EndTime:   body.EndTime,
			List:      list,
		}
	}

	return rpcMsg, nil
}

func MsgToRpcMsgWithState(msg types.IMessage, state *ContractStatus) (*RpcMessage, error) {
	rpcMsg := &RpcMessage{
		MsgHeader: &RpcMessageHeader{
			MsgHash: msg.Hash().String(),
			Type:    MessageType(msg.Type()),
			From:    addressToString(msg.From()),
			Nonce:   msg.Nonce(),
			Fee:     msg.Fee(),
			Time:    msg.Time(),
			Signature: &RpcSignature{
				Signature: msg.Signature(),
				PubKey:    msg.PublicKey(),
			}},
		MsgBody: nil,
	}
	switch MessageType(msg.Type()) {
	case Transaction:
		rpcRecis := []RpcReceiver{}
		for _, re := range msg.MsgBody().MsgTo().ReceiverList() {
			rpcRecis = append(rpcRecis, RpcReceiver{
				Address: re.Address.String(),
				Amount:  re.Amount,
			})
		}
		rpcMsg.MsgBody = RpcTransactionBody{
			Token:     msg.MsgBody().MsgContract().String(),
			Receivers: rpcRecis,
		}
	case Token:
		body, ok := msg.MsgBody().(*TokenBody)
		if !ok {
			return nil, errors.New("message type error")
		}

		rpcMsg.MsgBody = &RpcTokenBody{
			Address:        msg.MsgBody().MsgContract().String(),
			Receiver:       msg.MsgBody().MsgTo().ReceiverList()[0].Address.String(),
			Name:           body.Name,
			Shorthand:      body.Shorthand,
			IncreaseIssues: body.IncreaseIssues,
			Amount:         msg.MsgBody().MsgAmount(),
		}
	case TokenV2:
		body, ok := msg.MsgBody().(*TokenV2Body)
		if !ok {
			return nil, errors.New("message type error")
		}

		rpcMsg.MsgBody = &RpcTokenBody{
			Address:        msg.MsgBody().MsgContract().String(),
			Receiver:       msg.MsgBody().MsgTo().ReceiverList()[0].Address.String(),
			Name:           body.Name,
			Shorthand:      body.Shorthand,
			Amount:         msg.MsgBody().MsgAmount(),
			PledgeRate:     int(body.PledgeRate),
			IncreaseIssues: false,
		}
	case Redemption:
		body, ok := msg.MsgBody().(*RedemptionBody)
		if !ok {
			return nil, errors.New("message type error")
		}

		rpcMsg.MsgBody = &RpcRedemptionBody{
			Address:    msg.MsgBody().MsgContract().String(),
			PledgeRate: int(body.PledgeRate),
			Amount:     body.Amount,
		}
	case Contract:
		body, ok := msg.MsgBody().(*ContractBody)
		if !ok {
			return nil, errors.New("message type error")
		}

		msgBody, err := ContractBodyToRpcContractWithState(body, state)
		if err != nil {
			return nil, errors.New("message body error")
		}
		rpcMsg.MsgBody = msgBody
	case Candidate:
		body, ok := msg.MsgBody().(*CandidateBody)
		if !ok {
			return nil, errors.New("message type error")
		}
		rpcMsg.MsgBody = &RpcCandidateBody{
			PeerId: body.Peer.String(),
		}
	case Cancel:
		rpcMsg.MsgBody = &RpcCancelBody{}
	case Vote:
		rpcMsg.MsgBody = &RpcVoteBody{To: msg.MsgBody().MsgTo().ReceiverList()[0].Address.String()}
	case Work:
		body, ok := msg.MsgBody().(*WorkBody)
		if !ok {
			return nil, errors.New("message type error")
		}
		list := []RpcAddressWork{}
		for _, w := range body.List {
			list = append(list, RpcAddressWork{
				Address:  w.Address.String(),
				Workload: w.Workload,
				EndTime:  w.EndTime,
			})
		}
		rpcMsg.MsgBody = &RpcWorkBody{
			StartTime: body.StartTime,
			EndTime:   body.EndTime,
			List:      list,
		}
	}

	return rpcMsg, nil
}

func RpcSignatureToSignature(rpcSignScript *RpcSignature) (*Signature, error) {
	if rpcSignScript == nil {
		return nil, errors.New("signature is nil")
	}
	if rpcSignScript.Signature == "" || rpcSignScript.PubKey == "" {
		return nil, errors.New("signature content is nil")
	}
	signature, err := hex.DecodeString(rpcSignScript.Signature)
	if err != nil {
		return nil, err
	}
	pubKey, err := hex.DecodeString(rpcSignScript.PubKey)
	if err != nil {
		return nil, err
	}
	return &Signature{
		Bytes:  signature,
		PubKey: pubKey,
	}, nil
}

func RpcTransactionBodyToBody(rpcBody *RpcTransactionBody) (*TransactionBody, error) {
	if rpcBody == nil {
		return nil, errors.New("wrong transaction body")
	}
	recis := NewReceivers()
	for _, re := range rpcBody.Receivers {
		recis.Add(arry.StringToAddress(re.Address), re.Amount)
	}
	return &TransactionBody{
		TokenAddress: arry.StringToAddress(rpcBody.Token),
		Receivers:    recis,
	}, nil
}

func RpcTokenBodyToBody(rpcBody *RpcTokenBody) (*TokenBody, error) {
	if rpcBody == nil {
		return nil, errors.New("wrong token body")
	}
	return &TokenBody{
		TokenAddress:   arry.StringToAddress(rpcBody.Address),
		Receiver:       arry.StringToAddress(rpcBody.Receiver),
		Name:           rpcBody.Name,
		Shorthand:      rpcBody.Shorthand,
		IncreaseIssues: rpcBody.IncreaseIssues,
		Amount:         rpcBody.Amount,
	}, nil
}

func RpcTokenBodyToV2Body(rpcBody *RpcTokenBody) (*TokenV2Body, error) {
	if rpcBody == nil {
		return nil, errors.New("wrong token body")
	}
	return &TokenV2Body{
		TokenAddress: arry.StringToAddress(rpcBody.Address),
		Receiver:     arry.StringToAddress(rpcBody.Receiver),
		Name:         rpcBody.Name,
		Shorthand:    rpcBody.Shorthand,
		Amount:       rpcBody.Amount,
		PledgeRate:   PledgeRate(rpcBody.PledgeRate),
	}, nil
}

func RpcRedemptionBodyToBody(rpcBody *RpcRedemptionBody) (*RedemptionBody, error) {
	if rpcBody == nil {
		return nil, errors.New("wrong redemption body")
	}
	return &RedemptionBody{
		TokenAddress: arry.StringToAddress(rpcBody.Address),
		PledgeRate:   PledgeRate(rpcBody.PledgeRate),
		Amount:       rpcBody.Amount,
	}, nil
}

func RpcCandidateBodyToBody(rpcBody *RpcCandidateBody) (*CandidateBody, error) {
	if rpcBody == nil {
		return nil, errors.New("wrong candidate body")
	}
	body := &CandidateBody{}
	copy(body.Peer[:], rpcBody.PeerIdBytes())
	return body, nil
}

func RpcVoteBodyToBody(rpcBody *RpcVoteBody) (*VoteBody, error) {
	if rpcBody == nil {
		return nil, errors.New("wrong vote body")
	}

	return &VoteBody{To: arry.StringToAddress(rpcBody.To)}, nil
}

func RpcWorkBodyToBody(rpcBody *RpcWorkBody) (*WorkBody, error) {
	if rpcBody == nil {
		return nil, errors.New("work body is nil")
	}

	list := []AddressWork{}
	for _, work := range rpcBody.List {
		list = append(list, AddressWork{
			Address:  arry.StringToAddress(work.Address),
			Workload: work.Workload,
			EndTime:  work.EndTime,
		})
	}
	return &WorkBody{
		StartTime: rpcBody.StartTime,
		EndTime:   rpcBody.EndTime,
		List:      list,
	}, nil
}

func addressToString(address arry.Address) string {
	if address.IsEqual(CoinBase) {
		return CoinBase.String()
	}
	return address.String()
}

func ContractBodyToRpcContractWithState(body *ContractBody, contractState *ContractStatus) (*RpcContractBodyWithState, error) {
	var state *RpcContractState = &RpcContractState{
		StateCode: Status_Wait,
		Events:    make([]*RpcEvent, 0),
		Error:     "",
	}
	if contractState != nil {
		state.StateCode = contractState.State
		state.Error = contractState.Error
		if contractState.Event != nil {
			for _, e := range contractState.Event {
				state.Events = append(state.Events, &RpcEvent{
					EventType: int(e.EventType),
					From:      e.From.String(),
					To:        e.To.String(),
					Token:     e.Token.String(),
					Amount:    e.Amount,
					Height:    e.Height,
				})
			}
		}
	}

	funcBody, err := rpcFunction(body)
	if err != nil {
		return nil, err
	}
	return &RpcContractBodyWithState{
		Contract:     body.Contract.String(),
		Type:         body.Type,
		FunctionType: body.FunctionType,
		Function:     funcBody,
		State:        state,
	}, nil
}

func ContractBodyToRpcContract(body *ContractBody) (*RpcContractBody, error) {
	funcBody, err := rpcFunction(body)
	if err != nil {
		return nil, err
	}
	return &RpcContractBody{
		Contract:     body.Contract.String(),
		Type:         body.Type,
		FunctionType: body.FunctionType,
		Function:     funcBody,
	}, nil
}

func rpcFunction(body *ContractBody) (IRCFunction, error) {
	if body == nil {
		return nil, fmt.Errorf("invalid contract body")
	}
	var function IRCFunction
	switch body.FunctionType {
	case status.Exchange_Init:
		funcBody, ok := body.Function.(*exchange_func.ExchangeInitBody)
		if !ok {
			return nil, errors.New("wrong function body")
		}
		function = &RpcExchangeInitBody{
			Admin:  funcBody.Admin.String(),
			FeeTo:  funcBody.FeeTo.String(),
			Symbol: funcBody.Symbol,
		}
	case status.Exchange_SetAdmin:
		funcBody, ok := body.Function.(*exchange_func.ExchangeAdmin)
		if !ok {
			return nil, errors.New("wrong function body")
		}
		function = &RpcExchangeSetAdminBody{
			Address: funcBody.Address.String(),
		}
	case status.Exchange_SetFeeTo:
		funcBody, ok := body.Function.(*exchange_func.ExchangeFeeTo)
		if !ok {
			return nil, errors.New("wrong function body")
		}
		function = &RpcExchangeSetFeeToBody{
			Address: funcBody.Address.String(),
		}
	case status.Exchange_ExactIn:
		funcBody, ok := body.Function.(*exchange_func.ExactIn)
		if !ok {
			return nil, errors.New("wrong function body")
		}
		function = &RpcExchangeExactInBody{
			AmountIn:     funcBody.AmountIn,
			AmountOutMin: funcBody.AmountOutMin,
			Path:         hashAddrToAddr(funcBody.Path),
			To:           funcBody.To.String(),
			Deadline:     funcBody.Deadline,
		}
	case status.Exchange_ExactOut:
		funcBody, ok := body.Function.(*exchange_func.ExactOut)
		if !ok {
			return nil, errors.New("wrong function body")
		}
		function = &RpcExchangeExactOutBody{
			AmountOut:   funcBody.AmountOut,
			AmountInMax: funcBody.AmountInMax,
			Path:        hashAddrToAddr(funcBody.Path),
			To:          funcBody.To.String(),
			Deadline:    funcBody.Deadline,
		}
	case status.Pair_AddLiquidity:
		funcBody, ok := body.Function.(*exchange_func.ExchangeAddLiquidity)
		if !ok {
			return nil, errors.New("wrong function body")
		}
		function = &RpcExchangeAddLiquidity{
			Exchange:       funcBody.Exchange.String(),
			TokenA:         funcBody.TokenA.String(),
			TokenB:         funcBody.TokenB.String(),
			To:             funcBody.To.String(),
			AmountADesired: funcBody.AmountADesired,
			AmountBDesired: funcBody.AmountBDesired,
			AmountAMin:     funcBody.AmountAMin,
			AmountBMin:     funcBody.AmountBMin,
			Deadline:       funcBody.Deadline,
		}
	case status.Pair_RemoveLiquidity:
		funcBody, ok := body.Function.(*exchange_func.ExchangeRemoveLiquidity)
		if !ok {
			return nil, errors.New("wrong function body")
		}
		function = &RpcExchangeRemoveLiquidity{
			Exchange:   funcBody.Exchange.String(),
			TokenA:     funcBody.TokenA.String(),
			TokenB:     funcBody.TokenB.String(),
			To:         funcBody.To.String(),
			Liquidity:  funcBody.Liquidity,
			AmountAMin: funcBody.AmountAMin,
			AmountBMin: funcBody.AmountBMin,
			Deadline:   funcBody.Deadline,
		}
	}
	return function, nil
}

func RpcContractBodyToBody(rpcBody *RpcContractBody) (*ContractBody, error) {
	if rpcBody == nil {
		return nil, errors.New("wrong contract transaction body")
	}
	body := &RpcContractBody{}
	bytes, err := json.Marshal(rpcBody)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(bytes, body)
	if err != nil {
		return nil, err
	}
	switch body.FunctionType {
	case status.Exchange_Init:
		bytes, err := json.Marshal(body.Function)
		if err != nil {
			return nil, err
		}
		init := &RpcExchangeInitBody{
			Symbol: "",
			Admin:  "",
			FeeTo:  "",
		}
		err = json.Unmarshal(bytes, init)
		if err != nil {
			return nil, err
		}
		return &ContractBody{
			Contract:     arry.StringToAddress(body.Contract),
			Type:         body.Type,
			FunctionType: body.FunctionType,
			Function: &exchange_func.ExchangeInitBody{
				Admin:  arry.StringToAddress(init.Admin),
				FeeTo:  arry.StringToAddress(init.FeeTo),
				Symbol: init.Symbol,
			},
		}, nil
	case status.Exchange_SetAdmin:
		bytes, err := json.Marshal(body.Function)
		if err != nil {
			return nil, err
		}
		setBody := &RpcExchangeSetAdminBody{
			Address: "",
		}
		err = json.Unmarshal(bytes, setBody)
		if err != nil {
			return nil, err
		}
		return &ContractBody{
			Contract:     arry.StringToAddress(body.Contract),
			Type:         body.Type,
			FunctionType: body.FunctionType,
			Function: &exchange_func.ExchangeAdmin{
				Address: arry.StringToAddress(setBody.Address),
			},
		}, nil
	case status.Exchange_SetFeeTo:
		bytes, err := json.Marshal(body.Function)
		if err != nil {
			return nil, err
		}
		setBody := &RpcExchangeSetFeeToBody{
			Address: "",
		}
		err = json.Unmarshal(bytes, setBody)
		if err != nil {
			return nil, err
		}
		return &ContractBody{
			Contract:     arry.StringToAddress(body.Contract),
			Type:         body.Type,
			FunctionType: body.FunctionType,
			Function: &exchange_func.ExchangeFeeTo{
				Address: arry.StringToAddress(setBody.Address),
			},
		}, nil
	case status.Exchange_ExactIn:
		bytes, err := json.Marshal(body.Function)
		if err != nil {
			return nil, err
		}
		inBody := &RpcExchangeExactInBody{}
		err = json.Unmarshal(bytes, inBody)
		if err != nil {
			return nil, err
		}
		return &ContractBody{
			Contract:     arry.StringToAddress(body.Contract),
			Type:         body.Type,
			FunctionType: body.FunctionType,
			Function: &exchange_func.ExactIn{
				AmountIn:     inBody.AmountIn,
				AmountOutMin: inBody.AmountOutMin,
				Path:         addrListToHashAddr(inBody.Path),
				To:           arry.StringToAddress(inBody.To),
				Deadline:     inBody.Deadline,
			},
		}, nil
	case status.Exchange_ExactOut:
		bytes, err := json.Marshal(body.Function)
		if err != nil {
			return nil, err
		}
		outBody := &RpcExchangeExactOutBody{}
		err = json.Unmarshal(bytes, outBody)
		if err != nil {
			return nil, err
		}
		return &ContractBody{
			Contract:     arry.StringToAddress(body.Contract),
			Type:         body.Type,
			FunctionType: body.FunctionType,
			Function: &exchange_func.ExactOut{
				AmountOut:   outBody.AmountOut,
				AmountInMax: outBody.AmountInMax,
				Path:        addrListToHashAddr(outBody.Path),
				To:          arry.StringToAddress(outBody.To),
				Deadline:    outBody.Deadline,
			},
		}, nil
	case status.Pair_AddLiquidity:
		bytes, err := json.Marshal(body.Function)
		if err != nil {
			return nil, err
		}
		createBody := &RpcExchangeAddLiquidity{}
		err = json.Unmarshal(bytes, createBody)
		if err != nil {
			return nil, err
		}
		return &ContractBody{
			Contract:     arry.StringToAddress(body.Contract),
			Type:         body.Type,
			FunctionType: body.FunctionType,
			Function: &exchange_func.ExchangeAddLiquidity{
				Exchange:       arry.StringToAddress(createBody.Exchange),
				TokenA:         arry.StringToAddress(createBody.TokenA),
				TokenB:         arry.StringToAddress(createBody.TokenB),
				To:             arry.StringToAddress(createBody.To),
				AmountADesired: createBody.AmountADesired,
				AmountBDesired: createBody.AmountBDesired,
				AmountAMin:     createBody.AmountAMin,
				AmountBMin:     createBody.AmountBMin,
				Deadline:       createBody.Deadline,
			},
		}, nil
	case status.Pair_RemoveLiquidity:
		bytes, err := json.Marshal(body.Function)
		if err != nil {
			return nil, err
		}
		remove := &RpcExchangeRemoveLiquidity{}
		err = json.Unmarshal(bytes, remove)
		if err != nil {
			return nil, err
		}

		return &ContractBody{
			Contract:     arry.StringToAddress(body.Contract),
			Type:         body.Type,
			FunctionType: body.FunctionType,
			Function: &exchange_func.ExchangeRemoveLiquidity{
				Exchange:   arry.StringToAddress(remove.Exchange),
				TokenA:     arry.StringToAddress(remove.TokenA),
				TokenB:     arry.StringToAddress(remove.TokenB),
				To:         arry.StringToAddress(remove.To),
				Liquidity:  remove.Liquidity,
				AmountAMin: remove.AmountAMin,
				AmountBMin: remove.AmountBMin,
				Deadline:   remove.Deadline,
			},
		}, nil
	}
	return nil, errors.New("wrong transaction body")
}

func addrListToHashAddr(addrList []string) []arry.Address {
	hashList := make([]arry.Address, len(addrList))
	for i, addr := range addrList {
		hashList[i] = arry.StringToAddress(addr)
	}
	return hashList
}

func hashAddrToAddr(hashList []arry.Address) []string {
	addrList := make([]string, len(hashList))
	for i, hash := range hashList {
		addrList[i] = hash.String()
	}
	return addrList
}
