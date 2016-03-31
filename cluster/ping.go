// Implementation of ping operation
package cluster

import (
    "log"
    "fmt"
    "time"
    "errors"
    "github.com/noroutine/dominion/fsa"
)

const (
    START     = iota
    SENT_PING
    WAIT_PONG
    RCVD_PONG
    SUCCESS
    TIMEOUT
    ERROR
)

type PongActivity struct {
    c *Cluster
}

func NewPongActivity(c *Cluster) *PongActivity {
    return &PongActivity{
        c: c,
    }
}

func (a *PongActivity) Route(r *Request) (h Handler, err error) {
    // reply to ping with pong
    if r.Message.Type == PING && r.Message.Operation == 0 {
        return a, nil
    }

    return nil, errors.New("Cannot handle this")
}

func (a *PongActivity) Handle(r *Request) error {
    peer := string(r.Message.Load)
    pongAddr, err := a.c.GetPeerAddr(peer)
    if err != nil {
       return errors.New(fmt.Sprintf("Cannot pong peer %s", peer))
    }

    // send pong back
    go a.c.Send(pongAddr, &Message{
        Version: 1,
        Type: PING,
        Operation: 1,   // pong
        Length: 0,
        Load: make([]byte, 0, 0),
    })

    return nil
}

type PingActivity struct {
    Result chan int
    c *Cluster
    fsa *fsa.FSA
}

func NewPingActivity(c *Cluster) *PingActivity {
    return &PingActivity{
        Result: make(chan int, 1),
        c: c,
        fsa: nil,
    }
}

func (a *PingActivity) Route(r *Request) (h Handler, err error)  {
    if r.Message.Type == PING && r.Message.Operation == 1 {
        return a, nil
    }

    return nil, errors.New("Cannot handle this")
}

func (a *PingActivity) Handle(r *Request) error {
    go a.fsa.Send(RCVD_PONG)
    return nil
}

func (a *PingActivity) Run(target string) {
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