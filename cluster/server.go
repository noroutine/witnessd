package cluster

import (
    "net"
    "log"
)

type MessageReceiver interface {
    Receive(*net.UDPAddr, *Message) error
}

type MessageSender interface {
    Send(*net.UDPAddr, *Message) error
}

type Server struct {
    ipv4conn *net.UDPConn
    ipv6conn *net.UDPConn
    shouldShutdown bool
    receiver MessageReceiver
}

func NewServer(ip4 net.IP, ip6 net.IP, port int, mr MessageReceiver) *Server {
    l4, err := net.ListenUDP("udp4", &net.UDPAddr{ IP: ip4, Port: port })
    if err != nil {
        log.Fatal(err)
    }

    l6, err := net.ListenUDP("udp6", &net.UDPAddr{ IP: ip6, Port: port })
    if err != nil {
        log.Fatal(err)
    }

    return &Server{
        ipv4conn: l4,
        ipv6conn: l6,
        shouldShutdown: false,
        receiver: mr,
    }
}

func (s *Server) Start() {
    go s.serve(s.ipv4conn)    
    go s.serve(s.ipv6conn)    
}

func (s *Server) Shutdown() {
    s.shouldShutdown = true
}

func (s *Server) serve(c *net.UDPConn) {
    if c == nil {
        return
    }
    
    defer c.Close()

    buf := make([]byte, 65536)
    
    for !s.shouldShutdown {
        n, from, err := c.ReadFromUDP(buf)
        if err != nil {
            log.Fatalf("Error reading from UDP socket %v\n", c)
            continue
        }
        if err := s.handlePacket(buf[:n], from); err != nil {
            log.Printf("[ERR] cluster: Failed to handle query: %v", err)
        }
    }
}

func (s *Server) handlePacket(packet []byte, from *net.UDPAddr) error {
    m, err := Unmarshall(packet)
    if err != nil {
        return err
    }

    return s.receiver.Receive(from, m)
}