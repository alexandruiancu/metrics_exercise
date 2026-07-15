# Containerization Strategy

Some software components are to be deployed as contains.

## 📁 Directory Structure

```
.
├── composer.yaml # Docker Compose configuration (third party observability software)
└── .dockerignore  #Files to exclude from Docker builds

# Individual module Dockerfiles
bldrec/Dockerfile
loader/Dockerfile
```

## 🚀 Quick Start

### Local Testing (Docker Compose)

```bash

# View logs
docker compose -f composer.yaml logs -f

# Stop services
docker compose -f composer.yaml down
```
---

**Last Updated**: 2026-07-14
**Version**: 1.0.0
