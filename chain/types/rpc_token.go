package types

type RpcTokenBody struct {
	Address        string `json:"address"`
	Receiver       string `json:"receiver"`
	Name           string `json:"name"`
	Shorthand      string `json:"shorthand"`
	Amount         uint64 `json:"amount"`
	IncreaseIssues bool   `json:"allowedincrease"`
}
