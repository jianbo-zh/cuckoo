package groupnetworkproto

import (
	"fmt"
	"testing"

	"github.com/libp2p/go-libp2p/core/peer"
)

func TestNewConnAlgo(t *testing.T) {

	hostID, _ := peer.Decode("12D3KooWQ28EWtkcDssrTe1CAmXXqLbm8gDaUNXay7gXKZJvH5Rt")
	peerID, _ := peer.Decode("12D3KooWLAiy4n8ctb2agLq3FG4r3H9cgsvdKQKzmPh6F1v79Zkp")

	got := NewConnAlgo(hostID, []peer.ID{peerID})

	pp := got.GetClosestPeers()
	fmt.Println(peer.ID(pp[0]).String())
	fmt.Println(111)
}
