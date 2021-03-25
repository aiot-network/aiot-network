package types

type IRlpBlock interface {
	Bytes() []byte
	ToBlock() IBlock
}
