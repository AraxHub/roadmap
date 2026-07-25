#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
cd "$ROOT_DIR"

kubectl apply -k deployment/k8s/overlays/k3s

kubectl -n roadmap rollout status deploy/roadmap-api --timeout=180s
kubectl -n roadmap rollout status deploy/roadmap-frontend --timeout=180s

echo "Applied. Check:"
echo "  kubectl -n roadmap get pods,svc,ing"
echo "  kubectl -n roadmap logs -f deploy/roadmap-api"
