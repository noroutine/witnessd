package cluster

import (
    "github.com/noroutine/dominion/fsa"
    "errors"
    "log"
    "fmt"
    "github.com/reusee/mmh3"
    "bytes"
    "encoding/gob"
)

const (
    LOAD_START int = iota
    LOAD_SEND
    LOAD_WAIT_ACK
    LOAD_RCVD_ACK
    LOAD_RCVD_NACK
    LOAD_FULL_ACK
    LOAD_PARTIAL_ACK
    LOAD_NO_ACK
    LOAD_SUCCESS
    LOAD_PARTIAL_SUCCESS
    LOAD_FAILURE
    LOAD_ERROR
)

const (
    LOAD_OP_GET byte = iota
    LOAD_OP_ACK
    LOAD_OP_NACK
)

type BucketLoadActivity struct {
    c *Cluster
}

func NewBucketLoadActivity(c *Cluster) *BucketLoadActivity {
    return &BucketLoadActivity{
        c: c,
    }
}

func (a *BucketLoadActivity) Route(r *Request) (h Handler, err error) {
    if r.Message.Type == LOAD && r.Message.Operation == LOAD_OP_GET {
        return a, nil
    }

    return nil, errors.New("Cannot handle this")
}

func (a *BucketLoadActivity) Handle(r *Request) error {
    peer := string(r.Message.ReplyTo)
    ackAddr, err := a.c.GetPeerAddr(peer)
    if err != nil {
        return errors.New(fmt.Sprintf("Cannot ack request from %s", peer))
    }

    raw := bytes.NewBuffer(r.Message.Load)
    dec := gob.NewDecoder(raw)
    var dto StoreDTO
    err = dec.Decode(&dto)
    if err != nil {
        log.Fatal("Decode error: ", err)
    }

    data, ok := a.c.storage.Get(dto.Key)
    if ok {
        // send ack
        log.Printf("Got request for key %s, sending ACK to %s", dto.Key, peer)

        raw := new(bytes.Buffer)
        enc := gob.NewEncoder(raw)
        err := enc.Encode(StoreDTO {
            Key: dto.Key,
            Value: data,
        })

        if err != nil {
            panic(fmt.Sprintf("Can't encode data for transfer: %v", err))
        }

        if raw.Len() > MaxLoadLength {
            panic("Load is too big")
        }

        go a.c.Send(ackAddr, &Message{
            Version: 1,
            Type: LOAD,
            Operation: LOAD_OP_ACK,
            ReplyTo: *a.c.proxy.Name,
            Length: uint16(raw.Len()),
            Load: raw.Bytes(),
        })
    } else {
        // send NACK
        log.Printf("Got request for key %s, sending NACK to %s", dto.Key, peer)
        go a.c.Send(ackAddr, &Message{
            Version: 1,
            Type: LOAD,
            Operation: LOAD_OP_NACK,
            ReplyTo: *a.c.proxy.Name,
            Length: 0,
        })
    }

    return nil
}

type LoadActivity struct {
    c *Cluster
    fsa *fsa.FSA
    acks int
    nacks int
    Result chan int
    Data []byte
}

func NewLoadActivity(c *Cluster) *LoadActivity {
    return &LoadActivity{
        Result: make(chan int, 1),
        Data: []byte {},
        acks: 2,
        nacks: 2,
        c: c,
        fsa: nil,
    }
}

func (a *LoadActivity) Route(r *Request) (h Handler, err error)  {
    if r.Message.Type == LOAD {
        return a, nil
    }

    return nil, errors.New("Cannot handle this")
}

func (a *LoadActivity) Handle(r *Request) error {
    switch r.Message.Operation {
    case LOAD_OP_ACK:
        log.Println("Received ACK from ", r.Message.ReplyTo)

        raw := bytes.NewBuffer(r.Message.Load)
        dec := gob.NewDecoder(raw)
        var dto StoreDTO
        err := dec.Decode(&dto)
        if err != nil {
            log.Fatal("Decode error: ", err)
        }

        a.Data = dto.Value

        go a.fsa.Send(LOAD_RCVD_ACK)
    case LOAD_OP_NACK:
        log.Println("Received NACK from ", r.Message.ReplyTo)
        go a.fsa.Send(LOAD_RCVD_NACK)
    }
    return nil
}

func (a *LoadActivity) Run(key []byte) {
    a.fsa = fsa.New(func(state, input int) int {
        switch{
        case state == LOAD_START && input == LOAD_START:
            go a.fsa.Send(LOAD_SEND)
            return LOAD_SEND
        case state == LOAD_SEND && input == LOAD_SEND:
            raw := new(bytes.Buffer)
            enc := gob.NewEncoder(raw)
            err := enc.Encode(StoreDTO {
                Key: key,
            })

            if err != nil {
                panic(fmt.Sprintf("Can't encode data for transfer: %v", err))
            }

            if raw.Len() > MaxLoadLength {
                panic("Load is too big")
            }

            primary, secondary := a.c.HashNodes(mmh3.Sum128(key))
            primaryAddr, err := a.c.GetPeerAddr(*primary.Name)
            if err != nil {
                log.Println("Cannot contact peer", *primary.Name)
                a.Result <- LOAD_ERROR
                return LOAD_ERROR
            }

            secondaryAddr, err := a.c.GetPeerAddr(*secondary.Name)
            if err != nil {
                log.Println("Cannot contact peer", *secondary.Name)
                a.Result <- LOAD_ERROR
                return LOAD_ERROR
            }

            // send load command to primary and secondary nodes
            m := &Message{
                Version: 1,
                Type: LOAD,
                Operation: LOAD_OP_GET,   // load
                ReplyTo: *a.c.proxy.Name,
                Length: uint16(raw.Len()),
                Load: raw.Bytes(),
            }

            go a.c.Send(primaryAddr, m)
            go a.c.Send(secondaryAddr, m)

            return LOAD_WAIT_ACK
        case state == LOAD_WAIT_ACK && input == LOAD_RCVD_ACK:
            a.acks--
            switch {
            case a.acks == 0:
                go a.fsa.Send(LOAD_FULL_ACK)
                return LOAD_FULL_ACK
            case a.acks == 1 && a.nacks == 1:
                go a.fsa.Send(LOAD_PARTIAL_ACK)
                return LOAD_PARTIAL_ACK
            case a.acks == 0 && a.nacks == 2:
                go a.fsa.Send(LOAD_NO_ACK)
                return LOAD_NO_ACK
            default:
                return LOAD_WAIT_ACK
            }
        case state == LOAD_WAIT_ACK && input == LOAD_RCVD_NACK:
            a.nacks--
            switch {
            case a.acks == 0:
                go a.fsa.Send(LOAD_FULL_ACK)
                return LOAD_FULL_ACK
            case a.acks == 1 && a.nacks == 1:
                go a.fsa.Send(LOAD_PARTIAL_ACK)
                return LOAD_PARTIAL_ACK
            case a.acks == 0 && a.nacks == 2:
                go a.fsa.Send(LOAD_NO_ACK)
                return LOAD_NO_ACK
            default:
                return LOAD_WAIT_ACK
            }
        case state == LOAD_NO_ACK:
            a.Result <- LOAD_FAILURE
            return LOAD_FAILURE
        case state == LOAD_PARTIAL_ACK:
            a.Result <- LOAD_PARTIAL_SUCCESS
            return LOAD_PARTIAL_SUCCESS
        case state == LOAD_FULL_ACK:
            a.Result <- LOAD_SUCCESS
            return LOAD_SUCCESS
        }
        log.Println("Invalid automat")
        a.Result <- LOAD_ERROR
        return LOAD_ERROR
    }, fsa.TerminatesOn(LOAD_SUCCESS, LOAD_PARTIAL_SUCCESS, LOAD_FAILURE), fsa.NeverTimesOut())

    go a.fsa.Send(LOAD_START)
}