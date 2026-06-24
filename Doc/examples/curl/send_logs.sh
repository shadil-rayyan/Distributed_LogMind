#!/usr/bin/env sh
set -eu

BASE_URL="${BASE_URL:-http://localhost:8080}"

curl -sS -X POST "$BASE_URL/logs" \
  -H "Content-Type: application/json" \
  -d '{"service":"payment-api","level":"error","message":"Database connection failed"}'
printf '\n'

curl -sS -X POST "$BASE_URL/logs" \
  -H "Content-Type: application/json" \
  -d '{"service":"payment-api","level":"error","message":"Timeout while querying users"}'
printf '\n'

curl -sS -X POST "$BASE_URL/logs" \
  -H "Content-Type: application/json" \
  -d '{"service":"payment-api","level":"error","message":"Retry budget exhausted"}'
printf '\n'

curl -sS "$BASE_URL/incidents"
printf '\n'
