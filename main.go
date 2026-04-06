package main

import (
	"flag"
	"log"
	"net/http"
	"time"
)

func main() {
	addr := flag.String("bind", "127.0.0.1:8080", "listen address")
	freq := flag.String("checkfrequency", "30s", "background check interval")
	flag.Parse()

	interval, err := time.ParseDuration(*freq)
	if err != nil {
		log.Fatalf("invalid checkfrequency: %v", err)
	}

	db := NewStore()
	if err := db.Load(); err != nil {
		log.Fatalf("loading data: %v", err)
	}

	log.Printf("listening on %s", *addr)
	err = http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal(err)
	}
}
