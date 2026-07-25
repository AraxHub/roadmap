CREATE TABLE IF NOT EXISTS sprints (
	id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	slug         TEXT NOT NULL,
	title        TEXT NOT NULL,
	description  TEXT NOT NULL DEFAULT '',
	position     INT  NOT NULL,
	is_published BOOLEAN NOT NULL DEFAULT FALSE,
	CONSTRAINT sprints_slug_uk UNIQUE (slug),
	CONSTRAINT sprints_position_uk UNIQUE (position)
);

CREATE TABLE IF NOT EXISTS modules (
	id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	sprint_id    UUID NOT NULL REFERENCES sprints(id) ON DELETE CASCADE,
	slug         TEXT NOT NULL,
	title        TEXT NOT NULL,
	description  TEXT NOT NULL DEFAULT '',
	position     INT  NOT NULL,
	is_published BOOLEAN NOT NULL DEFAULT FALSE,
	CONSTRAINT modules_sprint_slug_uk UNIQUE (sprint_id, slug),
	CONSTRAINT modules_sprint_position_uk UNIQUE (sprint_id, position)
);

CREATE TABLE IF NOT EXISTS submodules (
	id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	module_id    UUID NOT NULL REFERENCES modules(id) ON DELETE CASCADE,
	slug         TEXT NOT NULL,
	title        TEXT NOT NULL,
	position     INT  NOT NULL,
	is_published BOOLEAN NOT NULL DEFAULT FALSE,
	CONSTRAINT submodules_module_slug_uk UNIQUE (module_id, slug),
	CONSTRAINT submodules_module_position_uk UNIQUE (module_id, position)
);

CREATE TABLE IF NOT EXISTS submodule_contents (
	submodule_id UUID PRIMARY KEY REFERENCES submodules(id) ON DELETE CASCADE,
	body_md      TEXT NOT NULL DEFAULT '',
	updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS users (
	id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	login            TEXT NOT NULL,
	password_hash    TEXT NOT NULL,
	role             TEXT NOT NULL DEFAULT 'user',
	is_blocked       BOOLEAN NOT NULL DEFAULT FALSE,
	telegram_chat_id BIGINT,
	created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT users_login_uk UNIQUE (login),
	CONSTRAINT users_role_chk CHECK (role IN ('user', 'admin')),
	CONSTRAINT users_telegram_chat_id_uk UNIQUE (telegram_chat_id)
);

CREATE TABLE IF NOT EXISTS refresh_tokens (
	id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	token_hash TEXT NOT NULL,
	expires_at TIMESTAMPTZ NOT NULL,
	revoked_at TIMESTAMPTZ,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT refresh_tokens_hash_uk UNIQUE (token_hash)
);

CREATE TABLE IF NOT EXISTS user_submodule_progress (
	user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	submodule_id UUID NOT NULL REFERENCES submodules(id) ON DELETE CASCADE,
	completed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	PRIMARY KEY (user_id, submodule_id)
);

CREATE TABLE IF NOT EXISTS user_sprint_timers (
	user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	sprint_id    UUID NOT NULL REFERENCES sprints(id) ON DELETE CASCADE,
	started_at   TIMESTAMPTZ NOT NULL,
	deadline_at  TIMESTAMPTZ NOT NULL,
	completed_at TIMESTAMPTZ,
	PRIMARY KEY (user_id, sprint_id)
);

CREATE TABLE IF NOT EXISTS feedback_requests (
	id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	requested_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	tg_message_id BIGINT,
	status        TEXT NOT NULL DEFAULT 'pending',
	answered_at   TIMESTAMPTZ,
	answer_text   TEXT NOT NULL DEFAULT '',
	CONSTRAINT feedback_requests_status_chk CHECK (status IN ('pending', 'answered'))
);

-- Одна строка: слот авто-раунда ОС + текст просьбы (редактируется в админке).
CREATE TABLE IF NOT EXISTS feedback_schedule (
	id            SMALLINT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
	last_round_at TIMESTAMPTZ,
	message_text  TEXT NOT NULL DEFAULT 'Привет! Оставь, пожалуйста, обратную связь по обучению — ответь на это сообщение.'
);

INSERT INTO feedback_schedule (id, last_round_at, message_text) VALUES (1, NULL, 'Привет! Оставь, пожалуйста, обратную связь по обучению — ответь на это сообщение.')
ON CONFLICT (id) DO NOTHING;

CREATE INDEX IF NOT EXISTS idx_modules_sprint_id ON modules(sprint_id);
CREATE INDEX IF NOT EXISTS idx_submodules_module_id ON submodules(module_id);
CREATE INDEX IF NOT EXISTS idx_progress_user_id ON user_submodule_progress(user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_user_sprint_timers_user ON user_sprint_timers(user_id);
CREATE INDEX IF NOT EXISTS idx_feedback_requests_user ON feedback_requests(user_id);
CREATE INDEX IF NOT EXISTS idx_users_telegram_chat_id ON users(telegram_chat_id);
