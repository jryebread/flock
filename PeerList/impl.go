package PeerList

type PeerListImpl struct {
	Peers []Peer
}

type Peer struct {
	name string
	ip string

}
func InitPeerList () *PeerListImpl {
	return &PeerListImpl{
		[]Peer{ },
	}
}
