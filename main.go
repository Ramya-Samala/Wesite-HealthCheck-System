package main

import (
	"flag"
	"log"
	"net/http"
	"time"
)

func main() {
	// read all config from command line
	addr := flag.String("bind", "127.0.0.1:8080", "listen address")
	useSSL := flag.Bool("ssl", false, "enable TLS")
	cert := flag.String("sslcert", "", "TLS cert path")
	key := flag.String("sslkey", "", "TLS key path")
	freq := flag.String("checkfrequency", "30s", "background check interval")
	flag.Parse()

	interval, err := time.ParseDuration(*freq)
	if err != nil {
		log.Fatalf("invalid checkfrequency: %v", err)
	}

	// can't run ssl without both cert and key
	if *useSSL && (*cert == "" || *key == "") {
		log.Fatal("ssl requires both sslcert and sslkey")
	}

	// load existing checks from disk so we don't lose data on restart
	db := NewStore()
	if err := db.Load(); err != nil {
		log.Fatalf("loading data: %v", err)
	}

	// quit channel is used to stop the scheduler cleanly on shutdown
	quit := make(chan struct{})
	StartScheduler(db, interval, quit)

	mux := http.NewServeMux()
	// two routes needed — one with trailing slash, one without
	mux.HandleFunc("/api/health/checks", HealthHandler(db))
	mux.HandleFunc("/api/health/checks/", HealthHandler(db))

	log.Printf("listening on %s", *addr)
	if *useSSL {
		log.Printf("tls cert=%s key=%s", *cert, *key)
		err = http.ListenAndServeTLS(*addr, *cert, *key, mux)
	} else {
		err = http.ListenAndServe(*addr, mux)
	}
	close(quit)
	if err != nil {
		log.Fatal(err)
	}
}
