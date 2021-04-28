package message

import (
	"github.com/aiot-network/aiotchain/chain/types"
	"github.com/aiot-network/aiotchain/tools/arry"
	"github.com/aiot-network/aiotchain/tools/crypto/ecc/secp256k1"
	"time"
)

func NewTransaction(from, token string, to []map[string]uint64, fee, nonce, t uint64) *types.Message {
	if t == 0 {
		t = uint64(time.Now().Unix())
	}
	recis := types.NewReceivers()
	for _, addrAmt := range to {
		for addr, amt := range addrAmt {
			recis.Add(arry.StringToAddress(addr), amt)
		}
	}
	tx := &types.Message{
		Header: &types.MsgHeader{
			Type:      types.Transaction,
			Hash:      arry.Hash{},
			From:      arry.StringToAddress(from),
			Nonce:     nonce,
			Fee:       fee,
			Time:      t,
			Signature: &types.Signature{},
		},

		Body: &types.TransactionBody{
			TokenAddress: arry.StringToAddress(token),
			Receivers:    recis,
		},
	}
	tx.SetHash()
	return tx
}

func NewCandidate(from string, peerStr string, fee, nonce, t uint64) *types.Message {
	if t == 0 {
		t = uint64(time.Now().Unix())
	}
	var peerID types.Peer
	copy(peerID[:], peerStr)
	can := &types.Message{
		Header: &types.MsgHeader{
			Type:      types.Candidate,
			Hash:      arry.Hash{},
			From:      arry.StringToAddress(from),
			Nonce:     nonce,
			Fee:       fee,
			Time:      t,
			Signature: &types.Signature{},
		},
		Body: &types.CandidateBody{peerID},
	}
	can.SetHash()
	return can
}

func NewCancel(from string, fee, nonce, t uint64) *types.Message {
	if t == 0 {
		t = uint64(time.Now().Unix())
	}
	cancel := &types.Message{
		Header: &types.MsgHeader{
			Type:      types.Cancel,
			Hash:      arry.Hash{},
			From:      arry.StringToAddress(from),
			Nonce:     nonce,
			Fee:       fee,
			Time:      t,
			Signature: &types.Signature{},
		},
		Body: &types.CancelBody{},
	}
	cancel.SetHash()
	return cancel
}

func NewVote(from, to string, fee, nonce, t uint64) *types.Message {
	if t == 0 {
		t = uint64(time.Now().Unix())
	}
	vote := &types.Message{
		Header: &types.MsgHeader{
			Type:      types.Vote,
			Hash:      arry.Hash{},
			From:      arry.StringToAddress(from),
			Nonce:     nonce,
			Fee:       fee,
			Time:      t,
			Signature: &types.Signature{},
		},
		Body: &types.VoteBody{arry.StringToAddress(to)},
	}
	vote.SetHash()
	return vote
}

func NewToken(from, to, tokenAddr string, amount, fee, nonce, t uint64, name, shorthand string, allowIncrease bool) *types.Message {
	if t == 0 {
		t = uint64(time.Now().Unix())
	}
	token := &types.Message{
		Header: &types.MsgHeader{
			Type:      types.Token,
			Hash:      arry.Hash{},
			From:      arry.StringToAddress(from),
			Nonce:     nonce,
			Fee:       fee,
			Time:      t,
			Signature: &types.Signature{},
		},
		Body: &types.TokenBody{
			TokenAddress:   arry.StringToAddress(tokenAddr),
			Receiver:       arry.StringToAddress(to),
			Name:           name,
			Shorthand:      shorthand,
			IncreaseIssues: allowIncrease,
			Amount:         amount,
		},
	}
	token.SetHash()
	return token
}

func NewWork(from string, nonce uint64, start uint64, end uint64, t uint64, works map[string]uint64) *types.Message {
	if t == 0 {
		t = uint64(time.Now().Unix())
	}
	list := []types.AddressWork{}
	for addr, work := range works {
		list = append(list, types.AddressWork{
			Address:  arry.StringToAddress(addr),
			Workload: work,
			EndTime:  end,
		})
	}
	work := &types.Message{
		Header: &types.MsgHeader{
			Type:      types.Work,
			Hash:      arry.Hash{},
			From:      arry.StringToAddress(from),
			Nonce:     nonce,
			Fee:       0,
			Time:      t,
			Signature: &types.Signature{},
		},
		Body: &types.WorkBody{
			StartTime: start,
			EndTime:   end,
			List:      list,
		},
	}
	work.SetHash()
	return work
}

func Sign(keyStr string, hash string) (*types.Signature, error) {
	key, err := secp256k1.ParseStringToPrivate(keyStr)
	if err != nil {
		return nil, err
	}
	h, err := arry.StringToHash(hash)
	if err != nil {
		return nil, err
	}
	return types.Sign(key, h)
}
