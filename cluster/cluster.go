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

const MessageHeaderSize = 20

type Message struct {
    Version  byte       //   1 byte
    OP       byte       // + 1 bytes    = 2
    Args     []byte     // + 8 bytes    = 10
    Reserved []byte     // + 8 bytes    = 18
    Length   uint16     // + 2 bytes    = 20
    Load     []byte
}

func Unmarshall(packet []byte) (m *Message, err error) {
    if len(packet) < MessageHeaderSize {
        err = errors.New("Packet is too small")
        return
    }

    return &Message{
        Version:    packet[0],
        OP:         packet[1],
        Args:       packet[2:10],
        Reserved:   packet[10:18],
        Length:     uint16(packet[18]) << 8 | uint16(packet[19]),
        Load:       packet[20:],
    }, nil
}

func Marshall(m *Message) []byte {
    l := MessageHeaderSize + len(m.Load)
    buf := make([]byte, l, l)
    buf[0] = m.Version
    buf[1] = m.OP
    for i := range(m.Args) {
        buf[2 + i] = m.Args[i]
    }
    for i := 0; i < 8; i++ {
        buf[10 + i] = 0
    }
    buf[18] = byte(m.Length >> 8)
    buf[19] = byte(m.Length)

    for i, p := range m.Load {
        buf[20 + i] = p
    }    
    return buf
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
    m, err := Unmarshall(packet)
    if err != nil {
        return err
    }

    log.Printf("%x, %s\n", m.OP, string(m.Load))
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