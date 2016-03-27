package cluster

import (
    "errors"
    "hash/fnv"
    "log"
    "fmt"
    "net"
    "github.com/noroutine/ffhash"
)

type Cluster struct {
    proxy *Node
    Server *Server
    Name string
    OrderId uint64
    Peers []string      // in contrast to proxy.Peers, this is stable ordered ring
}

func NewVia(node *Node) (c *Cluster, err error) {
    if ! node.IsOperational() {
        return nil, errors.New("Node is not ready")
    }

    return &Cluster{
        proxy: node,
        Name: *node.Group,
    }, nil
}

func (c *Cluster) Connect() {
    // start listening on the DHT
    serviceEntry := c.proxy.GetServiceEntry()
    c.Server = NewServer(serviceEntry.AddrIPv4, serviceEntry.AddrIPv6, c.proxy.Port, c)
    c.Server.Start()

    // determine order id
}

func (c *Cluster) Disconnect() {
    if c.proxy != nil {
        c.proxy = nil
    }
}

func (c *Cluster) Put(key string, data []byte) {
//    slot := keySlot(key, uint64(len(c.Peers)))
    // send data to node
}

func (c *Cluster) Get(key string) []byte {
//    slot := keySlot(key, uint64(len(c.Peers)))
    // get data from node
    return make([]byte, 0, 0)
}

func (c *Cluster) Receive(from *net.UDPAddr, m *Message) error {
    log.Println("Received", string(m.Load), "from", from)
    NewPingActivity(nil, c).Client()
    return nil
}

func (c *Cluster) Ping(peer string) error {
    p, ok := c.proxy.Peers[peer]
    if !ok {
        return errors.New(fmt.Sprintf("Peer not available: %s", peer))
    }

    targetAddr := &net.UDPAddr{
        IP: p.GetAddrIPv4(),
        Port: p.Port,
    }

    clientPingAutomat := NewPingActivity(targetAddr, c).Client()
    clientPingAutomat.Send(0)
    clientPingAutomat.Send(0)
    clientPingAutomat.Send(3)
    <- clientPingAutomat.Result
    return nil
}

func (c *Cluster) Send(to *net.UDPAddr, m *Message) error {
    log.Println("Sending", string(m.Load), "to", to)

    udpCl, err := NewClient(to)
    if err != nil {
        return err
    }

    defer udpCl.Close()

    return udpCl.Send(m)
}

func keySlot(key string, slots uint64) uint64 {
    fnv1a := fnv.New64a()
    fnv1a.Write([]byte(key))
    keyHash := fnv1a.Sum64()
    return ffhash.Sum64(keyHash, slots)
}