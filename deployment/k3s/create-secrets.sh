#!/usr/bin/env bash
set -euo pipefail

# Creates/updates Kubernetes Secret from deployment/k3s/secrets.env
#
#   cp deployment/k3s/secrets.env.example deployment/k3s/secrets.env
#   edit deployment/k3s/secrets.env
#   ./deployment/k3s/create-secrets.sh

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
    echo "Missing or placeholder secret: $name (set a real value in $ENV_FILE)" >&2
    exit 1
  fi
}

require ROADMAP_AUTH_JWT_SECRET
require ROADMAP_DB_PASSWORD
require ROADMAP_AUTH_BOOTSTRAP_PASSWORD

# Telegram optional at bootstrap, but required for prod bot features.
ROADMAP_TELEGRAM_BOT_TOKEN="${ROADMAP_TELEGRAM_BOT_TOKEN:-}"
ROADMAP_TELEGRAM_WEBHOOK_SECRET="${ROADMAP_TELEGRAM_WEBHOOK_SECRET:-}"

kubectl create namespace roadmap --dry-run=client -o yaml | kubectl apply -f -

kubectl -n roadmap create secret generic roadmap-secrets \
  --from-literal=ROADMAP_AUTH_JWT_SECRET="${ROADMAP_AUTH_JWT_SECRET}" \
  --from-literal=ROADMAP_DB_PASSWORD="${ROADMAP_DB_PASSWORD}" \
  --from-literal=ROADMAP_AUTH_BOOTSTRAP_PASSWORD="${ROADMAP_AUTH_BOOTSTRAP_PASSWORD}" \
  --from-literal=ROADMAP_TELEGRAM_BOT_TOKEN="${ROADMAP_TELEGRAM_BOT_TOKEN}" \
  --from-literal=ROADMAP_TELEGRAM_WEBHOOK_SECRET="${ROADMAP_TELEGRAM_WEBHOOK_SECRET}" \
  --dry-run=client -o yaml | kubectl apply -f -

echo "Secret applied: roadmap/roadmap-secrets"
if [[ -z "$ROADMAP_TELEGRAM_BOT_TOKEN" || -z "$ROADMAP_TELEGRAM_WEBHOOK_SECRET" ]]; then
  echo "NOTE: Telegram token/webhook secret empty — bot features disabled until set."
fi
