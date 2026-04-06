package main

import (
	"log"
	"time"
)

func StartScheduler(db *Store, freq time.Duration, quit chan struct{}) {
	t := time.NewTicker(freq)
	go func() {
		for {
			select {
			case <-t.C:
				log.Println("running checks...")
				RunAll(db, 10*time.Second)
			case <-quit:
				t.Stop()
				return
			}
		}
	}()
}
