package group;

import (
    "log"
    "fmt"
    "time"
    "github.com/noroutine/bonjour"
)

type Node struct {
    Id *string
    GroupId *string
    Name *string
}

type Group struct {
    Id *string
    Name *string
}

const ServiceType = "_dominion._tcp"
const DefaultPort = 9999

var bonjourServer *bonjour.Server = nil

func NodeList(domain string) {
    resolver, err := bonjour.NewResolver(nil)
    if err != nil {
        log.Println("Failed to initialize resolver:", err.Error())
        return
    }

    results := make(chan *bonjour.ServiceEntry)
    err = resolver.Browse(ServiceType, domain, results)

    if err != nil {
        log.Println("Failed to browse:", err.Error())
        return
    }

L:
    for {
        select {
        case e := <- results:
            fmt.Printf("%s @ %s (%v:%d)\n", e.Instance, e.HostName, e.AddrIPv4, e.Port)
        case <- time.After(100 * time.Millisecond):
            break L
        }
    }
}

func NodeRegister(name string, port int) {
    // Run registration (blocking call)
    if bonjourServer == nil {
        s, err := bonjour.Register(name, ServiceType, "", port, []string{"txtv=1", "app=test"}, nil)
        bonjourServer = s
        if err != nil {
            log.Fatalln(err.Error())
        }
        log.Printf("Registered")
    } else {
        log.Printf("Already registered")
    }
}

func NodeLeave() {
    if bonjourServer != nil {
        bonjourServer.Shutdown()
        bonjourServer = nil        
        log.Printf("Left")
    }
}
