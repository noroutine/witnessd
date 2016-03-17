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
    Peers map[string]Peer
    Groups map[string]GroupData
}

type Peer struct {
    Domain *string    
    Name *string
    HostName *string
    Port int
    Group *string

    serviceEntry *bonjour.ServiceEntry
}

type GroupData struct {
    SeenMembers int
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
        Peers:          map[string]Peer{},
        Groups:         map[string]GroupData{},
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

    ps := make(map[string]Peer)
    gs := make(map[string]GroupData)
L:
    for {
        select {
        case e := <- results:
            g := getPeerGroup(e)
            if g != nil {
                gData, ok := gs[*g]
                if ! ok {
                    gs[*g] = GroupData{
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
                        serviceEntry: e,
                    }
                }
            }
        case <- time.After(browseWindow):
            break L
        }
    }

    // find who left
    for name := range node.Peers {
        if _, ok := ps[name]; !ok {
            log.Println(name, "left")
        }
    }
    // find who joined
    for names := range ps {
        if _, ok := node.Peers[name]; !ok {
            log.Println(name, "joined")
        }
    }

    node.Peers = ps
    node.Groups = gs
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
    } else {
        log.Printf("Already registered")
    }
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
        if node.Group != nil {
            node.server.SetText([]string{ "group=" + *newGroup })
        } else {
            node.server.SetText([]string{})
        }
    }
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
        if strings.HasPrefix(s, "group=") {
            group := strings.TrimPrefix(s, "group=")
            return &group
        }
    }

    return nil
}