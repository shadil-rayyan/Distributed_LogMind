# Config Validation

## Decision

Configuration is validated at startup and the process exits early if a required value is invalid.

## Why

This service has a few values that are not safe to leave open-ended. A zero or negative sliding window can panic the metrics engine, and zero worker counts or batch sizes make the pipeline behave unpredictably. Silent fallback would hide deployment mistakes until the system is already in trouble.

## What Changed

- Added explicit validation for worker count, queue size, sliding window, threshold, batch size, and batch timeout.
- Fail fast during bootstrap instead of letting bad values leak into runtime behavior.

## Tradeoff

Misconfiguration now stops the process instead of being tolerated. That is deliberate: operational errors should be loud at startup, not buried in production traffic.
