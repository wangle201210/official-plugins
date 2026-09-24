BEGIN;

CREATE TEMP TABLE plugin_sicau_niu_iron (
    "id" BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "code" VARCHAR(64) NOT NULL,
    "deleted_at" TIMESTAMP
);

INSERT INTO plugin_sicau_niu_iron ("code") VALUES ('IRON-OLD');

\ir ../../../manifest/sql/015-sicau-niu-iron-bonus-toggle.sql

INSERT INTO plugin_sicau_niu_iron ("code") VALUES ('IRON-NEW');

DO $$
BEGIN
    IF (SELECT count(*) FROM plugin_sicau_niu_iron WHERE "bonus_enabled" = FALSE) <> 2 THEN
        RAISE EXCEPTION 'Existing and newly registered iron cows must default to bonus disabled';
    END IF;
END $$;

UPDATE plugin_sicau_niu_iron SET "bonus_enabled" = TRUE WHERE "code" = 'IRON-OLD';

\ir ../../../manifest/sql/015-sicau-niu-iron-bonus-toggle.sql

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM plugin_sicau_niu_iron
        WHERE "code" = 'IRON-OLD' AND "bonus_enabled" = TRUE
    ) OR NOT EXISTS (
        SELECT 1 FROM plugin_sicau_niu_iron
        WHERE "code" = 'IRON-NEW' AND "bonus_enabled" = FALSE
    ) THEN
        RAISE EXCEPTION 'Replaying the migration must preserve individual bonus switches';
    END IF;
END $$;

ROLLBACK;
