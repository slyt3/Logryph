# Vouch: Cloud Deployment & Production Ops

This document is **optional**. Read it only if you plan to run Vouch in Docker/Kubernetes or need production operations guidance.

Vouch is designed to be cloud-native, lightweight, and extremely easy to deploy in production environments.

## Containerization (Docker)

The recommended way to run Vouch in the cloud is via Docker.

```dockerfile
# Build Stage
FROM golang:1.22-alpine AS builder
RUN apk add --no-cache gcc musl-dev
WORKDIR /app
COPY . .
RUN go build -o vouch main.go

# Production Stage
FROM alpine:latest
RUN apk add --no-cache ca-certificates
WORKDIR /root/
COPY --from=builder /app/vouch .
COPY --from=builder /app/vouch-policy.yaml .

EXPOSE 9999 9998
CMD ["./vouch"]
```

## CI/CD Pipeline

Vouch uses GitHub Actions for automated quality assurance.

*   **Workflow**: `.github/workflows/ci.yml`
*   **Triggers**: Every push and PR.
*   **Checks**: `go test`, `go vet`, and cross-platform build verification.

## Production Operations

### 1. Persistence & Volumes
Vouch stores its ledger in `vouch.db`. In a cloud environment (Kubernetes/ECS), this **must** be mounted to a persistent volume (EBS/PVC) to ensure logs are not lost on container restart.

### 2. Monitoring
Vouch provides multiple monitoring interfaces:

**Prometheus Metrics** (Recommended):
- Endpoint: `http://localhost:9998/metrics`
- Metrics:
  - `vouch_ledger_events_processed_total`: Total events written to ledger
  - `vouch_ledger_events_dropped_total`: Events dropped due to backpressure
  - `vouch_pool_event_hits_total`: Event pool cache efficiency
  - `vouch_engine_active_tasks_total`: Currently active causal tasks

**Structured Logs**:
*   Vouch emits structured logs to `stdout`
*   **Backpressure**: Monitor for `[BACKPRESSURE]` messages. If frequent, increase persistence IOPS
*   **Health**: Monitor for `[CRITICAL]` ledger failure logs

### 3. Key Management
The `.vouch_key` is the root of trust.
*   **Security**: Do not commit this key to version control.
*   **Rotation**: Use the `vouch-cli rekey` command periodically to rotate the Ed25519 pair.

## Architecture Patterns

| Pattern | Implementation |
| :--- | :--- |
| **Sidecar** | Run Vouch as a sidecar container to your AI Agent service. |
| **Gateway** | Run Vouch as a centralized gateway for multiple agents. |
| **Offline** | Run Vouch locally for development and periodic audit uploads. |
