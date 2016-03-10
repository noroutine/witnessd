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
		"", 
		"> ",
		make(chan os.Signal, 1),
		nil,
		map[string]HandlerFunc{},
	}
}

func (self *REPL) Register(command string, handler HandlerFunc) {
	self.handlers[command] = handler 
}

func (self *REPL) GetKnownCommands() []string {
	commands := make([]string, len(self.handlers))
	i := 0
	for command := range self.handlers {
		commands[i] = command
		i++
	}	
	return commands
}

func (self *REPL) understands(commandWithArgs string) bool {
	if commandWithArgs == "" {
		return false
	}

	command := strings.Fields(commandWithArgs)[0]

	_, handlerPresent := self.handlers[command]

	return handlerPresent
}

func (self *REPL) handle(commandWithArgs string) bool {
	if commandWithArgs == "" {
		return false
	}

	commandWithArgsA := strings.Fields(commandWithArgs)

	handler, handlerPresent := self.handlers[commandWithArgsA[0]]

	if handlerPresent {
		handler(commandWithArgsA[1:])	
	}

	return true
}

func (self *REPL) Serve() {
	log.Println(self.Description, "started")	

	readline.SetCompletionFunction(func (input, line string, start, end int) []string {
		var completions []string
		for command := range self.handlers {
			if strings.HasPrefix(command, input) {
				completions = append(completions, command)
			}
		}
		return completions
	})

    // Ctrl+C handling, doesn't work properly
    int_ch := make(chan os.Signal, 1)
    signal.Notify(int_ch, os.Interrupt)
    go func(ch chan os.Signal) {
    	for s := range ch {
	    	self.Signals <- s
	    	readline.CleanupAfterSignel()
    	}
	}(int_ch)

	// This is generally what people expect in a modern Readline-based app
	readline.ParseAndBind("TAB: menu-complete")

	readline.SetCatchSignals(0)
	readline.ClearSignals()

L:	
	for {
		result := readline.Readline(&self.Prompt)
	
		switch {
		case result == nil:
			fmt.Println()
			break L // exit loop
		case len(strings.TrimSpace(*result)) == 0:
			if self.EmptyHandler != nil {
				self.EmptyHandler()
			}
		case *result == "exit":
			break L // exit loop
		case self.understands(*result):
			self.handle(*result)
		default:
			fmt.Printf("Unknown command '%s', try 'help'\n", *result)
			continue
		}
		readline.AddHistory(*result) // Allow user to recall this line
	}
}
