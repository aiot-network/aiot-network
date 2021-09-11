package types

import (
	"github.com/aiot-network/aiotchain/chain/types/functionbody/exchange_func"
	"github.com/aiot-network/aiotchain/chain/types/status"
	"github.com/aiot-network/aiotchain/tools/arry"
	"github.com/aiot-network/aiotchain/tools/rlp"
	"github.com/aiot-network/aiotchain/types"
)

type RlpMessage struct {
	MsgHeader *MsgHeader
	MsgBody   []byte
}

type RlpContract struct {
	MsgHeader *MsgHeader
	MsgBody   RlpContractBody
}

type RlpContractBody struct {
	Contract     arry.Address
	Type         status.ContractType
	FunctionType status.FunctionType
	Function     []byte
	State        ContractStatus
	Message      []byte
}

func (r *RlpMessage) ToMessage() types.IMessage {
	msg := &Message{}
	msg.Header = r.MsgHeader
	switch r.MsgHeader.Type {
	case Transaction:
		var body *TransactionBody
		rlp.DecodeBytes(r.MsgBody, &body)
		msg.Body = body
	case Token:
		var body *TokenBody
		rlp.DecodeBytes(r.MsgBody, &body)
		msg.Body = body
	case TokenV2:
		var body *TokenV2Body
		rlp.DecodeBytes(r.MsgBody, &body)
		msg.Body = body
	case Redemption:
		var body *RedemptionBody
		rlp.DecodeBytes(r.MsgBody, &body)
		msg.Body = body
	case Contract:
		var ct = &ContractBody{}
		var rlpCt *RlpContractBody
		rlp.DecodeBytes(r.MsgBody, &rlpCt)
		switch rlpCt.FunctionType {
		case status.Exchange_Init:
			var init *exchange_func.ExchangeInitBody
			rlp.DecodeBytes(rlpCt.Function, &init)
			ct.Function = init
		case status.Exchange_SetAdmin:
			var set *exchange_func.ExchangeAdmin
			rlp.DecodeBytes(rlpCt.Function, &set)
			ct.Function = set
		case status.Exchange_SetFeeTo:
			var set *exchange_func.ExchangeFeeTo
			rlp.DecodeBytes(rlpCt.Function, &set)
			ct.Function = set
		case status.Exchange_ExactIn:
			var in *exchange_func.ExactIn
			rlp.DecodeBytes(rlpCt.Function, &in)
			ct.Function = in
		case status.Exchange_ExactOut:
			var out *exchange_func.ExactOut
			rlp.DecodeBytes(rlpCt.Function, &out)
			ct.Function = out
		case status.Pair_AddLiquidity:
			var create *exchange_func.ExchangeAddLiquidity
			rlp.DecodeBytes(rlpCt.Function, &create)
			ct.Function = create
		case status.Pair_RemoveLiquidity:
			var create *exchange_func.ExchangeRemoveLiquidity
			rlp.DecodeBytes(rlpCt.Function, &create)
			ct.Function = create
		}
		rlp.DecodeBytes(r.MsgBody, &ct)
		return &Message{
			Header: r.MsgHeader,
			Body:   ct,
		}
	case Candidate:
		var body *CandidateBody
		rlp.DecodeBytes(r.MsgBody, &body)
		msg.Body = body
	case Cancel:
		var body *CancelBody
		rlp.DecodeBytes(r.MsgBody, &body)
		msg.Body = body
	case Vote:
		var body *VoteBody
		rlp.DecodeBytes(r.MsgBody, &body)
		msg.Body = body
	case Work:
		var body *WorkBody
		rlp.DecodeBytes(r.MsgBody, &body)
		msg.Body = body
	}
	return msg
}

func (r *RlpMessage) Bytes() []byte {
	bytes, _ := rlp.EncodeToBytes(r)
	return bytes
}

func DecodeMessage(bytes []byte) (*RlpMessage, error) {
	var msg *RlpMessage
	err := rlp.DecodeBytes(bytes, &msg)
	if err != nil {
		return nil, err
	}
	return msg, nil
}
