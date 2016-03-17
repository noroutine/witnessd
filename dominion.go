package main

import (
    "fmt"
    "log"
    "os"
    "strings"
    "strconv"

    "github.com/noroutine/dominion/protocol"
    "github.com/noroutine/dominion/cli"
    "github.com/noroutine/dominion/group"

    "github.com/reusee/mmh3"
)

const version = "0.0.7"
const description = "Dominion " + version

func main() {

    var err error

    portStr := os.Getenv("PORT")
    port := group.DefaultPort
    name := "Player"

    if portStr == "" {
        portStr = fmt.Sprintf("%d", group.DefaultPort)
    } else {
        port, err = strconv.Atoi(portStr)
        if err != nil || port <= 0 || port > 65535 {
            fmt.Printf("Invalid port: %v\n", portStr)
            os.Exit(42)
        }
    }

    client := protocol.NewClient( ":" + portStr, name )
    go client.Serve()

    repl := cli.New()
    repl.Description = description
    repl.Prompt = name + "> "

    go func() {
        for s := range repl.Signals {
            if s == os.Interrupt {
                log.Printf("Interrupted")
                os.Exit(42)
            }
        }
    }()

    var node = group.NewNode("local.", name)
    node.Port = port

    repl.EmptyHandler = func() {        
        fmt.Println("Feeling lost? Try 'help'")
        repl.EmptyHandler = nil
    }
    
    repl.Register("peers", func(args []string) {
        for _, peer := range node.Peers {
            fmt.Printf("%s (%s) (%s:%d)\n", *peer.Name, peer.GroupOrNone(), *peer.HostName, peer.Port)
        }
    })

    repl.Register("games", func(args []string) {
        games := make(map[string][]group.Peer)

        for _, peer := range node.Peers {
            if peer.Group == nil {
                continue
            }

            peers := games[*peer.Group]

            if peers == nil {
                games[*peer.Group] = []group.Peer { peer }
            } else {
                games[*peer.Group] = append(peers, peer)
            }
        }

        for game, peers := range games {
            fmt.Printf("%s (%d are playing)\n", game, len(peers))
        }
    })

    repl.Register("announce", func(args []string) {
        node.AnnouncePresence()
    })

    repl.Register("leave", func(args []string) {
        node.Leave()
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
            node.AnnounceName(name)
        } else {
            fmt.Println(name)
        }
    })

    repl.Register("group", func(args []string) {
        var group string
        if len(args) > 0 {
            group = args[0]
            fmt.Println("Your group is now", group)
            node.AnnounceGroup(group)
        } else {
            if (node.Group == nil) {
                group = "None"
            } else {
                group = *node.Group
            }
            fmt.Println(group)
        }
    })

    repl.Serve()
}
