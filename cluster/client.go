package cluster

import (
    "log"
    "net"
)

type Client struct {
    ipv4conn *net.UDPConn
}

func NewClient(raddr *net.UDPAddr) (*Client, error) {
    c, err := net.DialUDP("udp4", nil, raddr)

    if err != nil {
        log.Println("Cannot connect to", raddr)
        return nil, err
    }

    return &Client{
        ipv4conn: c,
    }, nil
}

func (c *Client) Send(m *Message) error {
    raw := Marshall(m)
    w, err := c.ipv4conn.Write(raw)
    if w < len(raw) || err != nil {
        log.Println("error writing message")
    }
    return err
}

func (c *Client) Close() {
    c.ipv4conn.Close()
}

