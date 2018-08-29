// Lower level cluster node discovery through Bonjour
package cluster;

import (
    "github.com/hashicorp/serf/serf"
    "github.com/reusee/mmh3"
    "log"
    "strconv"
)

type Node struct {
    Domain *string
    Name *string
    Bind string
    Port int
    server *serf.Serf
    config *serf.Config
    eCh chan serf.Event
    Left chan Peer
    Joined chan Peer
    Peers map[string]Peer
}

type Data struct {
    SeenMembers int
}

const DefaultPort = 9999

func NewNode(name string) *Node {
    n := &Node{
        Name:           &name,
        Bind:           "127.0.0.1",
        Port:           DefaultPort,
        server:         nil,
        config:         nil,
        Peers:          map[string]Peer{},
        Left:           make(chan Peer, 10),
        Joined:         make(chan Peer, 10),
    }
    return n
}

// One-time peers discovery over the network
func (node *Node) DiscoverPeers() {
   ps := make(map[string]Peer)

   members := node.server.Members()
   for _, m := range members {
       ps[m.Name] = Peer{
           Name:         m.Name,
           Partitions:   getPeerPartitions(m),
           Port:         m.Port,
           Addr:         m.Addr,
           Tags:         m.Tags,
       }
   }

   oldPeers := node.Peers
   node.Peers = ps

   // find who left
   for name, peer := range oldPeers {
       if _, ok := ps[name]; !ok {
           node.Left <- peer
       }
   }
   // find who joined
   for name, peer := range ps {
       if _, ok := oldPeers[name]; !ok {
           node.Joined <- peer
       }
   }
}

func (node *Node) newSerf() (*serf.Serf, error) {
    conf := serf.DefaultConfig()
    conf.Init()
    conf.NodeName = *node.Name
    conf.MemberlistConfig.BindAddr = node.Bind
    conf.MemberlistConfig.BindPort = node.Port
    conf.Tags = map[string]string {
        PartitionsTag: "255",
    }

    node.eCh = make(chan serf.Event)
    conf.EventCh = node.eCh

    node.config = conf

    return serf.Create(conf)
}

// Announce the node on the network
func (node *Node) AnnouncePresence() {
    if ! node.IsAnnounced() {
        s, err := node.newSerf()

        if err != nil {
            log.Fatalln(err.Error())
        } else {
            log.Printf("Announced myself")
            node.server = s
        }
    } else {
        log.Printf("Already announced")
    }
}

// Check if the node is announced
func (node *Node) IsAnnounced() bool {
    return node.server != nil
}

// Check if the node is operational - that is it is announced and joined some group
func (node *Node) IsOperational() bool {
    return node.IsAnnounced() && node.server.State() == serf.SerfAlive
}

// Shutdown the node, opposite of announcing
func (node *Node) Shutdown() {
    if node.server != nil {
        node.server.Shutdown()
        node.server = nil
        log.Printf("Shutdown")
    }
}

func (n *Node) Hash() []byte {
    return mmh3.Sum128([]byte(*n.Name))
}

func getPeerPartitions(m serf.Member) int {
    partitions, err := strconv.Atoi(m.Tags[PartitionsTag])
    if err != nil {
        return DefaultPartitions
    }

    return partitions
}