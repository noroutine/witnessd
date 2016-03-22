package cluster

import (
    "net"
    "log"
    "errors"
    "hash/fnv"
    "github.com/noroutine/ffhash"
    "github.com/noroutine/dominion/group"
)

type Cluster struct {
    proxy *group.Node
    s *Server
    Name string
    OrderId uint64
    Peers []string      // in contrast to proxy.Peers, this is stable ordered ring
}

type Server struct {
    ipv4conn *net.UDPConn
    ipv6conn *net.UDPConn
    shouldShutdown bool
}

const dhtPort int = 9999

func ConnectVia(node *group.Node) (c *Cluster, err error) {

    if ! node.IsOperational() {
        err = errors.New("Node is not ready")
        return
    }

    c = &Cluster{
        proxy: node,
        Name: *node.Group,
    }

    // start listening on the DHT
    c.clusterConnectionLoop(dhtPort)

    // determine order id

    c.proxy = node
    return
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

func (c *Cluster) clusterConnectionLoop(port int) {
    serviceEntry := c.proxy.GetServiceEntry()

    l4, err := net.ListenUDP("udp4", &net.UDPAddr{ IP: serviceEntry.AddrIPv4, Port: port })
    if err != nil {
        log.Fatal(err)
    }

    l6, err := net.ListenUDP("udp6", &net.UDPAddr{ IP: serviceEntry.AddrIPv6, Port: port })
    if err != nil {
        log.Fatal(err)
    }

    c.s = &Server{
        ipv4conn: l4,
        ipv6conn: l6,
        shouldShutdown: false,
    }

    if c.s.ipv4conn != nil {
        go c.s.serve(c.s.ipv4conn)
    }
    
    if c.s.ipv6conn != nil {
        go c.s.serve(c.s.ipv6conn)    
    }
}

func (s *Server) serve(c *net.UDPConn) {
    if c == nil {
        return
    }
    
    buf := make([]byte, 65536)
    
    for !s.shouldShutdown {
        n, from, err := c.ReadFrom(buf)
        if err != nil {
            log.Fatalf("Error reading from UDP socket %v\n", c)
            continue
        }
        if err := s.handlePacket(buf[:n], from); err != nil {
            log.Printf("[ERR] cluster: Failed to handle query: %v", err)
        }
    }
}

func (s *Server) handlePacket(packet []byte, from net.Addr) error {
    log.Printf(string(packet))
    return nil
}

func (s *Server) Shutdown() {
    s.shouldShutdown = true
}

func keySlot(key string, slots uint64) uint64 {
    fnv1a := fnv.New64a()
    fnv1a.Write([]byte(key))
    keyHash := fnv1a.Sum64()
    return ffhash.Sum64(keyHash, slots)
}