package protocol

import (
	"net/http"
	"log"
	"fmt"
	"html"
)

type Client struct {
	Address string
	PlayerID string
}

func NewClient(address string, id string) *Client {
	return &Client{
		Address:  address,
		PlayerID: id,
	}
}
func (client *Client) Serve() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("GET %s", html.EscapeString(r.URL.Path))
		fmt.Fprintf(w, "Hello from %s\n", html.EscapeString(client.PlayerID))
	})

	log.Fatal(http.ListenAndServe(client.Address, nil))
}