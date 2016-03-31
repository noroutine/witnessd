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

func (p *Peer) GetAddrIPv4() net.IP {
    return p.entry.AddrIPv4
}

func (p *Peer) GetAddrIPv6() net.IP {
    return p.entry.AddrIPv6
}

func (p *Peer) Hash() []byte {
    return mmh3.Sum128([]byte(*p.Name))
}

type hashSorter struct {
    peers []Peer
}

func PeersByHash(ps []Peer) *hashSorter {
    return &hashSorter{
        peers: ps,
    }
}

func (hs *hashSorter) Len() int {
    return len(hs.peers)
}

func (hs *hashSorter) Swap(i, j int) {
    ps := hs.peers
    ps[i], ps[j] = ps[j], ps[i]
}

func (hs *hashSorter) Less(i, j int) bool {
    ps := hs.peers
    iHash := new(big.Int).SetBytes(ps[i].Hash())
    jHash := new(big.Int).SetBytes(ps[j].Hash())
    return iHash.Cmp(jHash) < 0
}