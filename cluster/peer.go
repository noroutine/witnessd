// Peer is what another node is to current node
package cluster

import (
    "net"
    "github.com/noroutine/bonjour"
    "github.com/reusee/mmh3"
    "math/big"
)

type Peer struct {
    Domain *string
    Name *string
    HostName *string
    Port int
    Group *string

    entry *bonjour.ServiceEntry
}

type Peers []Peer

func (p *Peer) GetAddrIPv4() net.IP {
    return p.entry.AddrIPv4
}

func (p *Peer) GetAddrIPv6() net.IP {
    return p.entry.AddrIPv6
}

func (p *Peer) Hash() []byte {
    return mmh3.Sum128([]byte(*p.Name))
}

func (ps Peers) Len() int {
    return len(ps)
}

func (ps Peers) Swap(i, j int) {
    ps[i], ps[j] = ps[j], ps[i]
}

func (ps Peers) Less(i, j int) bool {
    iHash := new(big.Int).SetBytes(ps[i].Hash())
    jHash := new(big.Int).SetBytes(ps[j].Hash())
    return iHash.Cmp(jHash) < 0
}