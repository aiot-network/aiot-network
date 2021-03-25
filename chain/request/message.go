package request

const (
	Success Code = iota
	Failed
)

const timeOut = 30

type Code int

type Method string

type Message []byte

// Peer communication request body
type Request struct {
	Method Method  `json:"method"`
	Body   Message `json:"body"`
}

func NewRequest(method Method, body Message) *Request {
	return &Request{Method: method, Body: body}
}

// Peer node communication response message
type Response struct {
	Code    Code    `json:"code"`
	Message string  `json:"message"`
	Body    Message `json:"body"`
}

func NewResponse(code Code, message string, body []byte) *Response {
	return &Response{code, message, body}
}
