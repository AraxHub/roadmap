-- Блочный контент подмодуля + картинки в Postgres (без S3).

ALTER TABLE submodule_contents
	ADD COLUMN IF NOT EXISTS blocks JSONB NOT NULL DEFAULT '[]'::jsonb;

-- Существующий markdown → один блок type=markdown.
UPDATE submodule_contents
SET blocks = jsonb_build_array(
	jsonb_build_object(
		'id', gen_random_uuid()::text,
		'type', 'markdown',
		'md', body_md
	)
)
WHERE body_md IS NOT NULL AND body_md <> '' AND (blocks = '[]'::jsonb OR blocks IS NULL);

ALTER TABLE submodule_contents DROP COLUMN IF EXISTS body_md;

CREATE TABLE IF NOT EXISTS content_images (
	id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	submodule_id UUID NOT NULL REFERENCES submodules(id) ON DELETE CASCADE,
	mime_type    TEXT NOT NULL,
	bytes        BYTEA NOT NULL,
	byte_size    INT NOT NULL,
	created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT content_images_mime_chk CHECK (mime_type IN ('image/png', 'image/jpeg', 'image/webp', 'image/gif')),
	CONSTRAINT content_images_size_chk CHECK (byte_size > 0 AND byte_size <= 10485760)
);

CREATE INDEX IF NOT EXISTS idx_content_images_submodule_id ON content_images(submodule_id);
