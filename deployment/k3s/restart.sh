#!/usr/bin/env bash
set -euo pipefail

# After rebuilding :local images, force pods to pick them up.
kubectl -n roadmap rollout restart deploy/roadmap-api deploy/roadmap-frontend
kubectl -n roadmap rollout status deploy/roadmap-api --timeout=180s
kubectl -n roadmap rollout status deploy/roadmap-frontend --timeout=180s
