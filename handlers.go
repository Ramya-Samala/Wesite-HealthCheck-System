package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const perPage = 10

func jsonResp(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

func errResp(w http.ResponseWriter, code int, msg string) {
	jsonResp(w, code, ErrBody{Message: msg})
}

// newID generates a random UUID v4
func newID() string {
	buf := make([]byte, 16)
	rand.Read(buf)
	// set version and variant bits for v4
	buf[6] = (buf[6] & 0x0f) | 0x40
	buf[8] = (buf[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		buf[0:4], buf[4:6], buf[6:8], buf[8:10], buf[10:16])
}

func doList(db *Store, w http.ResponseWriter, r *http.Request) {
	pg := 1
	if raw := r.URL.Query().Get("page"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 1 {
			errResp(w, 400, "page must be a positive integer")
			return
		}
		pg = n
	}

	all := db.All()
	start := (pg - 1) * perPage
	if start > len(all) {
		start = len(all)
	}
	end := start + perPage
	if end > len(all) {
		end = len(all)
	}

	page := all[start:end]
	if page == nil {
		page = []HealthCheck{}
	}

	jsonResp(w, 200, PagedResponse{
		Items: page,
		Page:  pg,
		Total: len(all),
		Size:  perPage,
	})
}

func doGet(db *Store, w http.ResponseWriter, id string) {
	hc, ok := db.Find(id)
	if !ok {
		errResp(w, 404, "health check not found")
		return
	}
	jsonResp(w, 200, hc)
}

func doCreate(db *Store, w http.ResponseWriter, r *http.Request) {
	var body CreateReq
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		errResp(w, 400, "could not parse request body")
		return
	}

	ep := strings.TrimSpace(body.Endpoint)
	if ep == "" {
		errResp(w, 400, "endpoint cannot be blank")
		return
	}

	// make sure it's a proper http/https URL
	u, err := url.ParseRequestURI(ep)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		errResp(w, 400, "endpoint is not a valid URL")
		return
	}

	if db.HasEndpoint(ep) {
		errResp(w, 409, "a check for this endpoint already exists")
		return
	}

	hc := &HealthCheck{
		ID:       newID(),
		Endpoint: ep,
	}
	if err := db.Put(hc); err != nil {
		errResp(w, 500, "failed to persist health check")
		return
	}
	jsonResp(w, 201, hc)
}

func doTry(db *Store, w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPost {
		errResp(w, 405, "method not allowed")
		return
	}

	hc, ok := db.Find(id)
	if !ok {
		errResp(w, 404, "health check not found")
		return
	}

	// default timeout is 10s but caller can override via query string
	tout := 10 * time.Second
	if raw := r.URL.Query().Get("timeout"); raw != "" {
		d, err := time.ParseDuration(raw)
		if err != nil {
			errResp(w, 400, "invalid timeout value")
			return
		}
		tout = d
	}

	RunCheck(&hc, tout)
	if err := db.Update(&hc); err != nil {
		errResp(w, 500, "failed saving check result")
		return
	}
	jsonResp(w, 200, hc)
}

func doDelete(db *Store, w http.ResponseWriter, id string) {
	if err := db.Remove(id); err != nil {
		errResp(w, 404, "health check not found")
		return
	}
	w.WriteHeader(204)
}

// HealthHandler handles all /api/health/checks routes and figures out
// what to do based on the URL path and HTTP method
func HealthHandler(db *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// trim the base path to get whatever comes after /api/health/checks
		tail := strings.TrimPrefix(r.URL.Path, "/api/health/checks")
		tail = strings.TrimRight(tail, "/")

		if tail == "" {
			routeCollection(db, w, r)
			return
		}

		if strings.HasSuffix(tail, "/try") {
			id := strings.TrimSuffix(strings.TrimPrefix(tail, "/"), "/try")
			doTry(db, w, r, id)
			return
		}

		id := strings.TrimPrefix(tail, "/")
		if strings.Contains(id, "/") {
			errResp(w, 404, "not found")
			return
		}
		routeSingle(db, w, r, id)
	}
}

func routeCollection(db *Store, w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		doList(db, w, r)
	} else if r.Method == http.MethodPost {
		doCreate(db, w, r)
	} else {
		errResp(w, 405, "method not allowed")
	}
}

func routeSingle(db *Store, w http.ResponseWriter, r *http.Request, id string) {
	switch r.Method {
	case http.MethodGet:
		doGet(db, w, id)
	case http.MethodDelete:
		doDelete(db, w, id)
	default:
		errResp(w, 405, "method not allowed")
	}
}
