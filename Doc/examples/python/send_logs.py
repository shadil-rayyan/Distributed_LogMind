#!/usr/bin/env python3
import json
import os
import sys
import urllib.error
import urllib.request


def send_log(base_url, payload):
    request = urllib.request.Request(
        f"{base_url}/logs",
        data=json.dumps(payload).encode("utf-8"),
        headers={"Content-Type": "application/json"},
        method="POST",
    )

    with urllib.request.urlopen(request, timeout=5) as response:
        body = response.read().decode("utf-8")
        print(f"POST /logs -> {response.status} {body}")


def fetch_incidents(base_url):
    request = urllib.request.Request(f"{base_url}/incidents")
    with urllib.request.urlopen(request, timeout=5) as response:
        body = response.read().decode("utf-8")
        print(f"GET /incidents -> {response.status} {body}")


def main():
    base_url = os.environ.get("BASE_URL", "http://localhost:8080")
    logs = [
        {"service": "payment-api", "level": "error", "message": "Database connection failed"},
        {"service": "payment-api", "level": "error", "message": "Timeout while querying users"},
        {"service": "payment-api", "level": "error", "message": "Retry budget exhausted"},
    ]

    try:
        for payload in logs:
            send_log(base_url, payload)
        fetch_incidents(base_url)
    except urllib.error.URLError as exc:
        print(f"request failed: {exc}", file=sys.stderr)
        raise SystemExit(1)


if __name__ == "__main__":
    main()
