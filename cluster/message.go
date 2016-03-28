package cluster

import (
    "errors"
)

type OperationType byte

const (
    NOOP  OperationType = iota
    PING                          // ping operations
    JOIN                          // join operations
)

type Message struct {
    Version     byte               //   1 byte
    Type        OperationType      // + 1 bytes    = 2
    Operation   byte               // + 1 bytes    = 3
//  reserved1   byte               // + 1 byte     = 4
    Args        []byte             // + 8 bytes    = 12
//  reserved2   []byte             // + 8 bytes    = 20
    Length      uint16             // + 2 bytes    = 22
    Load        []byte
}

const HeaderSize = 22
const MaxLoadLength = 0xFFFF - HeaderSize

func Unmarshall(packet []byte) (m *Message, err error) {
    if len(packet) < HeaderSize {
        err = errors.New("Packet is too small")
        return
    }

    return &Message{
        Version:    packet[0],
        Type:       OperationType(packet[1]),
        Operation:  packet[2],
        Args:       packet[4:12],
        Length:     uint16(packet[20]) << 8 | uint16(packet[21]),
        Load:       packet[22:],
    }, nil
}

func Marshall(m *Message) []byte {
    l := HeaderSize + len(m.Load)
    buf := make([]byte, l, l)
    buf[0] = m.Version
    buf[1] = byte(m.Type)
    buf[2] = byte(m.Operation)
    // byte 3 is reserved
    if len(m.Args) > 8 {
        panic("Too many message arguments, max 8 allowed")
    }
    for i := range(m.Args) {
        buf[4 + i] = m.Args[i]
    }
    // bytes 12 to 19 are reserved
    buf[20] = byte(m.Length >> 8)
    buf[21] = byte(m.Length)

    if len(m.Load) > MaxLoadLength {
        panic("Message data is too big")
    }

    for i, p := range m.Load {
        buf[22 + i] = p
    }

    return buf
}