package cluster

import (
    "net"
    "github.com/noroutine/dominion/fsa"
)

type ClusterActivity interface {
    Server() *fsa.FSA
    Client() *fsa.FSA
}

type MessageReceiver interface {
    Receive(*net.UDPAddr, *Message) error
}

type MessageSender interface {
    Send(*net.UDPAddr, *Message) error
}