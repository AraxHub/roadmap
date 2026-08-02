#!/usr/bin/env bash
set -euo pipefail

# Creates/updates docker-registry Secret for pulling private GHCR images.
#
#   # add to secrets.env:
#   GHCR_USERNAME=AraxHub
#   GHCR_TOKEN=ghp_...   # classic PAT: read:packages (и delete:packages не нужен)
#
#   ./deployment/k3s/create-ghcr-pull-secret.sh
#
# Then: ./deployment/k3s/apply.sh

ROOT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
cd "$ROOT_DIR"

ENV_FILE=${ENV_FILE:-deployment/k3s/secrets.env}

if [[ ! -f "$ENV_FILE" ]]; then
  echo "Env file not found: $ENV_FILE" >&2
  echo "Create it from deployment/k3s/secrets.env.example" >&2
  exit 1
fi

set -a
# shellcheck disable=SC1090
source "$ENV_FILE"
set +a

require() {
  local name="$1"
  local val="${!name-}"
  if [[ -z "$val" || "$val" == "CHANGE_ME" ]]; then
    echo "Missing or placeholder: $name (set in $ENV_FILE)" >&2
    exit 1
  fi
}

require GHCR_USERNAME
require GHCR_TOKEN

GHCR_EMAIL="${GHCR_EMAIL:-noreply@users.noreply.github.com}"

kubectl create namespace roadmap --dry-run=client -o yaml | kubectl apply -f -

kubectl -n roadmap create secret docker-registry ghcr-pull \
  --docker-server=ghcr.io \
  --docker-username="${GHCR_USERNAME}" \
  --docker-password="${GHCR_TOKEN}" \
  --docker-email="${GHCR_EMAIL}" \
  --dry-run=client -o yaml | kubectl apply -f -

echo "Secret applied: roadmap/ghcr-pull"
echo "Next: ./deployment/k3s/apply.sh"
