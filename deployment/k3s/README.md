# Roadmap на VPS (тот же k3s, что и Voco)

Предполагается, что на сервере уже есть:
- k3s
- ingress-nginx (NodePort 80/443)
- cert-manager + ClusterIssuer `letsencrypt-prod`

Roadmap живёт в namespace `roadmap`, Voco — в `voco`.
Роутинг — по Host в Ingress (не нужен отдельный host nginx).

PostgreSQL — **на хосте в Docker**, вне куба (см. `deployment/postgres/`).

---

## 0) Перед деплоем

1. Замени домен `roadmap.example.ru` на свой в:
   - `deployment/k8s/overlays/k3s/base/ingress.yaml`
   - `deployment/k8s/overlays/k3s/api-configmap.yaml` (`ROADMAP_TELEGRAM_WEBHOOK_URL`)
2. DNS A-запись: `roadmap.<домен>` → публичный IP VPS.
3. Открой на фаерволле только то, что уже нужно для ingress (80/443). **Не открывай 5432 наружу.**

---

## 1) Shared Postgres на хосте

```bash
cd deployment/postgres
cp .env.example .env
nano .env   # пароли
chmod +x init/01-init-databases.sh
docker compose up -d
docker compose ps
```

Из подов k3s БД доступна как `host.k3s.internal:5432`
(это уже прописано в ConfigMap).

Проверка с ноды:

```bash
docker exec -it shared-postgres psql -U postgres -c '\l'
```

---

## 2) Секреты приложения

```bash
cp deployment/k3s/secrets.env.example deployment/k3s/secrets.env
nano deployment/k3s/secrets.env
./deployment/k3s/create-secrets.sh
```

`ROADMAP_DB_PASSWORD` должен совпадать с паролем роли `roadmap_app` в Postgres.

---

## 3) Сборка образов на VPS

```bash
./deployment/k3s/build-images.sh
```

Скрипт делает `docker build` и импортирует образы в containerd k3s
(`k3s ctr images import`). Без импорта поды не увидят `:local` образы.

---

## 4) Apply

```bash
./deployment/k3s/apply.sh
kubectl -n roadmap get pods,svc,ing
```

Сертификат: `kubectl -n roadmap describe certificate` (или `get certificate`).

Проверки:

```bash
curl -i https://roadmap.<домен>/liveness
curl -i https://roadmap.<домен>/readyness
curl -i https://roadmap.<домен>/
kubectl -n roadmap logs -f deploy/roadmap-api
```

Ожидаем: миграция `schema_migrations.version = 1`, bootstrap admin из секрета.

---

## 5) Обновление кода

```bash
git pull
./deployment/k3s/build-images.sh
./deployment/k3s/restart.sh
```

`apply.sh` нужен снова только если менялись манифесты/ConfigMap.

При смене секретов:

```bash
./deployment/k3s/create-secrets.sh
./deployment/k3s/restart.sh
```

---

## Telegram (prod)

После того как HTTPS живой:

1. В `secrets.env` — `ROADMAP_TELEGRAM_BOT_TOKEN` и `ROADMAP_TELEGRAM_WEBHOOK_SECRET`
2. В ConfigMap — `ROADMAP_TELEGRAM_BOT_USERNAME` и корректный
   `ROADMAP_TELEGRAM_WEBHOOK_URL=https://<host>/api/v1/telegram/webhook`
3. `create-secrets.sh` → `apply.sh` / `restart.sh`

При `ROADMAP_ENV=prod` API сам вызывает `setWebhook`.

---

## Как устроен трафик

```
Browser / Telegram
  → DNS roadmap.<домен>
  → ingress-nginx :443
  → Ingress "roadmap"
       /api/* → Service roadmap-api:8080
       /*     → Service roadmap-frontend:80
  roadmap-api → host.k3s.internal:5432 (Docker Postgres)
```

Voco на своих хостах (`www` / `lk` / …) не пересекается с Roadmap.

---

## Rollback

При теге `:local` история образов слабая. Практичный откат:

```bash
git checkout <prev>
./deployment/k3s/build-images.sh
./deployment/k3s/restart.sh
```

Либо `kubectl -n roadmap rollout undo deploy/roadmap-api` — только если
предыдущий ReplicaSet ещё на ноде с тем же digest (ненадёжно при перезаписи `:local`).

---

## Важно

- API: **1 реплика**, `strategy: Recreate` — из-за in-process feedback worker и миграций на старте.
- CORS в коде заточен под localhost; в проде same-origin через Ingress, CORS не нужен.
- ClusterIssuer `letsencrypt-prod` должен уже существовать (как у Voco).
- Пароли Postgres в `deployment/postgres/.env` и `deployment/k3s/secrets.env` не коммитить.
