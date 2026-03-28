package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go_project/config"
	"go_project/logger"
	"go_project/validator"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"
)

var (
	contracts    map[string]*config.Contract
	mu           sync.RWMutex
	violationLog []ViolationRecord
	vMu          sync.RWMutex
)

type ViolationRecord struct {
	Timestamp  string               `json:"timestamp"`
	Endpoint   string               `json:"endpoint"`
	Method     string               `json:"method"`
	Direction  string               `json:"direction"`
	Violations []validator.Violation `json:"violations"`
}

func corsHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func storeViolations(endpoint, method, direction string, violations []validator.Violation) {
	record := ViolationRecord{
		Timestamp:  time.Now().Format(time.RFC3339),
		Endpoint:   endpoint,
		Method:     method,
		Direction:  direction,
		Violations: violations,
	}
	vMu.Lock()
	violationLog = append(violationLog, record)
	vMu.Unlock()
}

func main() {
	if err := logger.Init(); err != nil {
		fmt.Println("Failed to initialize logger:", err)
		return
	}
	defer logger.Log.Sync()

	contracts = make(map[string]*config.Contract)
	violationLog = []ViolationRecord{}

	if c, err := config.LoadContract("contracts/user-service.json"); err == nil {
		key := c.Method + " " + c.Endpoint
		contracts[key] = c
		fmt.Println("Contract loaded from file for endpoint:", key)
	} else {
		fmt.Println("No contract file found — POST to /contract to load one")
	}

	// POST /contract — register a new contract
	http.HandleFunc("/contract", func(w http.ResponseWriter, r *http.Request) {
		corsHeaders(w)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read body", http.StatusBadRequest)
			return
		}
		var c config.Contract
		if err := json.Unmarshal(body, &c); err != nil {
			http.Error(w, "invalid JSON contract: "+err.Error(), http.StatusBadRequest)
			return
		}
		if c.Endpoint == "" || c.Method == "" || c.Target == "" || len(c.Request) == 0 {
			http.Error(w, "contract must have endpoint, method, target, and request fields", http.StatusBadRequest)
			return
		}
		key := c.Method + " " + c.Endpoint
		mu.Lock()
		contracts[key] = &c
		mu.Unlock()
		fmt.Printf("Contract added: %s\n", key)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message":  "contract added",
			"endpoint": key,
		})
	})

	// GET /contracts — list all loaded contracts
	http.HandleFunc("/contracts", func(w http.ResponseWriter, r *http.Request) {
		corsHeaders(w)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		mu.RLock()
		result := make([]config.Contract, 0, len(contracts))
		for _, c := range contracts {
			result = append(result, *c)
		}
		mu.RUnlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	})

	// GET /violations — return in-memory violation history
	http.HandleFunc("/violations", func(w http.ResponseWriter, r *http.Request) {
		corsHeaders(w)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		vMu.RLock()
		defer vMu.RUnlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(violationLog)
	})

	// / — main proxy handler
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		corsHeaders(w)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		reqBody, _ := io.ReadAll(r.Body)
		r.Body = io.NopCloser(bytes.NewBuffer(reqBody))

		fmt.Println("=== INCOMING REQUEST ===")
		fmt.Printf("Endpoint: %s %s\n", r.Method, r.URL.Path)
		fmt.Printf("Body: %s\n", string(reqBody))

		key := r.Method + " " + r.URL.Path
		mu.RLock()
		c := contracts[key]
		mu.RUnlock()

		if c == nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": "no contract found for this endpoint",
			})
			return
		}

		reqViolations := validator.ValidateJSON(reqBody, c.Request, "REQUEST", c)
		if len(reqViolations) > 0 {
			fmt.Println("REQUEST blocked — contract violations found")
			fmt.Println("========================")
			storeViolations(c.Endpoint, c.Method, "REQUEST", reqViolations)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":      "request violates contract",
				"violations": reqViolations,
			})
			return
		}

		recorder := newResponseRecorder()
		target, err := url.Parse(c.Target)
		if err != nil {
			http.Error(w, "invalid contract target: "+err.Error(), http.StatusInternalServerError)
			return
		}
		httputil.NewSingleHostReverseProxy(target).ServeHTTP(recorder, r)

		fmt.Println("=== OUTGOING RESPONSE ===")
		fmt.Printf("Status: %d\n", recorder.status)
		fmt.Printf("Body: %s\n", recorder.body.String())

		respViolations := validator.ValidateJSON(recorder.body.Bytes(), c.Response, "RESPONSE", c)
		if len(respViolations) > 0 {
			fmt.Println("RESPONSE blocked — contract violations found")
			fmt.Println("========================")
			storeViolations(c.Endpoint, c.Method, "RESPONSE", respViolations)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadGateway)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":      "response from upstream violates contract",
				"violations": respViolations,
			})
			return
		}

		for k, v := range recorder.header {
			for _, vv := range v {
				w.Header().Add(k, vv)
			}
		}
		w.WriteHeader(recorder.status)
		w.Write(recorder.body.Bytes())

		fmt.Println("========================")
	})

	fmt.Println("Proxy running on :8080")
	http.ListenAndServe(":8080", nil)
}

type ResponseRecorder struct {
	header http.Header
	status int
	body   *bytes.Buffer
}

func newResponseRecorder() *ResponseRecorder {
	return &ResponseRecorder{
		header: make(http.Header),
		body:   &bytes.Buffer{},
		status: http.StatusOK,
	}
}

func (r *ResponseRecorder) Header() http.Header {
	return r.header
}

func (r *ResponseRecorder) WriteHeader(status int) {
	r.status = status
}

func (r *ResponseRecorder) Write(b []byte) (int, error) {
	return r.body.Write(b)
}
