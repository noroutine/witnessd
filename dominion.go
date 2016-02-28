package main

import (
	"fmt"
	"strings"

	"github.com/fiorix/go-readline"
)

const version string = "0.0.1"
const description = "Dominion " + version

var phoneticAlphabet = map[string][]string{
	"e": {"exit"},
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

func main() {
	prompt := description + "> "

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
		case *result != "": // Ignore blank lines
			fmt.Println(*result)
			readline.AddHistory(*result) // Allow user to recall this line
		}
	}
}
