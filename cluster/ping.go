package cluster

import (
    "log"
    "net"
    "time"
    "github.com/noroutine/dominion/fsa"
)

const (
    START     = iota
    SENT_PING = iota
    WAIT_PONG = iota
    RCVD_PONG = iota
    SUCCESS   = iota
    TIMEOUT   = iota
    ERROR     = iota
)

type ClientPingActivity struct {
    Result chan int
    c *Cluster
    fsa *fsa.FSA
}

type ServerPingActivity struct {
    c *Cluster
}

func NewPingClient(c *Cluster) *ClientPingActivity {
    return &ClientPingActivity{
        Result: make(chan int, 1),
        c: c,
        fsa: nil,
    }
}
func (a *ClientPingActivity) Receive(from *net.UDPAddr, m *Message) error {
    switch m.Operation {
        case 0:
            pongAddr, err := a.c.GetPeerAddr(string(m.Load))
            if err != nil {
                log.Fatal("Cannot pong peer")
            }

            // send pong back
            go a.c.Send(pongAddr, &Message{
                Version: 1,
                Type: PING,
                Operation: 1,   // pong
                Length: 0,
                Load: make([]byte, 0, 0),
            })
        case 1:
            go a.fsa.Send(RCVD_PONG)
    }
    return nil
}

func (a *ClientPingActivity) Run(target string) {
    timeoutFunc := func(state int) (<-chan time.Time, func(int) int) {
        if state == WAIT_PONG {
            return time.After(100*time.Millisecond), func(s int) int {
                a.Result <- TIMEOUT
                return TIMEOUT
            }
        }

        return fsa.NeverTimesOut()(state)
    }

    a.fsa = fsa.New(func(state, input int) int {
        switch{
        case state == START && input == START:
            targetAddr, err := a.c.GetPeerAddr(target)
            if err != nil {
                log.Println("Cannot ping peer")
                a.Result <- ERROR
                return ERROR
            }

            // send ping 
            go a.c.Send(targetAddr, &Message{
                Version: 1,
                Type: PING,
                Operation: 0,   // ping
                Length: uint16(len(*a.c.proxy.Name)),
                Load: []byte(*a.c.proxy.Name),
            })

            go a.fsa.Send(SENT_PING)
            return SENT_PING
        case state == SENT_PING && input == SENT_PING:
            return WAIT_PONG
        case state == WAIT_PONG && input == RCVD_PONG:            
            go a.fsa.Send(SUCCESS)
            return RCVD_PONG
        case state == RCVD_PONG && input == SUCCESS:
            a.Result <- SUCCESS
            return SUCCESS
        }
        log.Println("Invalid automat")
        a.Result <- ERROR
        return ERROR
    }, fsa.TerminatesOn(TIMEOUT, SUCCESS, ERROR), timeoutFunc)

    go a.fsa.Send(START)
}