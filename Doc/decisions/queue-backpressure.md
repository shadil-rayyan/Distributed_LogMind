# Queue Backpressure

## Decision

The ingestion handler blocks until it can enqueue a log, instead of dropping the log immediately when the channel is full.

## Why

The previous default-case send turned overload into silent data loss. That made the queue look healthy while logs were being discarded. For a log monitoring system, preserving data under pressure matters more than returning an immediate success/failure boundary.

## What Changed

- Removed the non-blocking send fallback.
- Let the request wait for queue capacity or cancellation.
- Keep queue depth metrics aligned with actual queue usage.

## Tradeoff

Requests can take longer under sustained load, and callers need a timeout or retry policy. That is the correct pressure release valve here: upstream clients should absorb the load-shedding decision, not the ingestion layer silently.
