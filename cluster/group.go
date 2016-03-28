package cluster;

import (
    "log"
    "time"
    "strings"
    "github.com/noroutine/bonjour"
)

type Node struct {
    Domain *string
    Name *string
    Port int
    Group *string
    server *bonjour.Server
    discoverLoopCh chan int
    Left chan Peer
    Joined chan Peer
    Peers map[string]Peer
    Groups map[string]Data
}

type Data struct {
    SeenMembers int
}

const ServiceType = "_dominion._tcp"
const DefaultPort = 9999
const browseWindow = 200 * time.Millisecond
const discoveryInterval = 5 * time.Second
const groupTextParameter = "group="

func NewNode(domain string, name string) *Node {
    return &Node{
        Domain:         &domain,
        Name:           &name,
        Port:           DefaultPort,
        Group:          nil,
        server:         nil,
        discoverLoopCh: nil,
        Peers:          map[string]Peer{},
        Left:           make(chan Peer, 10),
        Joined:         make(chan Peer, 10),
        Groups:         map[string]Data{},
    }
}

func (node *Node) DiscoverPeers() {

    resolver, err := bonjour.NewResolver(nil)
    if err != nil {
        log.Fatal("Failed to initialize resolver:", err.Error())
    }

    results := make(chan *bonjour.ServiceEntry)

    err = resolver.Browse(ServiceType, *node.Domain, results)
    if err != nil {
        log.Println("Failed to browse:", err.Error())
        return
    }

    ps := make(map[string]Peer)
    gs := make(map[string]Data)
L:
    for {
        select {
        case e := <- results:
            g := getPeerGroup(e)
            if g != nil {
                gData, ok := gs[*g]
                if ! ok {
                    gs[*g] = Data{
                        SeenMembers: 1,
                    }
                } else {
                    gData.SeenMembers = gData.SeenMembers + 1
                    gs[*g] = gData
                }

                if node.Group != nil && *node.Group == *g {
                    ps[e.Instance] = Peer{
                        Domain:       node.Domain,
                        Name:         &e.Instance,
                        Group:        g,
                        HostName:     &e.HostName,
                        Port:         e.Port,
                        entry:        e,
                    }
                }
            }
        case <- time.After(browseWindow):
            resolver.Exit <- true
            break L
        }
    }

    oldPeers := node.Peers
    node.Peers = ps
    node.Groups = gs

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

func (node *Node) StartDiscovery() {
    if node.discoverLoopCh == nil {
        node.discoverLoopCh = make(chan int, 1)
        go func(quit chan int) {
            for {
                node.DiscoverPeers()
                select {
                case <- time.Tick(discoveryInterval):
                case <- quit:
                    return
                }
            }
        }(node.discoverLoopCh)
    }
}

func (node *Node) StopDiscovery() {
    if node.discoverLoopCh != nil {
        node.discoverLoopCh <- 1
        node.discoverLoopCh = nil
    }
}

func (node *Node) IsDiscoveryActive() bool {
    return node.discoverLoopCh != nil
}

func (node *Node) AnnouncePresence() {
    // Run registration (blocking call)
    if node.server == nil {
        s, err := bonjour.Register(*node.Name, ServiceType, "", node.Port, node.getGroupText(), nil)
        if err != nil {
            log.Fatalln(err.Error())
        } else {
            log.Printf("Registered")
            node.server = s
        }
    } else {
        log.Printf("Already registered")
    }
}

func (node *Node) IsAnnounced() bool {
    return node.server != nil
}

func (node *Node) IsClustered() bool {
    return node.Group != nil
}

func (node *Node) IsOperational() bool {
    return node.IsClustered() && node.IsAnnounced()
}

func (node *Node) AnnounceName(newName string) {
    if node.server != nil {
        node.Shutdown()
        node.Name = &newName
        node.AnnouncePresence()
    } else {
        node.Name = &newName
    }
}

func (node *Node) AnnounceGroup(newGroup *string) {
    node.Group = newGroup
    if (node.server != nil) {
        node.server.SetText(node.getGroupText())
    }
}

func (node *Node) GetServiceEntry() *bonjour.ServiceEntry {
    return node.Peers[*node.Name].entry
}

func (node *Node) Shutdown() {
    if node.server != nil {
        node.server.Shutdown()
        node.server = nil
        log.Printf("Shutdown")
    }
}

func getPeerGroup(e *bonjour.ServiceEntry) *string {
    for _, s := range e.Text {
        if strings.HasPrefix(s, groupTextParameter) {
            group := strings.TrimPrefix(s, groupTextParameter)
            return &group
        }
    }

    return nil
}

func (node *Node) getGroupText() []string {
    if node.Group != nil {
        return []string{ groupTextParameter + *node.Group}
    } else {
        return []string{}
    }
}