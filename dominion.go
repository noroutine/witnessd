package main

import (
    "fmt"
    "log"
    "os"
    "time"
    "strings"

    "github.com/noroutine/bonjour"
    "github.com/noroutine/dominion/protocol"
    "github.com/noroutine/dominion/cli"

    "github.com/reusee/mmh3"
)

const version = "0.0.7"
const description = "Dominion " + version
const serviceType = "_dominion._tcp"
const domain = "local."
const servicePort = 9999

type GameService struct {
    BonjourServer *bonjour.Server
}

func (gameService *GameService) serviceList() {
    resolver, err := bonjour.NewResolver(nil)
    if err != nil {
        log.Println("Failed to initialize resolver:", err.Error())
        return
    }

    results := make(chan *bonjour.ServiceEntry)
    err = resolver.Browse(serviceType, domain, results)

    if err != nil {
        log.Println("Failed to browse:", err.Error())
        return
    }

L:
    for {
        select {
        case e := <- results:
            fmt.Printf("%s @ %s (%v:%d)\n", e.Instance, e.HostName, e.AddrIPv4, e.Port)
        case <- time.After(1 * time.Second):
            break L
        }
    }
}

func (gameService *GameService) serviceRegister(name string) {
    // Run registration (blocking call)
    s, err := bonjour.Register(name, serviceType, "", servicePort, []string{"txtv=1", "app=test"}, nil)
    // s.TTL(0)
    if err != nil {
        log.Fatalln(err.Error())
    }
    gameService.BonjourServer = s
    log.Printf("Registered")
}

func (gameService *GameService) serviceShutdown() {
    if gameService.BonjourServer != nil {
        gameService.BonjourServer.Shutdown()
        gameService.BonjourServer = nil
        log.Printf("Unregistered")
    } else {
        log.Printf("Not registered")
    }

}

func main() {

    port := os.Getenv("PORT")
    name := "Dominion Player"

    if port == "" {
        port = fmt.Sprintf("%d", servicePort)
    }

    client := &protocol.Client{ ":" + port, name }
    go client.Serve()

    repl := cli.New()
    repl.Description = description
    repl.Prompt = name + "> "

    gameService := &GameService{}

    go func() {
        for s := range repl.Signals {
            if s == os.Interrupt {
                log.Printf("Interrupted")
                os.Exit(0)
            }
        }
    }()

    repl.EmptyHandler = func() {        
        fmt.Println("Feeling lost? Try 'help'")
        repl.EmptyHandler = nil
    }
    
    repl.Register("list", func(args []string) {
        gameService.serviceList()
    })

    repl.Register("register", func(args []string) {
        gameService.serviceRegister(name)
    })

    repl.Register("unregister", func(args []string) {
        gameService.serviceShutdown()
    })

    repl.Register("help", func(args []string) {
        fmt.Printf("Commands: %s\n", strings.Join(repl.GetKnownCommands(), ", "))
    })

    repl.Register("mmh3", func(args []string) {
        key := ""
        if len(args) > 0 {
            key = args[0]
        }

        fmt.Printf("murmur3(\"%s\") = %x\n", key, mmh3.Sum128([]byte(key)))
    })

    repl.Register("name", func(args []string) {
        if len(args) > 0 {
            name = args[0]
            fmt.Println("You are now", name)
            repl.Prompt = name + "> "
            client.PlayerID = name
        } else {
            fmt.Println(name)
        }
    })

    repl.Serve()
}
