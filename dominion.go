package dominion

import (
	"fmt"
	"log"
	"os"
	"time"
	"net/http"
	"html"
	"os/signal"
	"strings"

	"github.com/fiorix/go-readline"
	"github.com/noroutine/bonjour"
)

const version = "0.0.1"
const description = "Dominion " + version
const serviceType = "_dominion._tcp"
const domain = "local."
const servicePort = 9999

var phoneticAlphabet = map[string][]string{
	"e": {"exit"},
	"r": {"register"},
	"l": {"list"},
	"n": {"name"},
	"h": {"help"},
}

func completer(input, line string, start, end int) []string {
	if len(input) == 1 {
		letters, exists := phoneticAlphabet[strings.ToLower(input)]
		if exists {
			return letters
		}
	}
	return []string{}
}

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

func show_help() {
	fmt.Printf("Commands: help, name, list. register, exit\n")
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, %q!!!", html.EscapeString(r.URL.Path))
}

func main() {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
	})

	log.Fatal(http.ListenAndServe(":5000", nil))

    // Ctrl+C handling, doesn't work properly
    handler := make(chan os.Signal, 1)
    signal.Notify(handler, os.Interrupt)
    go func(handler chan os.Signal) {
    	for sig := range handler {
	        if sig == os.Interrupt {
	        	log.Printf("Interrupted")
	            os.Exit(0)
	        }
	    }
	}(handler)

	name := "Dominion Player"

	log.Println(description, "started")
	
	prompt := name + "> "

	readline.SetCompletionFunction(completer)

	// This is generally what people expect in a modern Readline-based app
	readline.ParseAndBind("TAB: menu-complete")

	// Loop until Readline returns nil (signalling EOF)
L:
	for {
		result := readline.Readline(&prompt)
		switch {
		case result == nil:
			fmt.Println()
			break L // exit loop
		case *result == "exit":
			break L // exit loop
		case *result == "list":
			service_list()
		case *result == "register":
			service_register(name)
		case *result == "help" || *result == "":
			show_help()
		case strings.HasPrefix(*result, "name"):
			name_args := strings.Fields(*result)
			if len(name_args) > 1 {
				name = name_args[1]
				fmt.Println("You are now", name)
				prompt = name + "> "
			} else {
				fmt.Println(name)
			}			
		default:
			fmt.Printf("Unknown command '%s', try 'help'\n", *result)
			continue
		}
		readline.AddHistory(*result) // Allow user to recall this line
	}
}
