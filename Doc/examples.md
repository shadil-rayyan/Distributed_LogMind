# Examples

These examples are small, dependency-light client programs that show how different users can talk to LogMind.

## Available Samples

- [curl example](examples/curl/send_logs.sh)
- [Go example](examples/go/main.go)
- [Python example](examples/python/send_logs.py)
- [Node.js example](examples/node/send_logs.mjs)

## How To Run Them

From the repository root, start LogMind first:

```bash
docker compose up -d
```

Then run whichever client you want:

```bash
sh Doc/examples/curl/send_logs.sh
go run Doc/examples/go/main.go
python3 Doc/examples/python/send_logs.py
node Doc/examples/node/send_logs.mjs
```

If your LogMind instance is not on `http://localhost:8080`, set `BASE_URL` before running the curl example.
