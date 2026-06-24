# Getting Started

This page is the shortest path from a fresh clone to a working LogMind instance.

## Run The Service

```bash
docker compose up -d
```

Or run it locally:

```bash
go run ./cmd/logmind
```

## Send A Log

```bash
curl -X POST http://localhost:8080/logs \
  -H "Content-Type: application/json" \
  -d '{"service":"payment-api","level":"error","message":"Database connection failed"}'
```

The service accepts `service`, `level`, and `message`. It assigns the timestamp itself.

## View Incidents

```bash
curl http://localhost:8080/incidents
```

If there are enough recent errors for a service, the response includes an active incident entry.

## Check Service Health

```bash
curl http://localhost:8080/healthz
curl http://localhost:8080/readyz
curl http://localhost:8080/metrics
```

## Example Workflows

### Curl

Send three errors to trigger the default threshold:

```bash
curl -X POST http://localhost:8080/logs \
  -H "Content-Type: application/json" \
  -d '{"service":"payment-api","level":"error","message":"Database connection failed"}'
curl -X POST http://localhost:8080/logs \
  -H "Content-Type: application/json" \
  -d '{"service":"payment-api","level":"error","message":"Timeout while querying users"}'
curl -X POST http://localhost:8080/logs \
  -H "Content-Type: application/json" \
  -d '{"service":"payment-api","level":"error","message":"Retry budget exhausted"}'
curl http://localhost:8080/incidents
```

### Go

```go
req, _ := http.NewRequest(http.MethodPost, baseURL+"/logs", bytes.NewReader(body))
req.Header.Set("Content-Type", "application/json")
resp, err := client.Do(req)
```

Runnable file: [Go example](examples/go/main.go)

### Python

```python
request = urllib.request.Request(
    f"{base_url}/logs",
    data=json.dumps(payload).encode("utf-8"),
    headers={"Content-Type": "application/json"},
    method="POST",
)
```

Runnable file: [Python example](examples/python/send_logs.py)

### Node.js

```javascript
await fetch(`${baseUrl}/logs`, {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify(payload),
});
```

Runnable file: [Node.js example](examples/node/send_logs.mjs)
