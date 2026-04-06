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
		hc.Status = "Error"
		hc.Code = 0
		hc.Error = err.Error()
		return
	}

	// close the body even if we don't read it, otherwise connections leak
	resp.Body.Close()
	hc.Status = http.StatusText(resp.StatusCode)
	hc.Code = int32(resp.StatusCode)
	hc.Error = "" // clear any previous error
}
