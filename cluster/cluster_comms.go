package cluster

import (
	"errors"
	"github.com/hashicorp/serf/serf"
	"log"
)

type Request struct {
	Event serf.Event
	Message *Message
}

type Handler interface {
	Handle(*Request) error
}

type Router interface {
	Route(*Request) (Handler, error)
}

//Connects to the cluster and start responding for cluster communications
func (c *Cluster) Join(existing []string) error {
	if ! c.proxy.IsAnnounced() {
		c.proxy.AnnouncePresence()
	}

	c.proxy.server.Join(existing, true)

	// hook into EventCh
	go c.listenOnEventCh()

	return nil
}

func (c *Cluster) listenOnEventCh() {
	c.shutdownCh = make(chan int)
	for {
		select {
		case event := <- c.proxy.eCh:
			switch event.EventType() {
			case serf.EventUser, serf.EventQuery:
				log.Printf("Intercepted %s: %s", event.EventType(), event.String())

				var m *Message
				var err error

				if event.EventType() == serf.EventUser {
					userEvent := event.(serf.UserEvent)
					m, err = Unmarshall(userEvent.Payload)
				}

				if event.EventType() == serf.EventQuery {
					queryEvent := event.(*serf.Query)
					m, err = Unmarshall(queryEvent.Payload)
				}

				if m == nil {
					log.Printf("Intercepted %s: %s, with no message", event.EventType(), event.String())
					break
				}

				r := &Request{
					Event: event,
					Message: m,
				}

				h, err := c.Route(r)
				if err != nil {
					log.Println(err)
					continue
				}

				err = h.Handle(r)
				if err != nil {
					log.Println(err)
				}
			}

		case <- c.shutdownCh:
			log.Printf("Shutting down event handlers\n")
			break
		}
	}
}

// Disconnect from the cluster and stop responding to cluster communications
func (c *Cluster) Disconnect() {
	// stop EventCh listener
	c.shutdownCh <- 1

	c.proxy.server.Leave()
	c.proxy.server.Shutdown()

	c.proxy.server = nil
}

// Route cluster request to concrete handler, makes Cluster a Router
func (c *Cluster) Route(r *Request) (h Handler, err error) {
	for e := c.handlers.Front(); e != nil; e = e.Next() {
		h, err := e.Value.(Router).Route(r)
		if h != nil && err == nil {
			return h, nil
		}
	}

	return nil, errors.New("Not supported")
}
