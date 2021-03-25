package types

type RpcCandidateBody struct {
	PeerId string `json:"peerid"`
}

func (r *RpcCandidateBody) PeerIdBytes() []byte {
	return []byte(r.PeerId)
}
