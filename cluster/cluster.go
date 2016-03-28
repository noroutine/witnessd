package cluster

import (
    "errors"
    "hash/fnv"
    "fmt"
    "net"
    "container/list"
    "github.com/noroutine/ffhash"
)

type Cluster struct {
    proxy *Node
    Server *Server
    Name string
    OrderId uint64
    Peers []string      // in contrast to proxy.Peers, this is stable ordered ring
    handlers *list.List
}

func NewVia(node *Node) (c *Cluster, err error) {
    if ! node.IsOperational() {
        return nil, errors.New("Node is not ready")
    }

    c = &Cluster{
        proxy: node,
        Name: *node.Group,
        handlers: list.New(),
    }
    c.handlers.PushBack(NewPongActivity(c))
    return c, nil
}

func (c *Cluster) Connect() {
    // start listening on the DHT
    c.Name = *c.proxy.Group
    serviceEntry := c.proxy.GetServiceEntry()
    c.Server = NewServer(serviceEntry.AddrIPv4, serviceEntry.AddrIPv6, c.proxy.Port, c)
    c.Server.Start()

    // determine order id
}

func (c *Cluster) Disconnect() {
    if c.Server != nil {
        c.Server.Shutdown()
        c.Server = nil
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

func (c *Cluster) Route(r *Request) (h Handler, err error) {
    for e := c.handlers.Front(); e != nil; e = e.Next() {
        h, err := e.Value.(Router).Route(r)
        if h != nil && err == nil {
            return h, nil
        }
    }

    return nil, errors.New("Not supported")
}

func (c *Cluster) Ping(peer string) int {
    activity := NewPingActivity(c)

    e := c.handlers.PushBack(activity)
    defer c.handlers.Remove(e)

    activity.Run(peer)
    return <- activity.Result
}

func (c *Cluster) Send(to *net.UDPAddr, m *Message) error {
    udpCl, err := NewClient(to)
    if err != nil {
        return err
    }

    defer udpCl.Close()

    return udpCl.Send(m)
}

func (c *Cluster) GetPeerAddr(peer string) (*net.UDPAddr, error) {
    p, ok := c.proxy.Peers[peer]
    if !ok {
        return nil, errors.New(fmt.Sprintf("Peer not available: %s", peer))
    }

    return &net.UDPAddr{
        IP: p.GetAddrIPv4(),
        Port: p.Port,
    }, nil
}

func keySlot(key string, slots uint64) uint64 {
    fnv1a := fnv.New64a()
    fnv1a.Write([]byte(key))
    keyHash := fnv1a.Sum64()
    return ffhash.Sum64(keyHash, slots)
}