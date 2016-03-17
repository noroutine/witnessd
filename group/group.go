package group;

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
    Peers []Peer
}

type Peer struct {
    Domain *string    
    Name *string
    HostName *string
    Port int
    Group *string

    serviceEntry *bonjour.ServiceEntry
}

const ServiceType = "_dominion._tcp"
const DefaultPort = 9999
const browseWindow = 200 * time.Millisecond
const discoveryInterval = 5 * time.Second

var bonjourServer *bonjour.Server = nil

func NewNode(domain string, name string) *Node {
    return &Node{
        Domain:         &domain,
        Name:           &name,
        Port:           DefaultPort,
        Group:          nil,
        server:         nil,
        discoverLoopCh: nil,
        Peers:          []Peer{},
    }
}

func (node *Node) DiscoverPeers() {
    resolver, err := bonjour.NewResolver(nil)
    if err != nil {
        log.Println("Failed to initialize resolver:", err.Error())
        return
    }

    results := make(chan *bonjour.ServiceEntry)
    err = resolver.Browse(ServiceType, *node.Domain, results)

    if err != nil {
        log.Println("Failed to browse:", err.Error())
        return
    }

    node.Peers = make([]Peer, 0, 10)

L:
    for {
        select {
        case e := <- results:
            node.Peers = append(node.Peers, Peer{
                    Domain: node.Domain,
                    Name: &e.Instance,
                    Group: getPeerGroup(e),
                    HostName: &e.HostName,
                    Port: e.Port,
                    serviceEntry: e,
                })
        case <- time.After(browseWindow):
            break L
        }
    }

    log.Println("Discovered", len(node.Peers), "peers")
}

func (node *Node) peerDiscoveryLoop(quit chan int) {
    for {
        select {            
        case <- time.Tick(discoveryInterval):
            node.DiscoverPeers()
        case <- quit:
            return
        }
    }
}

func (node *Node) AnnouncePresence() {
    // Run registration (blocking call)
    if node.server == nil {
        text := []string {}
        if node.Group != nil {
            text = []string { "group=" + *node.Group }
        }        
        s, err := bonjour.Register(*node.Name, ServiceType, "", node.Port, text, nil)
        if err != nil {
            log.Fatalln(err.Error())
        } else {
            log.Printf("Registered")
            node.server = s
        }
        node.discoverLoopCh = make(chan int, 1)
        go node.peerDiscoveryLoop(node.discoverLoopCh)
    } else {
        log.Printf("Already registered")
    }
}

func (node *Node) AnnounceName(newName string) {
    if node.server != nil {
        node.Leave()
        node.Name = &newName
        node.AnnouncePresence()
    } else {
        node.Name = &newName
    }
}

func (node *Node) AnnounceGroup(newGroup string) {
    node.Group = &newGroup
    if (node.server != nil) {
        node.server.SetText([]string{ "group=" + newGroup })    
    }
}

func (node *Node) Leave() {
    if node.server != nil {
        node.discoverLoopCh <- 0
        node.discoverLoopCh = nil
        node.server.Shutdown()
        node.server = nil
        log.Printf("Left")
    }
}

func getPeerGroup(e *bonjour.ServiceEntry) *string {
    for _, s := range e.Text {
        if strings.HasPrefix(s, "group=") {            
            group := strings.TrimPrefix(s, "group=")
            return &group
        }
    }

    return nil
}

func (peer *Peer) GroupOrNone() string {
    if peer.Group == nil {
        return "None"
    } else {
        return *peer.Group
    }
}