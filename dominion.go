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

func service_list() {
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
    		fmt.Printf("%s @ %s (%v)\n", e.Instance, e.HostName, e.AddrIPv4)
    	case <- time.After(5 * time.Second):
    		break L
    	}
    }
}

func service_register(name string) {
	// Run registration (blocking call)
    _, err := bonjour.Register(name, serviceType, "", servicePort, []string{"txtv=1", "app=test"}, nil)
    if err != nil {
        log.Fatalln(err.Error())
    }
    log.Printf("Registered")
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
		service_list()
	})

	repl.Register("register", func(args []string) {
		service_register(name)
	})

	repl.Register("help", func(args []string) {
		fmt.Printf("Commands: %s\n", strings.Join(repl.GetKnownCommands(), ", "))
	})

	repl.Register("sleep", func(args []string) {
		fmt.Println("Sleeping for 5 seconds...")
		time.Sleep(5 * time.Second)
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
			client.PlayerId = name
		} else {
			fmt.Println(name)
		}
	})

	repl.Serve()
}
