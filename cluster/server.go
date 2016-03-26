package cluster

import (
    "net"
    "log"
)

type Server struct {
    ipv4conn *net.UDPConn
    ipv6conn *net.UDPConn
    shouldShutdown bool
    Messages chan *Message
}

func NewServer(ip4 net.IP, ip6 net.IP, port int) *Server {
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
        Messages: make(chan *Message),
    }
}

func (s *Server) Start() {
    if s.ipv4conn != nil {
        go s.serve(s.ipv4conn)
    }
    
    if s.ipv6conn != nil {
        go s.serve(s.ipv6conn)    
    }
}

func (s *Server) Shutdown() {
    s.shouldShutdown = true
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

    close(s.Messages)
}

func (s *Server) handlePacket(packet []byte, from net.Addr) error {
    m, err := Unmarshall(packet)
    if err != nil {
        return err
    }

    s.Messages <- m

    return nil
}