package cluster

import (
    "log"
    "net"
    "github.com/noroutine/dominion/fsa"
)

type ClientPingActivity struct {
    pong chan bool
    c *Cluster
    fsa *fsa.FSA
}

type ServerPingActivity struct {
    c *Cluster
}

const (
    START     = iota
    SENT_PING = iota
    WAIT_PONG = iota
    RCVD_PONG = iota
    TIMEOUT   = iota
    ERROR     = iota
)

func NewPingClient(c *Cluster) *ClientPingActivity {
    return &ClientPingActivity{
        c: c,
        fsa: nil,
    }
}
func (a *ClientPingActivity) Receive(from *net.UDPAddr, m *Message) error {
    log.Println("Received ping from", from.IP)
    switch m.Operation {
        case 0:

            pongAddr, err := a.c.GetPeerAddr(string(m.Load))
            if err != nil {
                log.Fatal("Cannot pong peer")
            }

            // send pong back
            a.c.Send(pongAddr, &Message{
                Version: 1,
                Type: PING,
                Operation: 1,   // pong
                Length: 0,
                Load: make([]byte, 0, 0),
            })
        case 1:
            if a.fsa != nil {
                a.fsa.Send(RCVD_PONG)
            } else {
                log.Fatal("client automat is null")
            }
    }
    return nil
}

func (a *ClientPingActivity) Run(target string) *fsa.FSA {

    a.fsa = fsa.New(func(state, input int) int {
        switch{
        case state == START && input == START:
            log.Println("START")

            targetAddr, err := a.c.GetPeerAddr(target)
            if err != nil {
                log.Fatal("Cannot ping peer")
            }

            // send ping 
            a.c.Send(targetAddr, &Message{
                Version: 1,
                Type: PING,
                Operation: 0,   // ping
                Length: uint16(len(*a.c.proxy.Name)),
                Load: []byte(*a.c.proxy.Name),
            })

            go a.fsa.Send(SENT_PING)

            return SENT_PING
        case state == SENT_PING && input == SENT_PING:
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
    }, fsa.TerminatesOn(TIMEOUT, RCVD_PONG, ERROR), fsa.NeverTimesOut())

    // func(state int) chan int {
    //     if state == WAIT_PONG {
    //         pingTimeoutCh := make(chan int)
    //         go func(ch chan int) {
    //             <- time.After(100*time.Millisecond)
    //             ch <- TIMEOUT
    //             log.Println("TIMEOUT")
    //             close(ch)
    //         }(pingTimeoutCh)
    //         return pingTimeoutCh
    //     }
    //     return make(chan int, 0)
    // })

    a.fsa.Send(START)
    return a.fsa
}

func (a ServerPingActivity) Run() *fsa.FSA {
    const (
        PONG_SENT = iota
    )

    return fsa.New(func(state, input int) int {
        return 0
    }, fsa.TerminatesOn(0), fsa.NeverTimesOut())
}