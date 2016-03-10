package protocol

import (
	"net/http"
	"log"
	"os/exec"
	"fmt"
	"html"
)

type Client struct {
	Address string
	PlayerId string
}


func (self *Client) Serve() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		out, err := exec.Command("/bin/hostname").Output()
		if err != nil {
			log.Fatal(err)
		} else {
			fmt.Fprintf(w, "Hello from player, %q from %s\n", html.EscapeString(self.PlayerId), out)
		}			
	})

	log.Fatal(http.ListenAndServe(self.Address, nil))
}