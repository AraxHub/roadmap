DROP TABLE IF EXISTS content_images;

ALTER TABLE submodule_contents
	ADD COLUMN IF NOT EXISTS body_md TEXT NOT NULL DEFAULT '';

-- Восстанавливаем body_md из первого markdown-блока (best-effort).
UPDATE submodule_contents
SET body_md = COALESCE(
	(
		SELECT elem->>'md'
		FROM jsonb_array_elements(blocks) AS elem
		WHERE elem->>'type' = 'markdown'
		LIMIT 1
	),
	''
);

ALTER TABLE submodule_contents DROP COLUMN IF EXISTS blocks;
