package group;

import (
    "log"
    "fmt"
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
}

type Peer struct {
    Domain *string    
    Name *string
    Group *string
}

const ServiceType = "_dominion._tcp"
const DefaultPort = 9999

var bonjourServer *bonjour.Server = nil

func NewNode(domain string, name string) *Node {
    return &Node{
        Domain: &domain,
        Name:   &name,
        Port:   DefaultPort,
        Group:  nil,
        server: nil,
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

L:
    for {
        select {
        case e := <- results:
            fmt.Printf("%s (%s) @ %s (%v:%d)\n", e.Instance, nilAs(getPeerGroup(e), "None"), e.HostName, e.AddrIPv4, e.Port)
        case <- time.After(100 * time.Millisecond):
            break L
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

func nilAs(value *string, nilValue string) string{
    if value == nil {
        return nilValue
    } else {
        return *value
    }
}
