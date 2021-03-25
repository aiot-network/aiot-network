package types

type ISignature interface {
	PubicKey() []byte
	SignatureBytes() []byte
}
