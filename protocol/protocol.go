package protocol

import (
	"net/http"
	"log"
	"fmt"
	"html"
)

type Client struct {
	Address string
	PlayerId string
}

func (self *Client) Serve() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("GET %s", html.EscapeString(r.URL.Path))
		fmt.Fprintf(w, "Hello from %s\n", html.EscapeString(self.PlayerId))
	})

	log.Fatal(http.ListenAndServe(self.Address, nil))
}