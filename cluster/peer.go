package cluster

import (
    "net"
    "github.com/noroutine/bonjour"
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