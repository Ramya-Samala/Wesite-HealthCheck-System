package main

import (
	"fmt"
	"net/http"
	"time"
)

func RunCheck(hc *HealthCheck, timeout time.Duration) {
	cl := &http.Client{Timeout: timeout}

	t0 := time.Now()
	resp, err := cl.Get(hc.Endpoint)
	took := time.Since(t0)

	hc.Checked = time.Now().Unix()
	hc.Duration = fmt.Sprintf("%dms", took.Milliseconds())

	if err != nil {
		zero := int32(0)
		hc.Status = "Error"
		hc.Code = &zero
		hc.Error = err.Error()
		return
	}

	// close the body even if we don't read it, otherwise connections leak
	resp.Body.Close()

	code := int32(resp.StatusCode)
	hc.Code = &code
	hc.Status = http.StatusText(resp.StatusCode)
	hc.Error = "" // clear any previous error
}

// RunAll fires off every check at the same time and waits for all to finish
func RunAll(db *Store, timeout time.Duration) {
	checks := db.All()
	if len(checks) == 0 {
		return
	}

	// buffered channel so goroutines don't block waiting to send
	ch := make(chan HealthCheck, len(checks))

	for _, c := range checks {
		go func(hc HealthCheck) {
			RunCheck(&hc, timeout)
			ch <- hc
		}(c)
	}

	// collect all results and save them
	for i := 0; i < len(checks); i++ {
		done := <-ch
		db.Update(&done) //nolint: errcheck
	}
}
