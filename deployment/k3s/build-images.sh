#!/usr/bin/env bash
set -euo pipefail

# Build local images on the VPS, then import them into k3s containerd.
# (docker build alone is not enough — k3s uses its own containerd.)

ROOT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
cd "$ROOT_DIR"

echo "Building API image..."
docker build -t roadmap-api:local -f deployment/k8s/images/api/Dockerfile .

echo "Building frontend image..."
docker build -t roadmap-frontend:local -f deployment/k8s/images/frontend/Dockerfile .

if command -v k3s >/dev/null 2>&1; then
  echo "Importing images into k3s containerd..."
  docker save roadmap-api:local roadmap-frontend:local | sudo k3s ctr images import -
else
  echo "WARN: k3s not found in PATH — skipped containerd import." >&2
  echo "On the VPS run: docker save roadmap-api:local roadmap-frontend:local | sudo k3s ctr images import -" >&2
fi

echo "Done."
echo "If deployments already exist with imagePullPolicy IfNotPresent + :local tag:"
echo "  kubectl -n roadmap rollout restart deploy/roadmap-api deploy/roadmap-frontend"
