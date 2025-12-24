package main

import (
	"flag"
	"log"

	"linkedin-automation-poc/internal/mockserver"
)

func main() {
	addr := flag.String("addr", ":7777", "listen address")
	flag.Parse()

	server := mockserver.New()
	log.Printf("mock LinkedIn server listening on http://localhost%s", *addr)
	log.Fatal(server.ListenAndServe(*addr))
}
