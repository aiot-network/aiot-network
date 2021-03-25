package types

type IBlock interface {
	IHeader
	BlockHeader() IHeader
	BlockBody() IBody
	ToRlpBlock() IRlpBlock
	CheckMsgRoot() bool
}

type IBlocks interface {
	Blocks() []IBlock
}
