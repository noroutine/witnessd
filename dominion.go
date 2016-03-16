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
    name := "Dominion Player"

    if portStr == "" {
        portStr = fmt.Sprintf("%d", group.DefaultPort)
    } else {
        port, err = strconv.Atoi(portStr)
        if err != nil || port <= 0 || port > 65535 {
            fmt.Printf("Invalid port: %v\n", portStr)
            os.Exit(42)
        }
    }

    client := &protocol.Client{ ":" + portStr, name }
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

    repl.EmptyHandler = func() {        
        fmt.Println("Feeling lost? Try 'help'")
        repl.EmptyHandler = nil
    }
    
    repl.Register("list", func(args []string) {
        group.NodeList("local.")
    })

    repl.Register("register", func(args []string) {
        group.NodeRegister(name, port)
    })

    repl.Register("leave", func(args []string) {
        group.NodeLeave()
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
