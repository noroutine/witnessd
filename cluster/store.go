package cluster

import (
    "github.com/noroutine/dominion/fsa"
    "errors"
    "log"
)

const (
    STORE_START int = iota
    STORE_SEND
    STORE_WAIT_ACK
    STORE_RCVD_ACK
    STORE_FULL_ACK
    STORE_PARTIAL_ACK
    STORE_NO_ACK
    STORE_SUCCESS
    STORE_PARTIAL_SUCCESS
    STORE_FAILURE
    STORE_ERROR
)

type BucketActivity struct {
    c *Cluster
}

func NewBucketActivity(c *Cluster) *BucketActivity {
    return &BucketActivity{
        c: c,
    }
}

func (a *BucketActivity) Route(r *Request) (h Handler, err error) {
    // reply to ping with pong
    if r.Message.Type == STORE && r.Message.Operation == 0 {
        return a, nil
    }

    return nil, errors.New("Cannot handle this")
}

func (a *BucketActivity) Handle(r *Request) error {
    log.Println("Store action is not implemented")
    return nil
}

type StoreActivity struct {
    c *Cluster
    fsa *fsa.FSA
    Result chan int
}

func NewStoreActivity(c *Cluster) *StoreActivity {
    return &StoreActivity{
        Result: make(chan int, 1),
        c: c,
        fsa: nil,
    }
}

func (a *StoreActivity) Route(r *Request) (h Handler, err error)  {
    if r.Message.Type == STORE && r.Message.Operation == 1 {
        return a, nil
    }

    return nil, errors.New("Cannot handle this")
}

func (a *StoreActivity) Handle(r *Request) error {
    go a.fsa.Send(STORE_RCVD_ACK)
    return nil
}

func (a *StoreActivity) Run() {
    a.fsa = fsa.New(func(state, input int) int {
        log.Println("Store:", state)
        switch{
        case state == STORE_START && input == STORE_START:
            go a.fsa.Send(STORE_SEND)
            return STORE_SEND
        case state == STORE_SEND && input == STORE_SEND:
            go a.fsa.Send(STORE_WAIT_ACK)
            return STORE_WAIT_ACK
        case state == STORE_WAIT_ACK && input == STORE_WAIT_ACK:
            go a.fsa.Send(STORE_FULL_ACK)
            return STORE_FULL_ACK
        case state == STORE_RCVD_ACK && input == STORE_RCVD_ACK:
            go a.fsa.Send(STORE_WAIT_ACK)
            return STORE_WAIT_ACK
        case state == STORE_NO_ACK && input == STORE_NO_ACK:
            a.Result <- STORE_FAILURE
            return STORE_FAILURE
        case state == STORE_PARTIAL_ACK && input == STORE_PARTIAL_ACK:
            a.Result <- STORE_PARTIAL_SUCCESS
            return STORE_PARTIAL_SUCCESS
        case state == STORE_FULL_ACK && input == STORE_FULL_ACK:
            a.Result <- STORE_SUCCESS
            return STORE_SUCCESS
        }
        log.Println("Invalid automat")
        a.Result <- STORE_ERROR
        return STORE_ERROR
    }, fsa.TerminatesOn(STORE_SUCCESS, STORE_PARTIAL_SUCCESS, STORE_FAILURE), fsa.NeverTimesOut())

    go a.fsa.Send(PING_START)
}