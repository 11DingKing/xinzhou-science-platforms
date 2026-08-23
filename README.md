# Xinzhou Innovation Platform Governance

Xinzhou Innovation Platform Governance is a production-oriented Go backend for the city's “十五五” science-and-technology platform program. It coordinates platform applications, expert review, milestone acceptance, staged funding, annual reports, audit events, and recoverable background work. The quality-governance flows retained in the foundation provide reusable evidence and remediation patterns for platform supervision; the product deliberately does not implement e-commerce, inventory, booking, voting, or dashboard-only behavior.

Run locally:

```text
GOTOOLCHAIN=local go run ./cmd/server
```

Health endpoints: `/healthz` and `/readyz`. Innovation platform APIs live under `/v1/platforms` and use the same revocable session and role-aware authorization as the existing governance APIs.
