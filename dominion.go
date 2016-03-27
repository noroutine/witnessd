package main

import (
    "fmt"
    "log"
    "os"
    "strings"
    "flag"
    
    "github.com/noroutine/go-cli"
    "github.com/noroutine/dominion/protocol"
    "github.com/noroutine/dominion/cluster"
)

const version = "0.0.7"
const description = "Dominion " + version

type Options struct {
    port int
    name string
    join string
    announce bool
}

func main() {

    opts := Options{}

    flag.IntVar(&opts.port, "port", cluster.DefaultPort, "client API port")
    flag.StringVar(&opts.name, "name", "Player", "name of the player")
    flag.StringVar(&opts.join, "join", "", "name of the group of the node")
    flag.BoolVar(&opts.announce, "announce", false, "auto announce")
    flag.Parse()

    if opts.port <= 0 || opts.port > 65535 {
        fmt.Printf("Invalid port: %v\n", opts.port)
        os.Exit(42)
    }

    opts.join = strings.TrimSpace(opts.join)

    client := protocol.NewClient(fmt.Sprintf(":%d", opts.port), opts.name)
    go client.Serve()

    repl := cli.New()
    repl.Description = description
    repl.Prompt = opts.name + "> "

    go func() {
        for s := range repl.Signals {
            if s == os.Interrupt {
                log.Printf("Interrupted")
                os.Exit(42)
            }
        }
    }()

    node := cluster.NewNode("local.", opts.name)
    node.Port = opts.port

    var cl *cluster.Cluster

    if len(opts.join) > 0 {
        node.Group = &opts.join
    }

    node.StartDiscovery()

    if opts.announce {
        node.AnnouncePresence()
    }

    go func() {
        for {
            select {
            case peer := <- node.Joined:
                log.Println(*peer.Name, "joined")
            case peer := <- node.Left:
                log.Println(*peer.Name, "left")
            }
        }
    }()

    repl.EmptyHandler = func() {        
        fmt.Println("Feeling lost? Try 'help'")
        repl.EmptyHandler = nil
    }
    
    repl.Register("peers", func(args []string) {
        if node.Group == nil {
            fmt.Println("You are not a member of any group, use 'join'")
            return
        }

        if (! node.IsDiscoveryActive()) {
            node.DiscoverPeers()
        }

        fmt.Printf("Your peers in group %s:\n", *node.Group)

        for name, peer := range node.Peers {
            fmt.Printf("%s (%s:%d)\n", name, *peer.HostName, peer.Port)
        }
    })

    repl.Register("groups", func(args []string) {
        if (! node.IsDiscoveryActive()) {
            node.DiscoverPeers()
        }

        for name, data := range node.Groups {
            fmt.Printf("%s (%d members)\n", name, data.SeenMembers)
        }
    })

    repl.Register("announce", func(args []string) {
        node.AnnouncePresence()
    })

    repl.Register("denounce", func(args []string) {
        node.Shutdown()
    })

    repl.Register("group", func(args []string) {
        var group string
        if (node.Group == nil) {
            group = "None"
        } else {
            group = *node.Group
        }
        fmt.Println(group)
    })

    repl.Register("join", func(args []string) {
        var group string
        if len(args) > 0 {
            group = args[0]
            fmt.Println("Your group is now", group)
            node.AnnounceGroup(&group)
        } else {
            fmt.Println("Provide group name, discover with 'groups")
        }
    })

    repl.Register("leave", func(args []string) {
        node.AnnounceGroup(nil)
    })

    repl.Register("cluster_start", func(args []string) {
        var err error
        cl, err = cluster.NewVia(node)
        if err != nil {
            fmt.Println(err)
        }

        cl.Connect()
    })

    repl.Register("cluster_stop", func(args []string) {
        if cl != nil {
            cl.Disconnect()
            cl = nil            
        }
    })

    repl.Register("ping", func(args []string) {
        if len(args) < 1 {
            fmt.Println("Usage: ping <peer>")
        }

        cl.Ping(args[0])
    })

    repl.Register("help", func(args []string) {
        fmt.Printf("Commands: %s\n", strings.Join(repl.GetKnownCommands(), ", "))
    })

    repl.Register("name", func(args []string) {
        if len(args) > 0 {
            name := args[0]
            fmt.Println("You are now", name)
            repl.Prompt = name + "> "
            client.PlayerID = name
            node.AnnounceName(name)
        } else {
            fmt.Println(node.Name)
        }
    })

    repl.Serve()
}
