# Contract-Validating Proxy

A Go project that demonstrates a **contract-validating reverse proxy** sitting between two services. The proxy intercepts all traffic, validates the request and response payloads against a defined contract schema, and forwards everything transparently.

---

## How It Works

```
Service A  →  Proxy (:8080)  →  Service B (:8002)
  :8001         ↕ validate           ↕
             contract.json      responds
```

1. **Service A** sends a POST request with a JSON payload to the proxy
2. **Proxy** intercepts the request, reads the body, and validates it against the contract schema
3. **Proxy** forwards the request to **Service B**
4. **Service B** processes and responds
5. **Proxy** captures the response, validates it against the contract schema, and passes it back to A
6. **Service A** receives B's response

Nothing is blocked or modified — the proxy observes and reports violations only.

---

## Project Structure

```
go_project/
├── serviceA/
│   └── main.go          # Sends POST request to the proxy
├── serviceB/
│   └── main.go          # Receives and processes the request
├── proxy/
│   └── main.go          # Intercepts, validates, and forwards traffic
├── validator/
│   └── validator.go     # Validates JSON payloads against a schema
├── config/
│   └── config.go        # Loads the contract JSON file
├── contracts/
│   └── user-service.json  # Contract schema definition
└── go.mod
```

---

## Contract Schema

The contract is defined in `contracts/user-service.json`. It specifies the endpoint, HTTP method, and the expected fields + types for both the request and response.

```json
{
  "endpoint": "/api/user",
  "method": "POST",
  "request": {
    "user_id": "string",
    "email": "string",
    "age": "number"
  },
  "response": {
    "status": "string",
    "message": "string"
  }
}
```

Supported types: `string`, `number`, `boolean`, `object`, `array`, `null`

---

## Services

| Service   | Port  | Role                                      |
|-----------|-------|-------------------------------------------|
| Service A | 8001  | Sends `POST /api/user` to the proxy       |
| Proxy     | 8080  | Validates and forwards traffic            |
| Service B | 8002  | Receives request, returns JSON response   |

### Service A — Request Payload
```json
{
  "user_id": "123",
  "email": "test@test.com",
  "age": 25
}
```

### Service B — Response
```json
{
  "status": "success",
  "message": "user processed"
}
```

---

## Running the Project

Open three terminals and run each service:

```bash
# Terminal 1 — Service B (start first, proxy needs it running)
cd serviceB && go run main.go

# Terminal 2 — Proxy
cd proxy && go run main.go

# Terminal 3 — Service A
cd serviceA && go run main.go
```

Then trigger the flow by hitting Service A:

```bash
curl http://localhost:8001/send
```

---

## Example Proxy Output

```
Contract loaded for endpoint: /api/user
Proxy running on :8080

=== INCOMING REQUEST ===
Endpoint: POST /api/user
Body: {"user_id":"123","email":"test@test.com","age":25}
[REQUEST] Contract OK

=== OUTGOING RESPONSE ===
Status: 200
Body: {"status":"success","message":"user processed"}
[RESPONSE] Contract OK
========================
```

If a field is missing or has the wrong type, violations are printed:

```
[REQUEST] Contract VIOLATIONS FOUND:
  - Field: age   | Issue: wrong type    | Expected: number | Got: string
  - Field: email | Issue: missing field | Expected: string | Got: null
```

---

## Requirements

- Go 1.24+

---

## Tech Stack

- **Language:** Go
- **Proxy:** `net/http/httputil.ReverseProxy`
- **Validation:** Custom JSON schema validator
- **Config:** JSON-based contract files
