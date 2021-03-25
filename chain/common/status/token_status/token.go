package token_status

import (
	"errors"
	"github.com/aiot-network/aiot-network/chain/db/status/token_db"
	chaintypes "github.com/aiot-network/aiot-network/chain/types"
	"github.com/aiot-network/aiot-network/common/config"
	"github.com/aiot-network/aiot-network/tools/arry"
	"github.com/aiot-network/aiot-network/types"
	"sync"
)

const tokenDB = "token_db"

type TokenStatus struct {
	db    ITokenDB
	mutex sync.RWMutex
}

func NewTokenStatus() (*TokenStatus, error) {
	db, err := token_db.Open(config.Param.Data + "/" + tokenDB)
	if err != nil {
		return nil, err
	}
	return &TokenStatus{db: db}, nil
}

func (t *TokenStatus) SetTrieRoot(hash arry.Hash) error {
	return t.db.SetRoot(hash)
}

func (t *TokenStatus) TrieRoot() arry.Hash {
	return t.db.Root()
}

func (t *TokenStatus) Commit() (arry.Hash, error) {
	return t.db.Commit()
}

func (t *TokenStatus) CheckMessage(msg types.IMessage) error {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	if chaintypes.MessageType(msg.Type()) != chaintypes.Token {
		return nil
	}
	body, ok := msg.MsgBody().(*chaintypes.TokenBody)
	if !ok {
		return errors.New("incorrect message type and message body")
	}
	token := t.db.Token(body.TokenAddress)
	if token != nil {
		return token.Check(msg)
	}
	return nil
}

// Update contract status
func (t *TokenStatus) UpdateToken(msg types.IMessage, height uint64) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	msgBody, ok := msg.MsgBody().(*chaintypes.TokenBody)
	if !ok {
		return errors.New("wrong message type")
	}
	record := &chaintypes.Record{
		Height:   height,
		MsgHash:  msg.Hash(),
		Receiver: msgBody.Receiver,
		Time:     msg.Time(),
		Amount:   msgBody.MsgAmount(),
	}
	tokenAddr := msgBody.TokenAddress
	token := t.db.Token(tokenAddr)
	if token != nil && token.IncreaseIssues {
		token.IncreaseRecord(record)
		token.Name = msgBody.Name
	} else {
		token = &chaintypes.TokenRecord{
			Address:        tokenAddr,
			Sender:         msg.From(),
			Name:           msgBody.Name,
			Shorthand:      msgBody.Shorthand,
			IncreaseIssues: msgBody.IncreaseIssues,
			Records: &chaintypes.RecordList{
				record,
			},
		}
	}
	t.db.SetToken(token)
	return nil
}

func (t *TokenStatus) Token(address arry.Address) (types.IToken, error) {
	token := t.db.Token(address)
	if token == nil {
		return nil, errors.New("not found")
	}
	return token, nil
}
