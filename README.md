# roadmap

Обучающая платформа-роадмап: спринты → модули → подмодули (MD), auth (login/password), роли user/admin.

## Auth

- Access JWT (TTL из конфига, default 15m) — в `Authorization: Bearer`, на React держать в memory.
- Refresh (default 72h) — httpOnly cookie `refresh_token`, path `/api/v1/auth`.
- Роли: `user`, `admin`.
- Bootstrap-админ: `ROADMAP_AUTH_BOOTSTRAP_LOGIN/PASSWORD` (создаётся только если users пуст).

| Метод | Путь | Auth | Описание |
|-------|------|------|----------|
| POST | `/api/v1/auth/login` | нет | `{login,password}` → access + cookie |
| POST | `/api/v1/auth/refresh` | cookie | новая пара токенов |
| POST | `/api/v1/auth/logout` | access | revoke refresh |
| GET | `/api/v1/auth/me` | access | текущий юзер |
| POST | `/api/v1/admin/users` | admin | создать ЛК → `{id,login,password,role}` для модалки |
| GET | `/api/v1/admin/users` | admin | список пользователей |
| GET | `/api/v1/admin/users/:id/feedback` | admin | история обратной связи |
| POST | `/api/v1/admin/feedback/request-round` | admin | разослать просьбу об ОС сейчас |
| GET | `/api/v1/admin/feedback/schedule` | admin | когда следующий авто-раунд ОС + текст просьбы |
| PUT | `/api/v1/admin/feedback/message` | admin | обновить текст просьбы об ОС |
| POST | `/api/v1/telegram/webhook` | secret | webhook бота (`X-Telegram-Bot-Api-Secret-Token`) |
| GET | `/api/v1/home` | access | главная (`feedback_required`, таймеры спринтов) |
| GET | `/api/v1/modules/:slug` | access | модуль |
| GET | `/api/v1/modules/:m/submodules/:s` | access | контент |
| POST | `/api/v1/submodules/:id/complete` | access | завершить |

## Запуск

Нужен чистый Postgres (старая таблица `users` с email несовместима — `DROP_VOLUMES=1`).

```bash
make from-zero DROP_VOLUMES=1
```

- UI: http://localhost:3000 (`admin` / `admin`)
- API напрямую: http://localhost:8080
- Postgres: localhost:5434

Локально без docker для фронта: `cd frontend && npm run dev` (proxy на `:8080`).

## Спринты и Telegram

| Переменная | Default | Смысл |
|------------|---------|--------|
| `ROADMAP_ENV` | `dev` | `dev` → long polling; `prod` → `setWebhook` из приложения |
| `ROADMAP_SPRINT_DURATION` | `168h` | срок спринта на пользователя |
| `ROADMAP_TELEGRAM_BOT_TOKEN` | — | токен бота (без него TG выключен) |
| `ROADMAP_TELEGRAM_BOT_USERNAME` | — | username без `@` |
| `ROADMAP_TELEGRAM_WEBHOOK_SECRET` | — | secret для `X-Telegram-Bot-Api-Secret-Token` |
| `ROADMAP_TELEGRAM_WEBHOOK_URL` | — | полный HTTPS URL webhook (нужен в `prod`) |
| `ROADMAP_TELEGRAM_POLLING_TIMEOUT` | `30` | timeout getUpdates в секундах (`dev`) |
| `ROADMAP_TELEGRAM_FEEDBACK_TZ` | `Europe/Moscow` | TZ слота авто-ОС |

Авто-опрос ОС: каждый **четверг в 12:00** (TZ из `FEEDBACK_TZ`). Слот последнего раунда и **текст просьбы** хранятся в `feedback_schedule` — рестарт не сбрасывает расписание, текст правится в админке (ЛК) без редеплоя. Если приложение было выключено в четверг, раунд догоняется при старте. Ручная кнопка в админке шлёт одному пользователю вне расписания.

Привязка пользователя: в боте Start → следующим сообщением логин. Сообщения от неизвестных `chat_id` игнорируются (кроме сценария привязки).

В `prod` приложение само вызывает Telegram `setWebhook` на `ROADMAP_TELEGRAM_WEBHOOK_URL`. В `dev` снимает webhook и слушает updates через polling.