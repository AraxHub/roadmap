На VPS из `~/roadmap`:

## GitHub Actions (CI + CD)

### Workflows

- **CI** — PR / push в `main`: тесты и сборка.
- **Publish images** — push в `main`: сборка → GHCR → **деплой на VPS по SSH** (`kubectl set image` + rollout).

Образы:

- `ghcr.io/araxhub/roadmap-api:sha-<7>` / `:latest`
- `ghcr.io/araxhub/roadmap-frontend:sha-<7>` / `:latest`

### Один раз: secrets в GitHub

Repo → **Settings → Secrets and variables → Actions**:

| Secret | Значение |
|--------|----------|
| `VPS_HOST` | IP или hostname VPS |
| `VPS_USER` | SSH-пользователь (часто `root`) |
| `VPS_SSH_KEY` | приватный ключ целиком (`-----BEGIN…`) |
| `VPS_SSH_PORT` | опционально, по умолчанию `22` |

На VPS у этого пользователя должен работать `kubectl` (обычно `~/.kube/config` от k3s).

### Один раз: pull из GHCR на VPS

1. GitHub → Settings → Developer settings → **Personal access tokens (classic)**  
   scopes: `read:packages` (и `write:packages` не обязателен для pull).
2. На VPS в `~/roadmap/deployment/k3s/secrets.env` добавь:

```bash
GHCR_USERNAME=AraxHub
GHCR_TOKEN=ghp_...
```

3. Примени:

```bash
cd ~/roadmap
git pull
./deployment/k3s/create-ghcr-pull-secret.sh
./deployment/k3s/apply.sh
```

`apply.sh` переведёт деплои на `ghcr.io/araxhub/...:latest` + `imagePullSecrets: ghcr-pull`.

Дальше обычный push в `main` сам обновит поды на `sha-…`.

Если пакеты private и pull падает с `401` — проверь PAT и секрет `ghcr-pull`.  
Альтернатива: сделать пакеты Public в GitHub → Packages → package settings.

---

## Ручной деплой (fallback, без CI)

Раньше: сборка на VPS. Сейчас предпочтительно CI; локальный билд только если GHCR/SSH недоступны:

```bash
cd ~/roadmap
git pull
./deployment/k3s/build-images.sh
# Внимание: манифесты ждут GHCR. Для :local временно откатить images в kustomization
# или: kubectl -n roadmap set image ... roadmap-api:local
./deployment/k3s/restart.sh
```

Менялись только манифесты/ConfigMap — `./deployment/k3s/apply.sh`.

## Секреты приложения / Telegram

```bash
# правка deployment/k3s/secrets.env →
./deployment/k3s/create-secrets.sh
./deployment/k3s/restart.sh   # или дождаться следующего CI-деплоя
```

**Миграции** накатываются сами при старте API.
