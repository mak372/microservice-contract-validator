package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
)

type Contract struct {
	Endpoint string            `json:"endpoint"`
	Method   string            `json:"method"`
	Request  map[string]string `json:"request"`
	Response map[string]string `json:"response"`
}

var (
	currentContract *Contract
	mu              sync.RWMutex
)

func main() {
	// POST /contract — load a JSON contract into serviceA
	http.HandleFunc("/contract", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read body", http.StatusBadRequest)
			return
		}
		var c Contract
		if err := json.Unmarshal(body, &c); err != nil {
			http.Error(w, "invalid JSON contract: "+err.Error(), http.StatusBadRequest)
			return
		}
		if c.Endpoint == "" || c.Method == "" || len(c.Request) == 0 {
			http.Error(w, "contract must have endpoint, method, and request fields", http.StatusBadRequest)
			return
		}
		mu.Lock()
		currentContract = &c
		mu.Unlock()
		fmt.Printf("Contract loaded: %s %s | fields: %v\n", c.Method, c.Endpoint, fieldNames(c.Request))
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message":  "contract loaded",
			"endpoint": c.Endpoint,
			"method":   c.Method,
			"fields":   c.Request,
		})
	})

	// GET /schema — show the currently loaded contract
	http.HandleFunc("/schema", func(w http.ResponseWriter, r *http.Request) {
		mu.RLock()
		c := currentContract
		mu.RUnlock()
		if c == nil {
			http.Error(w, "no contract loaded — POST to /contract first", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(c)
	})

	// POST /send — send dynamic data to the proxy based on the loaded contract
	http.HandleFunc("/send", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		mu.RLock()
		c := currentContract
		mu.RUnlock()
		if c == nil {
			http.Error(w, "no contract loaded — POST to /contract first", http.StatusBadRequest)
			return
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read body", http.StatusBadRequest)
			return
		}
		var data map[string]interface{}
		if err := json.Unmarshal(body, &data); err != nil {
			http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		// Check all contract fields are present in the payload
		var missing []string
		for field := range c.Request {
			if _, ok := data[field]; !ok {
				missing = append(missing, field)
			}
		}
		if len(missing) > 0 {
			http.Error(w, "missing contract fields: "+strings.Join(missing, ", "), http.StatusBadRequest)
			return
		}
		proxyURL := "http://localhost:8080" + c.Endpoint
		resp, err := http.Post(proxyURL, "application/json", bytes.NewReader(body))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()
		fmt.Fprintf(w, "ServiceB responded with status: %d", resp.StatusCode)
	})

	fmt.Println("ServiceA running on :8001")
	http.ListenAndServe(":8001", nil)
}

func fieldNames(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
