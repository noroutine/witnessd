// Peer is what another node is to current node
package cluster

import (
    "net"
    "github.com/reusee/mmh3"
    "math/big"
    "sort"
    "strings"
    "strconv"
)

const hash_byte_len = 16        // 128 / 8

type Peer struct {
    Domain *string
    Name *string
    HostName *string
    Port int
    Group *string
    AddrIPv4 net.IP
    AddrIPv6 net.IP
    Text []string
}

type Partition struct {
    Name string
}

func (p *Peer) Clone() *Peer {
    return &Peer{
        Domain: p.Domain,
        Name: p.Name,
        HostName: p.HostName,
        Port: p.Port,
        Group: p.Group,
        AddrIPv4: p.AddrIPv4,
        AddrIPv6: p.AddrIPv6,
        Text: p.Text,
    }
}

func (p *Peer) GetPartitions() int {
    partitionsValue := p.getText(partitionsKey)
    if partitionsValue == nil {
        return 1
    } else {
        partitions, err := strconv.Atoi(*partitionsValue)
        if err != nil {
            return 1
        } else {
            return partitions
        }
    }
}

func (p *Peer) getText(key string) *string {
    ketEq := key + "="
    for _, s := range p.Text {
        if strings.HasPrefix(s, ketEq) {
            value := strings.TrimPrefix(s, ketEq)
            return &value
        }
    }

    return nil
}

func (p *Peer) Hash() []byte {
    return mmh3.Sum128([]byte(*p.Name))
}

type peerSorter struct {
    peers []*Peer
    less func (*Peer, *Peer) bool
}

func PeerSorter(ps []*Peer) *peerSorter {
    return &peerSorter{
        peers: ps,
        less: nil,
    }
}

func (sorter *peerSorter) ByHash() *peerSorter {
    sorter.less = func (p1, p2 *Peer) bool {
        iHash := new(big.Int).SetBytes(p1.Hash())
        jHash := new(big.Int).SetBytes(p2.Hash())
        return iHash.Cmp(jHash) < 0
    }

    return sorter
}

func (sorter *peerSorter) Sort() []*Peer {
    sort.Sort(sorter)
    return sorter.peers
}

func (hs *peerSorter) Len() int {
    return len(hs.peers)
}

func (hs *peerSorter) Swap(i, j int) {
    ps := hs.peers
    ps[i], ps[j] = ps[j], ps[i]
}

func (sorter *peerSorter) Less(i, j int) bool {
    return sorter.less(sorter.peers[i], sorter.peers[j])
}