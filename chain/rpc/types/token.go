package types

import (
	"github.com/aiot-network/aiotchain/chain/types"
	"github.com/aiot-network/aiotchain/tools/amount"
)

type RpcToken struct {
	Address        string    `json:"address"`
	Sender         string    `json:"sender"`
	Name           string    `json:"name"`
	Shorthand      string    `json:"shorthand"`
	IncreaseIssues bool      `json:"increaseissues"`
	PledgeRate     int       `json:"pledgerate"`
	PledgeAmount   float64   `json:"pledgeamount"`
	Records        []*Record `json:"records"`
}

type Record struct {
	Height   uint64  `json:"height"`
	Type     string  `json:"type"`
	Receiver string  `json:"receiver"`
	MsgHash  string  `json:"msghash"`
	Time     uint64  `json:"time"`
	Amount   float64 `json:"amount"`
}

func TokenToRpcToken(token *types.TokenRecord) *RpcToken {
	rpcToken := &RpcToken{
		Address:        token.Address.String(),
		Sender:         token.Sender.String(),
		Name:           token.Name,
		Shorthand:      token.Shorthand,
		IncreaseIssues: token.IncreaseIssues,
		PledgeRate:     int(token.PledgeRate),
		PledgeAmount:   amount.Amount(token.PledgeAmount).ToCoin(),
		Records:        make([]*Record, token.Records.Len()),
	}
	for i, record := range *token.Records {
		rpcToken.Records[i] = &Record{
			Height:   record.Height,
			Type:     record.Type,
			MsgHash:  record.MsgHash.String(),
			Receiver: record.Receiver.String(),
			Time:     record.Time,
			Amount:   amount.Amount(record.Amount).ToCoin(),
		}
	}
	return rpcToken
}
