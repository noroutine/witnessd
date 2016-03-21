package main

import (
    "fmt"
    "log"
    "os"
    "strings"
    "flag"
    "strconv"
    "math/rand"
    "time"

    "github.com/noroutine/dominion/protocol"
    "github.com/noroutine/dominion/cli"
    "github.com/noroutine/dominion/group"
    "github.com/noroutine/ffhash"

    "github.com/reusee/mmh3"
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

    flag.IntVar(&opts.port, "port", group.DefaultPort, "client API port")
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

    var node = group.NewNode("local.", opts.name)
    node.Port = opts.port

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

    repl.Register("hashstats", func(args []string) {
        if len(args) < 2 {
            fmt.Println("Need two integer arguments")
            return
        }

        var keySpace uint64 = 0xFFFFFFFFFFFFFFFF
        kss, err := strconv.ParseUint(args[0], 10, 64)
        n, err := strconv.ParseUint(args[1], 10, 64)

        rng := rand.New(rand.NewSource(time.Now().UnixNano()))

        var randomKey uint64
        if err != nil {
            fmt.Println(err)
            return
        }

        loads := make(map[uint64]uint64)
        startTime := time.Now()
        for k := uint64(0); k < kss; k++ {
            randomKey = uint64(rng.Uint32()) << 32 + uint64(rng.Uint32())
            hash := ffhash.Sum64(randomKey, uint64(n))

            nodeLoad, ok := loads[hash]
            if ok {
                loads[hash] = nodeLoad + 1
            } else {
                loads[hash] = 1
            }
        }
        endTime := time.Now()

        spentTime := (endTime.UnixNano() - startTime.UnixNano()) / 1000
        totalBuckets := ffhash.Fact(n)
        bucketRange := keySpace / totalBuckets
        fmt.Printf("bucketRange: %v, buckets: %v, buckets/node: %v\n", bucketRange, totalBuckets, totalBuckets/n)
        fmt.Printf("hash/ms: %f, ns/hash: %v\n", float64(kss)/float64(spentTime), 1000*float64(spentTime)/float64(kss))
        fmt.Println("Keyspace distribution")
        idealLoad := float64(kss) / float64(n)
        for node, load := range loads {
            var deviation float64 = (float64(load) - idealLoad) / idealLoad * 100
            fmt.Printf("  %d : %d, deviation: %.2f%%\n", node, load, deviation)
        }
    })

    repl.Register("mmstats", func(args []string) {
        if len(args) < 1 {
            fmt.Println("Need an integer argument")
            return
        }

        kss, err := strconv.ParseUint(args[0], 10, 64)
        if err != nil {
            fmt.Println("Need number")
            return
        }
        rng := rand.New(rand.NewSource(time.Now().UnixNano()))

        startTime := time.Now()
        for k := uint64(0); k < kss; k++ {
            mmh3.Sum128([]byte {
                byte(rng.Intn(256)), 
                byte(rng.Intn(256)),
                byte(rng.Intn(256)),
                byte(rng.Intn(256)),
            })
        }
        endTime := time.Now()

        spentTime := (endTime.UnixNano() - startTime.UnixNano()) / 1000
        fmt.Printf("hash/ms: %f, ns/hash: %v\n", float64(kss)/float64(spentTime), 1000*float64(spentTime)/float64(kss))
    })


    repl.Register("hash", func(args []string) {
        if len(args) < 2 {
            fmt.Println("Need two integer arguments")
            return
        }

        k, err := strconv.ParseUint(args[0], 10, 64)
        n, err := strconv.ParseUint(args[1], 10, 64)
        
        if err != nil {
            fmt.Println(err)
            return
        }

        fmt.Printf("hash(%d, %d) = %d\n", k, n, ffhash.Sum64(k, n))
    
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
