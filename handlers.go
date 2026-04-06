package main

import (
	"encoding/json"
	"net/http"
	"strconv"
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
