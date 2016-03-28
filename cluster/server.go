package cluster

import (
    "net"
    "log"
)

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

    if s.ipv6conn != nil {
        s.ipv6conn.Close()
    }

    if s.ipv4conn != nil {
        s.ipv4conn.Close()
    }
}

func (s *Server) serve(c *net.UDPConn) {
    if c == nil {
        return
    }

    buf := make([]byte, 65536)
    
    for {
        n, from, err := c.ReadFromUDP(buf)
        if err != nil {
            continue
        }
        if err := s.handlePacket(buf[:n], from); err != nil {
            log.Printf("[ERR] cluster: Failed to handle query: %v\n", err)
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