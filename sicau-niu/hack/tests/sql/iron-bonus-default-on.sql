BEGIN;

CREATE TEMP TABLE plugin_sicau_niu_iron (
    "id" BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "code" VARCHAR(64) NOT NULL,
    "bonus_enabled" BOOLEAN NOT NULL DEFAULT FALSE,
    "deleted_at" TIMESTAMP
);

INSERT INTO plugin_sicau_niu_iron ("code", "deleted_at") VALUES
    ('IRON-EXISTING', NULL),
    ('IRON-ARCHIVED', NOW());

\ir ../../../manifest/sql/016-sicau-niu-iron-bonus-default-on.sql

INSERT INTO plugin_sicau_niu_iron ("code") VALUES ('IRON-NEW');

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM plugin_sicau_niu_iron
        WHERE "code" = 'IRON-EXISTING' AND "bonus_enabled" = TRUE
    ) OR NOT EXISTS (
        SELECT 1 FROM plugin_sicau_niu_iron
        WHERE "code" = 'IRON-NEW' AND "bonus_enabled" = TRUE
    ) OR NOT EXISTS (
        SELECT 1 FROM plugin_sicau_niu_iron
        WHERE "code" = 'IRON-ARCHIVED' AND "bonus_enabled" = FALSE
    ) THEN
        RAISE EXCEPTION 'Active existing and new iron cows must default on without changing archived rows';
    END IF;
END $$;

UPDATE plugin_sicau_niu_iron SET "bonus_enabled" = FALSE WHERE "code" = 'IRON-EXISTING';

\ir ../../../manifest/sql/016-sicau-niu-iron-bonus-default-on.sql

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM plugin_sicau_niu_iron
        WHERE "code" = 'IRON-EXISTING' AND "bonus_enabled" = FALSE
    ) OR (
        SELECT column_default FROM information_schema.columns
        WHERE table_schema LIKE 'pg_temp_%' AND table_name = 'plugin_sicau_niu_iron'
          AND column_name = 'bonus_enabled'
    ) IS DISTINCT FROM 'true' THEN
        RAISE EXCEPTION 'Replaying the migration must preserve an explicitly disabled iron cow';
    END IF;
END $$;

ROLLBACK;
