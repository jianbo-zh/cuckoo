package message

import (
	"fmt"
	"net"

	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
)

var publicCIDR6 = "2000::/3"
var public6 *net.IPNet

func init() {
	_, public6, _ = net.ParseCIDR(publicCIDR6)
}

func isPublicAddr(a ma.Multiaddr) bool {
	ip, err := manet.ToIP(a)
	if err != nil {
		return false
	}
	if ip.To4() != nil {
		return !inAddrRange(ip, manet.Private4) && !inAddrRange(ip, manet.Unroutable4)
	}

	return public6.Contains(ip)
}

func inAddrRange(ip net.IP, ipnets []*net.IPNet) bool {
	for _, ipnet := range ipnets {
		if ipnet.Contains(ip) {
			return true
		}
	}

	return false
}

func msgID(lamptime uint64, peerID peer.ID) string {
	return fmt.Sprintf("%019d_%s", lamptime, peerID.String())
}
