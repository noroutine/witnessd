package cluster

import (
    "log"
    "net"
    "time"
    "github.com/noroutine/dominion/fsa"
)

type PingActivity struct {
    c *Cluster
    target *net.UDPAddr
}

func NewPingActivity(target *net.UDPAddr, c *Cluster) *PingActivity {
    return &PingActivity{
        c: c,
        target: target,
    }
}

func (a *PingActivity) Client() *fsa.FSA {
    const (
        START     = iota
        SENT_PING = iota
        WAIT_PONG = iota
        RCVD_PONG = iota
        TIMEOUT   = iota
        ERROR     = iota
    )

    return fsa.New(func(state, input int) int {
        switch{
        case state == START && input == START:
            log.Println("START")
            // send ping 
            a.c.Send(a.target, &Message{
                Version: 1,
                Type: PING,
                Operation: 0,
                Length: 0,
                Load: make([]byte, 0, 0),
            })
            return SENT_PING
        case state == SENT_PING:
            log.Println("SENT_PING")
            return WAIT_PONG
        case state == WAIT_PONG && input == RCVD_PONG:
            // shall somehow get a notification from cluster
            log.Println("WAIT_PONG")
            return RCVD_PONG
        case state == TIMEOUT:
            log.Println("TIMEOUT")
        case state == ERROR:
            log.Fatal("ERROR")
        }
        return ERROR
    }, fsa.TerminatesOn(TIMEOUT, RCVD_PONG, ERROR), func(state int) chan int {
        if state == WAIT_PONG {
            pingTimeoutCh := make(chan int)
            go func(ch chan int) {
                <- time.After(100*time.Millisecond)
                ch <- TIMEOUT
                close(ch)
            }(pingTimeoutCh)
            return pingTimeoutCh
        }
        return make(chan int, 0)
    })
}

func (a *PingActivity) Server() *fsa.FSA {
    const (
        PONG_SENT = iota
    )

    return fsa.New(func(state, input int) int {
        return 0
    }, fsa.TerminatesOn(0), fsa.NeverTimesOut())
}