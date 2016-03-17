package cli

import (
	"log"
	"fmt"
	"strings"
	"os"
	"os/signal"
	"github.com/noroutine/go-readline"
)

type HandlerFunc func([]string)

type REPL struct {
	Description string
	Prompt string
	Signals chan os.Signal
	EmptyHandler func()	
	handlers map[string]HandlerFunc
}

func New() *REPL {
	return &REPL{
		Description:  "",
		Prompt:       "> ",
		Signals:      make(chan os.Signal, 1),
		EmptyHandler: nil,
		handlers:     map[string]HandlerFunc{},
	}
}

func (repl *REPL) Register(command string, handler HandlerFunc) {
	repl.handlers[command] = handler 
}

func (repl *REPL) GetKnownCommands() []string {
	commands := make([]string, len(repl.handlers))
	i := 0
	for command := range repl.handlers {
		commands[i] = command
		i++
	}	
	return commands
}

func (repl *REPL) understands(commandWithArgs string) bool {
	if commandWithArgs == "" {
		return false
	}

	command := strings.Fields(commandWithArgs)[0]

	_, handlerPresent := repl.handlers[command]

	return handlerPresent
}

func (repl *REPL) handle(commandWithArgs string) bool {
	if commandWithArgs == "" {
		return false
	}

	commandWithArgsA := strings.Fields(commandWithArgs)

	handler, handlerPresent := repl.handlers[commandWithArgsA[0]]

	if handlerPresent {
		handler(commandWithArgsA[1:])	
	}

	return true
}

func (repl *REPL) Serve() {
	log.Println(repl.Description, "started")	

	readline.SetCompletionFunction(func (input, line string, start, end int) []string {
		var completions []string
		for command := range repl.handlers {
			if strings.HasPrefix(command, input) {
				completions = append(completions, command)
			}
		}
		return completions
	})

    // Ctrl+C handling, doesn't work properly
    intCh := make(chan os.Signal, 1)
    signal.Notify(intCh, os.Interrupt)
    go func(ch chan os.Signal) {
    	for s := range ch {
	    	repl.Signals <- s
	    	readline.CleanupAfterSignel()
    	}
	}(intCh)

	// This is generally what people expect in a modern Readline-based app
	readline.ParseAndBind("TAB: menu-complete")

	readline.SetCatchSignals(0)
	readline.ClearSignals()

L:	
	for {
		result := readline.Readline(&repl.Prompt)
	
		switch {
		case result == nil:
			fmt.Println()
			break L // exit loop
		case len(strings.TrimSpace(*result)) == 0:
			if repl.EmptyHandler != nil {
				repl.EmptyHandler()
			}
		case *result == "exit":
			break L // exit loop
		case repl.understands(*result):
			repl.handle(*result)
		default:
			fmt.Printf("Unknown command '%s', try 'help'\n", *result)
			continue
		}
		readline.AddHistory(*result) // Allow user to recall this line
	}
}
