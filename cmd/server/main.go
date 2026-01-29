package main

import (
	"flag"
	"log"

	"tic-tac-toe-go/internal/server"
)

func main() {
	addr := flag.String("addr", ":9000", "listen address")
	flag.Parse()

	srv := &server.Server{Addr: *addr}
	log.Printf("listening on %s", *addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
