package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

type KYCRequest struct {
	CustomerID     string  `json:"customerId"`
	FullName       string  `json:"fullName"`
	DateOfBirth    string  `json:"dateOfBirth"`
	DocumentType   string  `json:"documentType"`
	DocumentNumber string  `json:"documentNumber"`
	Address        Address `json:"address"`
}

type Address struct {
	Street  string `json:"street"`
	City    string `json:"city"`
	Pincode string `json:"pincode"`
}

type KYCResponse struct {
	CustomerID     string  `json:"customerId"`
	VerificationID string  `json:"verificationId"`
	Status         string  `json:"status"`
	RiskScore      float64 `json:"riskScore"`
	VerifiedAt     string  `json:"verifiedAt"`
}

type CreateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type CreateUserResponse struct {
	UserID    string `json:"userId"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	CreatedAt string `json:"createdAt"`
}

type AccountStatusRequest struct {
	AccountID string `json:"accountId"`
}

type AccountStatusResponse struct {
	AccountID string `json:"accountId"`
	Status    string `json:"status"`
	Balance   float64 `json:"balance"`
	Currency  string `json:"currency"`
	UpdatedAt string `json:"updatedAt"`
}

// In-memory identity registry: documentNumber -> fullName
var registry = map[string]string{
	"DL1234567":  "Amit Sharma",
	"PAN9876543": "Priya Mehta",
	"PASS112233": "Rahul Verma",
}

func main() {
	http.HandleFunc("/api/kyc/verify", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req KYCRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		status := "rejected"
		riskScore := 85.0 + rand.Float64()*15 // high risk 85-100

		registeredName, exists := registry[req.DocumentNumber]
		if exists && registeredName == req.FullName {
			status = "verified"
			riskScore = rand.Float64() * 30 // low risk 0-30
		} else if exists {
			status = "pending"
			riskScore = 40 + rand.Float64()*30 // medium risk 40-70
		}

		resp := KYCResponse{
			CustomerID:     req.CustomerID,
			VerificationID: fmt.Sprintf("VER-%d", time.Now().UnixNano()),
			Status:         status,
			RiskScore:      riskScore,
			VerifiedAt:     time.Now().Format(time.RFC3339),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	http.HandleFunc("/api/user/create", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req CreateUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		resp := CreateUserResponse{
			UserID:    fmt.Sprintf("USR-%d", time.Now().UnixNano()),
			Name:      req.Name,
			Email:     req.Email,
			Role:      req.Role,
			CreatedAt: time.Now().Format(time.RFC3339),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	http.HandleFunc("/api/account/status", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req AccountStatusRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		statuses := []string{"active", "suspended", "pending"}
		resp := AccountStatusResponse{
			AccountID: req.AccountID,
			Status:    statuses[rand.Intn(len(statuses))],
			Balance:   float64(rand.Intn(100000)) / 100,
			Currency:  "INR",
			UpdatedAt: time.Now().Format(time.RFC3339),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	fmt.Println("Identity Registry Service running on :8002")
	http.ListenAndServe(":8002", nil)
}
