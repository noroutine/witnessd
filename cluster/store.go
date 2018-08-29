package cluster

import (
    "errors"
    "github.com/hashicorp/serf/serf"
    "log"
    "fmt"
    "encoding/gob"
    "github.com/noroutine/witnessd/fsa"
    "github.com/reusee/mmh3"
    "bytes"
    "time"
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

const (
    STORE_OP_PUT byte = iota
    STORE_OP_ACK
)
type StoreDTO struct {
    Key []byte
    Value []byte
}

type BucketStoreActivity struct {
    c *Cluster
}

func NewBucketStoreActivity(c *Cluster) *BucketStoreActivity {
    return &BucketStoreActivity{
        c: c,
    }
}

func (a *BucketStoreActivity) Route(r *Request) (h Handler, err error) {
    if r.Message.Type == STORE && r.Message.Operation == STORE_OP_PUT {
        return a, nil
    }

    return nil, errors.New("Cannot handle this")
}

func (a *BucketStoreActivity) Handle(r *Request) error {

    query := r.Event.(*serf.Query)

    raw := bytes.NewBuffer(r.Message.Load)
    dec := gob.NewDecoder(raw)
    var dto StoreDTO
    err := dec.Decode(&dto)
    if err != nil {
        log.Fatal("Decode error: ", err)
    }

    a.c.storage.Put(dto.Key, dto.Value)

    query.Respond(Marshall(&Message{
        Version: 1,
        Type: STORE,
        Operation: STORE_OP_ACK,
        ReplyTo: *a.c.proxy.Name,
        Length: 0,
    }))

    log.Printf("Got %d bytes of data to store", r.Message.Length)

    return nil
}

type StoreActivity struct {
    c *Cluster
    level ConsistencyLevel
    fsa *fsa.FSA
    ackCh <-chan string
    responseCh <-chan serf.NodeResponse
    acks int
    Result chan int
}

func NewStoreActivity(c *Cluster, level ConsistencyLevel) *StoreActivity {
    return &StoreActivity{
        Result: make(chan int, 1),
        level: level,
        acks: c.Copies(level),
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
    //log.Println("Received ACK from ", r.Message.ReplyTo)
    go a.fsa.Send(STORE_RCVD_ACK)
    return nil
}

func (a *StoreActivity) Run(key, data []byte) {
    a.fsa = fsa.New(func(state, input int) int {
        switch{
        case state == STORE_START && input == STORE_START:
            go a.fsa.Send(STORE_SEND)
            return STORE_SEND
        case state == STORE_SEND && input == STORE_SEND:

            raw := new(bytes.Buffer)
            enc := gob.NewEncoder(raw)
            err := enc.Encode(StoreDTO {
                Key: key,
                Value: data,
            })

            if err != nil {
                panic(fmt.Sprintf("Can't encode data for transfer: %v", err))
            }

            if raw.Len() > MaxLoadLength {
                panic("Load is too big")
            }

            // send store command to primary and secondary nodes
            m := &Message{
                Version: 1,
                Type: STORE,
                Operation: STORE_OP_PUT,
                ReplyTo: *a.c.proxy.Name,
                Length: uint16(raw.Len()),
                Load: raw.Bytes(),
            }

            nodes := a.c.HashNodes(mmh3.Sum128(key), a.level)
            storageNodeNames := make([]string, len(nodes))
            for _, node := range nodes {
                storageNodeNames = append(storageNodeNames, node.Name)
            }

            a.acks = len(nodes)

            serfInstance := a.c.proxy.server
            queryParams := serfInstance.DefaultQueryParams()
            queryParams.RequestAck = true

            queryParams.FilterNodes = storageNodeNames

            resp, err := a.c.SendQuery("store", m, queryParams)
            if err != nil {
                log.Printf("Error sending serf query while storing data: %s", err)
                a.Result <- STORE_ERROR
                return STORE_ERROR
            }

            // setup ack and response channels
            a.ackCh = resp.AckCh()
            a.responseCh = resp.ResponseCh()

            var acks []string
            var responses []string

            for i := 0; i < a.acks; i++ {
                select {
                case a := <-a.ackCh:
                    log.Printf("Received ACK %s", a)
                    acks = append(acks, a)
                    // TODO: match acks with stoarage node list to make sure we get acks from specific nodes
                case r := <-a.responseCh:
                    log.Printf("Received response %v", r)
                    responses = append(responses, r.From)

                case <-time.After(time.Second):
                    log.Printf("Timeout waiting for ACKs")
                    return STORE_NO_ACK
                }
            }

            go a.fsa.Send(STORE_FULL_ACK)
            return STORE_FULL_ACK
        case state == STORE_WAIT_ACK && input == STORE_RCVD_ACK:
            a.acks--
            if a.acks == 0 {
                go a.fsa.Send(STORE_FULL_ACK)
                return STORE_FULL_ACK
            } else {
                return STORE_WAIT_ACK
            }
        case state == STORE_NO_ACK:
            a.Result <- STORE_FAILURE
            return STORE_FAILURE
        case state == STORE_PARTIAL_ACK:
            a.Result <- STORE_PARTIAL_SUCCESS
            return STORE_PARTIAL_SUCCESS
        case state == STORE_FULL_ACK:
            a.Result <- STORE_SUCCESS
            return STORE_SUCCESS
        }
        log.Println("Invalid automaton")
        a.Result <- STORE_ERROR
        return STORE_ERROR
    }, fsa.TerminatesOn(STORE_SUCCESS, STORE_PARTIAL_SUCCESS, STORE_FAILURE), fsa.NeverTimesOut())

    go a.fsa.Send(STORE_START)
}