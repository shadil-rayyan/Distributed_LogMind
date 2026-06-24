# Metrics Cardinality

## Decision

`logmind_logs_ingested_total` is now a single counter without user-controlled labels.

## Why

Prometheus labels are part of the time-series identity. Using raw service names as labels turns every new service string into a new series, which can grow without bound if the input is noisy or adversarial. That creates avoidable memory pressure in the monitoring layer itself.

## What Changed

- Removed the `service` and `level` labels from the ingestion counter.
- Kept the metric name stable so the dashboard and scrape config still work.
- Preserved the useful aggregate signal: total ingestion rate.

## Tradeoff

We lose per-service Prometheus breakdown for this counter, but the system still exposes service-specific incident information through the API. That is a better place for high-cardinality data than Prometheus.
