# Docker Containerization Strategy for metrics_exercise

Some software components are to be deployed as contains.

## 📁 Directory Structure

```
.
├── composer.yaml # Docker Compose configuration (third party observability software)
└── .dockerignore  #Files to exclude from Docker builds

# Individual module Dockerfiles
bldrec/Dockerfile
loader/Dockerfile
prometheus/Dockerfile
otelcol/Dockerfile
```

## 🚀 Quick Start

### Local Testing (Docker Compose)

```bash

# View logs
docker compose -f composer.yaml logs -f

# Stop services
docker compose -f composer.yaml down
```

## 📊 Service Architecture

```
┌─────────────┐      ZMQ      ┌─────────────┐
│   bldrec    │ ◄──────────► │   loader    │
│ (Producer)  │   (Router)   │ (Consumer)  │
└─────────────┘              └─────────────┘
       │                          │
       └──────────┬───────────────┘
                  │
                  ▼
           ┌─────────────┐
           │   driver    │
           │ (Orchestrator)│
           └─────────────┘
                  │
         ┌────────┴────────┐
         │                 │
         ▼                 ▼
   ┌──────────┐     ┌──────────┐
   │Prometheus│     │OTel Col. │
   │:9090     │     │:4317/8889│
   └──────────┘     └──────────┘
```
---

**Last Updated**: 2026-07-16
**Version**: 1.1.0
