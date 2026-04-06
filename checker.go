package main

import (
	"net/http"
	"time"
)

func RunCheck(hc *HealthCheck, timeout time.Duration) {
	cl := &http.Client{Timeout: timeout}
	resp, err := cl.Get(hc.Endpoint)
	if err != nil {
		hc.Status = "Error"
		return
	}
	resp.Body.Close()
	hc.Status = http.StatusText(resp.StatusCode)
}
