# Container Hardening

## Decision

The production image runs as a non-root user, with a read-only root filesystem, explicit memory and CPU limits, and a writable data volume only for SQLite.

## Why

The earlier setup described hardening that the runtime did not actually enforce. Running as root, relying on a writable root filesystem, or depending only on Swarm-style `deploy` limits creates a false sense of safety. The container should enforce the same constraints the documentation promises.

## What Changed

- Added a dedicated non-root user in the image.
- Marked the root filesystem read-only in production compose.
- Added a `/tmp` tmpfs for runtime scratch space.
- Added direct `mem_limit` and `cpus` constraints for normal Docker Compose usage.

## Tradeoff

The image is slightly more opinionated, but the deployment story is now honest and repeatable. The app still writes to SQLite through the mounted data volume, which is the only mutable path it needs.
