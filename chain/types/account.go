package types

import (
	"errors"
	"fmt"
	"github.com/aiot-network/aiot-network/common/config"
	"github.com/aiot-network/aiot-network/tools/arry"
	"github.com/aiot-network/aiot-network/tools/rlp"
	"github.com/aiot-network/aiot-network/types"
)

type Account struct {
	Address     arry.Address `json:"address"`
	Nonce       uint64       `json:"nonce"`
	Tokens      Tokens       `json:"tokens"`
	CurrentWork uint64       `json:"currentwork"`
	Confirmed   uint64       `json:"confirmed"`
	JournalIn   *journalIn   `json:"-"`
	JournalOut  *journalOut  `json:"-"`
	Works       *Works       `json:"works"`
}

func NewAccount() *Account {
	return &Account{
		Tokens:     make(Tokens, 0),
		JournalOut: newJournalOut(),
		JournalIn:  newJournalIn(),
		Works:      &Works{},
	}
}

func DecodeAccount(bytes []byte) (*Account, error) {
	var account *Account
	err := rlp.DecodeBytes(bytes, &account)
	return account, err
}

func (a *Account) NeedUpdate() bool {
	for _, token := range a.Tokens {
		if token.LockedIn != 0 || token.LockedOut != 0 {
			return true
		}
	}
	return false
}

// Update through the account transfer log information
func (a *Account) UpdateLocked(confirmed uint64) error {
	for _, out := range a.JournalOut.GetJournalOuts(confirmed) {
		coinAccount, ok := a.Tokens.Get(out.TokenAddress)
		if !ok {
			return errors.New("wrong journal")
		}
		if coinAccount.LockedOut >= out.Amount {
			coinAccount.LockedOut -= out.Amount
			a.Tokens.Set(coinAccount)

			tokenAccount, ok := a.Tokens.Get(config.Param.MainToken.String())
			if !ok {
				return errors.New("wrong journal")
			}
			if tokenAccount.LockedOut >= out.Fees {
				tokenAccount.LockedOut -= out.Fees
				a.Tokens.Set(tokenAccount)
			} else {
				return errors.New("locked out amount not enough when update account journal")
			}
			a.JournalOut.Remove(out.Height)

		} else {
			return errors.New("locked out amount not enough when update account journal")
		}
	}

	// Update through account transfer log information
	for _, in := range a.JournalIn.GetJournalIns(confirmed) {
		coinAccount, ok := a.Tokens.Get(in.TokenAddress)
		if !ok {
			coinAccount = &TokenAccount{
				Address:   in.TokenAddress,
				Balance:   0,
				LockedIn:  0,
				LockedOut: 0,
			}
		}
		if coinAccount.LockedIn >= in.Amount {
			coinAccount.Balance += in.Amount
			coinAccount.LockedIn -= in.Amount
			a.Tokens.Set(coinAccount)
			a.JournalIn.Remove(in.Height, in.TokenAddress)
		} else {
			return errors.New("locked in amount not enough when update account Journal")
		}
	}
	a.Confirmed = confirmed
	return nil
}

func (a *Account) FromMessage(msg types.IMessage, height uint64) error {
	if a.Nonce+1 != msg.Nonce() {
		return fmt.Errorf("wrong nonce value")
	}

	if MessageType(msg.Type()) == Token {
		return a.addToken(msg, height)
	}
	body := msg.MsgBody()
	tokenAddr := body.MsgToken()
	if tokenAddr == config.Param.MainToken {
		return a.changeMain(msg, height)
	} else {
		return a.changeToken(msg, height)
	}
}

// Change of contract information
func (a *Account) addToken(msg types.IMessage, height uint64) error {
	fees := msg.Fee()
	mainAccount, ok := a.Tokens.Get(config.Param.MainToken.String())
	if !ok {
		return errors.New("account is not exist")
	}
	if mainAccount.Balance < fees {
		return fmt.Errorf("balance %d is not enough to pay the fee %d", mainAccount.Balance, fees)
	}
	mainAccount.Balance -= fees
	mainAccount.LockedOut += fees

	consume := config.Param.Consume
	if mainAccount.Balance < consume {
		return fmt.Errorf("insufficient balance, %d %s is required to publish tokens", consume, config.Param.MainToken.String())
	}
	mainAccount.Balance -= consume
	mainAccount.LockedOut += consume

	a.Tokens.Set(mainAccount)
	a.Nonce = msg.Nonce()
	a.JournalOut.Add(msg, height)
	return nil
}

// Change the primary account status of one party to the transaction transfer
func (a *Account) changeMain(msg types.IMessage, height uint64) error {
	amount := msg.Fee() + msg.MsgBody().MsgAmount()
	if !a.Exist() {
		a.Address = msg.From()
	}
	mainAccount, ok := a.Tokens.Get(config.Param.MainToken.String())
	if !ok {
		mainAccount.Address = config.Param.MainToken.String()
	}

	if mainAccount.Balance < amount {
		return fmt.Errorf("insufficient balance")
	}
	if a.Nonce+1 != msg.Nonce() {
		return fmt.Errorf("wrong nonce value")
	}

	mainAccount.Balance -= amount
	mainAccount.LockedOut += amount
	a.Tokens.Set(mainAccount)
	a.Nonce = msg.Nonce()
	a.JournalOut.Add(msg, height)
	return nil
}

// Change the status of the secondary account of the transaction transfer party.
// The transaction of the secondary account needs to consume the fee of the
// primary account.
func (a *Account) changeToken(msg types.IMessage, height uint64) error {
	fees := msg.Fee()
	msgBody := msg.MsgBody()

	amount := msgBody.MsgAmount()
	mainAccount, ok := a.Tokens.Get(config.Param.MainToken.String())
	if !ok {
		return errors.New("account is not exist")
	}
	if mainAccount.Balance < fees {
		return fmt.Errorf("insufficient balance")
	}
	tokenAddr := msgBody.MsgToken()
	coinAccount, ok := a.Tokens.Get(tokenAddr.String())
	if !ok {
		return errors.New("account is not exist")
	}
	if coinAccount.Balance < amount {
		return fmt.Errorf("insufficient balance")
	}

	mainAccount.Balance -= fees
	mainAccount.LockedOut += fees
	coinAccount.Balance -= amount
	coinAccount.LockedOut += amount
	a.Tokens.Set(mainAccount)
	a.Tokens.Set(coinAccount)
	a.Nonce = msg.Nonce()
	a.JournalOut.Add(msg, height)
	return nil
}

func (a *Account) ToMessage(msgType int, address, token arry.Address, amount, height uint64) error {
	if !a.Exist() {
		a.Address = address
	}
	if MessageType(msgType) == Token {
		return a.toTokenChange(token, amount, height)
	}
	tokenAccount, ok := a.Tokens.Get(token.String())
	if ok {
		tokenAccount.LockedIn += amount
	} else {
		tokenAccount = &TokenAccount{
			Address:   token.String(),
			Balance:   0,
			LockedIn:  amount,
			LockedOut: 0,
		}
	}
	a.Tokens.Set(tokenAccount)
	a.JournalIn.Add(amount, height, token.String())
	return nil
}

func (a *Account) WorkMessage(address arry.Address, workload, cycle, endTime uint64) {
	if !a.Exist() {
		a.Address = address
	}
	if a.Works.Cycle < cycle {
		a.Works.Cycle = cycle
		a.Works.WorkLoad = workload
		a.Works.EndTime = endTime
	} else if a.Works.Cycle == cycle {
		a.Works.WorkLoad += workload
		a.Works.EndTime = endTime
	}
}

func (a *Account) EaterMessage(height uint64) error {
	if !a.Exist() {
		a.Address = config.Param.EaterAddress
	}

	mainAccount, ok := a.Tokens.Get(config.Param.MainToken.String())
	if ok {
		mainAccount.LockedIn += config.Param.Consume
	} else {
		mainAccount = &TokenAccount{
			Address:   config.Param.MainToken.String(),
			Balance:   0,
			LockedIn:  config.Param.Consume,
			LockedOut: 0,
		}
	}
	a.Tokens.Set(mainAccount)
	a.JournalIn.Add(config.Param.Consume, height, config.Param.MainToken.String())
	return nil
}

// Change of contract information
func (a *Account) toTokenChange(token arry.Address, amount, height uint64) error {
	tokenAccount, ok := a.Tokens.Get(token.String())
	if ok {
		tokenAccount.LockedIn += amount
	} else {
		tokenAccount = &TokenAccount{
			Address:   token.String(),
			Balance:   0,
			LockedOut: 0,
			LockedIn:  amount,
		}
	}

	a.Tokens.Set(tokenAccount)
	a.JournalIn.Add(amount, height, token.String())
	return nil
}

func (a *Account) GetBalance(tokenAddr arry.Address) uint64 {
	token, ok := a.Tokens.Get(tokenAddr.String())
	if !ok {
		return 0
	}
	return token.Balance
}

func (a *Account) GetWorks() types.IWorks {
	return a.Works
}

func (a *Account) Check(msg types.IMessage, strict bool) error {
	if !a.Exist() {
		a.Address = msg.From()
	}

	if strict {
		if msg.Nonce() != a.Nonce+1 {
			return fmt.Errorf("nonce value must be %d", a.Nonce+1)
		}
	} else if msg.Nonce() <= a.Nonce {
		return fmt.Errorf("the nonce value of the message must be greater than %d", a.Nonce)
	}

	// The nonce value cannot be greater than the
	// maximum number of address transactions
	if msg.Nonce() > a.Nonce+config.Param.MaxAddressMsg {
		return fmt.Errorf("the nonce value of the message cannot be greater "+
			"than the nonce value of the account %d", config.Param.MaxAddressMsg)
	}

	// Verify the balance of the token
	switch MessageType(msg.Type()) {
	case Transaction:
		body, ok := msg.MsgBody().(*TransactionBody)
		if !ok {
			return errors.New("incorrect message type and message body")
		}
		if body.TokenAddress.IsEqual(config.Param.MainToken) {
			return a.checkMainBalance(msg)
		} else {
			return a.checkTokenBalance(msg, body)
		}
	case Token:
		return a.checkConsume(msg)
	default:
		if msg.MsgBody().MsgAmount() != 0 {
			return errors.New("wrong amount")
		}
		return a.checkFees(msg)
	}
}

func (a *Account) checkConsume(msg types.IMessage) error {
	main := config.Param.MainToken.String()
	consume := msg.Fee() + config.Param.Consume
	token, ok := a.Tokens.Get(main)
	if !ok {
		return fmt.Errorf("insufficient balance, %d %s is required to publish tokens", consume, main)
	} else if token.Balance < consume {
		return fmt.Errorf("insufficient balance, %d %s is required to publish tokens", consume, main)
	}
	return nil
}

// Verify the account balance of the primary transaction, the transaction
// value and transaction fee cannot be greater than the balance.
func (a *Account) checkMainBalance(msg types.IMessage) error {
	main := config.Param.MainToken.String()
	token, ok := a.Tokens.Get(main)
	if !ok {
		return fmt.Errorf("%s does not have enough balance", main)
	} else if token.Balance < msg.Fee()+msg.MsgBody().MsgAmount() {
		return fmt.Errorf("%s does not have enough balance", main)
	}
	return nil
}

// Verify the account balance of the secondary transaction, the transaction
// value cannot be greater than the balance.
func (a *Account) checkTokenBalance(msg types.IMessage, body *TransactionBody) error {
	if err := a.checkFees(msg); err != nil {
		return err
	}

	coinAccount, ok := a.Tokens.Get(body.TokenAddress.String())
	if !ok {
		return fmt.Errorf("%s does not have enough balance", body.TokenAddress.String())
	} else if coinAccount.Balance < body.MsgAmount() {
		return fmt.Errorf("%s does not have enough balance", body.TokenAddress.String())
	}
	return nil
}

// Verification fee
func (a *Account) checkFees(msg types.IMessage) error {
	main := config.Param.MainToken.String()
	token, ok := a.Tokens.Get(main)
	if !ok && msg.Fee() != 0 {
		return fmt.Errorf("%s does not have enough balance to pay the handling fee", main)
	} else if token.Balance < msg.Fee() {
		return fmt.Errorf("%s does not have enough balance to pay the handling fee", main)
	}
	return nil
}

func (a *Account) Exist() bool {
	return !arry.EmptyAddress(a.Address)
}

func (a *Account) Bytes() []byte {
	bytes, _ := rlp.EncodeToBytes(a)
	return bytes
}

func (a *Account) GetAddress() arry.Address {
	return a.Address
}

type TokenAccount struct {
	Address   string `json:"address"`
	Balance   uint64 `json:"balance"`
	LockedIn  uint64 `json:"locked"`
	LockedOut uint64 `json:"-"`
}

// List of secondary accounts
type Tokens []*TokenAccount

func (t *Tokens) Get(contract string) (*TokenAccount, bool) {
	for _, coin := range *t {
		if coin.Address == contract {
			return coin, true
		}
	}
	return &TokenAccount{}, false
}

func (t *Tokens) Set(newCoin *TokenAccount) {
	for i, coin := range *t {
		if coin.Address == newCoin.Address {
			(*t)[i] = newCoin
			return
		}
	}
	*t = append(*t, newCoin)
}

// Account transfer log
type journalOut struct {
	Outs *TxOutList
}

func newJournalOut() *journalOut {
	return &journalOut{Outs: &TxOutList{}}
}

func (j *journalOut) Add(msg types.IMessage, height uint64) {
	body := msg.MsgBody()
	TokenAddress := body.MsgToken()
	amount := body.MsgAmount()
	if MessageType(msg.Type()) == Token {
		TokenAddress = config.Param.MainToken
		amount = 0
	}
	j.Outs.Set(&txOut{
		TokenAddress: TokenAddress.String(),
		Amount:       amount,
		Fees:         msg.Fee(),
		Nonce:        msg.Nonce(),
		Time:         uint64(msg.Time()),
		Height:       height,
	})
}

func (j *journalOut) Get(height uint64) *txOut {
	in, ok := j.Outs.Get(height)
	if ok {
		return in
	}
	return nil
}

func (j *journalOut) Remove(height uint64) uint64 {
	tx, _ := j.Outs.Get(height)
	j.Outs.Remove(height)
	return tx.Amount
}

func (j *journalOut) IsExist(height uint64) bool {
	for _, txIn := range *j.Outs {
		if txIn.Height >= height {
			return true
		}
	}
	return false
}

func (j *journalOut) GetJournalOuts(confirmedHeight uint64) []*txOut {
	txIns := make([]*txOut, 0)
	for _, txIn := range *j.Outs {
		if txIn.Height <= confirmedHeight {
			txIns = append(txIns, txIn)
		}
	}
	return txIns
}

func (j *journalOut) Amount() map[string]uint64 {
	amounts := map[string]uint64{}
	for _, txIn := range *j.Outs {
		_, ok := amounts[txIn.TokenAddress]
		if ok {
			amounts[txIn.TokenAddress] += txIn.Amount
		} else {
			amounts[txIn.TokenAddress] = txIn.Amount
		}
	}
	return amounts
}

func (j *journalOut) IsEmpty() bool {
	if j.Outs == nil || len(*j.Outs) == 0 {
		return true
	}
	return false
}

type txOut struct {
	TokenAddress string
	Amount       uint64
	Fees         uint64
	Nonce        uint64
	Time         uint64
	Height       uint64
}

type TxOutList []*txOut

func (t *TxOutList) Get(height uint64) (*txOut, bool) {
	for _, txIn := range *t {
		if txIn.Height == height {
			return txIn, true
		}
	}
	return &txOut{}, false
}

func (t *TxOutList) Set(txIn *txOut) {
	for i, in := range *t {
		if in.Height == txIn.Height {
			(*t)[i] = txIn
			return
		}
	}
	*t = append(*t, txIn)
}

func (t *TxOutList) Remove(height uint64) {
	for i, in := range *t {
		if in.Height == height {
			*t = append((*t)[0:i], (*t)[i+1:]...)
			return
		}
	}
}

// Account transfer log
type journalIn struct {
	Ins *InList
}

func newJournalIn() *journalIn {
	return &journalIn{Ins: &InList{}}
}

func (j *journalIn) Add(amount uint64, height uint64, tokenAddr string) {
	in, ok := j.Ins.Get(height, tokenAddr)
	if ok {
		in.Amount += amount
	} else {
		in = &InAmount{}
		in.Amount = amount
		in.Height = height
		in.TokenAddress = tokenAddr
	}
	j.Ins.Set(in)
}

func (j *journalIn) Get(height uint64, contract string) *InAmount {
	txOut, ok := j.Ins.Get(height, contract)
	if ok {
		return txOut
	}
	return &InAmount{"", 0, 0}
}

func (j *journalIn) IsExist(height uint64) bool {
	for _, out := range *j.Ins {
		if out.Height >= height {
			return true
		}
	}
	return false
}

func (j *journalIn) Remove(height uint64, contract string) *InAmount {
	return j.Ins.Remove(height, contract)
}

func (j *journalIn) GetJournalIns(confirmedHeight uint64) map[string]*InAmount {
	txOuts := make(map[string]*InAmount)
	for _, out := range *j.Ins {
		if out.Height <= confirmedHeight {
			key := fmt.Sprintf("%s_%d", out.TokenAddress, out.Height)
			txOuts[key] = out
		}
	}
	return txOuts
}

func (j *journalIn) IsEmpty() bool {
	if j.Ins == nil || len(*j.Ins) == 0 {
		return true
	}
	return false
}

type InAmount struct {
	TokenAddress string
	Amount       uint64
	Height       uint64
}

type InList []*InAmount

func (o *InList) Get(height uint64, tokenAddr string) (*InAmount, bool) {
	for _, out := range *o {
		if out.Height == height && out.TokenAddress == tokenAddr {
			return out, true
		}
	}
	return &InAmount{}, false
}

func (o *InList) Set(outAmount *InAmount) {
	for i, out := range *o {
		if out.Height == outAmount.Height && out.TokenAddress == outAmount.TokenAddress {
			(*o)[i] = outAmount
			return
		}
	}
	*o = append(*o, outAmount)
}

func (o *InList) Remove(height uint64, TokenAddress string) *InAmount {
	for i, out := range *o {
		if out.Height == height && out.TokenAddress == TokenAddress {
			*o = append((*o)[0:i], (*o)[i+1:]...)
			return out
		}
	}
	return nil
}

type Works struct {
	Cycle    uint64 `json:"cycle"`
	WorkLoad uint64 `json:"workload"`
	EndTime  uint64 `json:"end"`
}

func (w *Works) GetCycle() uint64 {
	return w.Cycle
}

func (w *Works) GetEndTime() uint64 {
	return w.EndTime
}

func (w *Works) GetWorkLoad() uint64 {
	return w.WorkLoad
}

type Reword struct {
	Cycle    uint64 `json:"cycle"`
	Amount   uint64 `json:"amount"`
	EndTime  uint64 `json:"end"`
	Address  string `json:"address"`
	WorkLoad uint64 `json:"workload"`
	Blocks   uint64 `json:"blocks"`
}

func (r *Reword) GetCycle() uint64 {
	return r.Cycle
}

func (r *Reword) GetEndTime() uint64 {
	return r.EndTime
}

func (r *Reword) GetReword() uint64 {
	return r.Amount
}

func (r *Reword) GetAddress() string {
	return r.Address
}

func (r *Reword) GetWorkLoad() uint64 {
	return r.WorkLoad
}

func (r *Reword) GetBlocks() uint64 {
	return r.Blocks
}
