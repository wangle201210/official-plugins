BEGIN;

CREATE TEMP TABLE plugin_sicau_niu_niu (
    "id" BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "code" VARCHAR(32) NOT NULL,
    "niu_type" VARCHAR(16) NOT NULL,
    "name" VARCHAR(64) NOT NULL DEFAULT '',
    "lat" DOUBLE PRECISION NOT NULL,
    "lng" DOUBLE PRECISION NOT NULL,
    "online_at" TIMESTAMPTZ,
    "status" VARCHAR(16) NOT NULL,
    "deleted_at" TIMESTAMP
);

CREATE UNIQUE INDEX uk_sicau_niu_niu_code
    ON plugin_sicau_niu_niu ("code") WHERE "deleted_at" IS NULL;

\ir ../../../manifest/sql/013-sicau-niu-chengdu-cattle-seed.sql

DO $$
BEGIN
    IF (SELECT count(*) FROM plugin_sicau_niu_niu) <> 12
       OR (SELECT count(DISTINCT "code") FROM plugin_sicau_niu_niu) <> 12
       OR (SELECT count(DISTINCT ("lat", "lng")) FROM plugin_sicau_niu_niu) <> 12
       OR EXISTS (
           SELECT 1 FROM plugin_sicau_niu_niu
           WHERE "code" NOT BETWEEN 'NIU-001' AND 'NIU-012'
              OR "status" <> 'inactive'
              OR "niu_type" <> 'common'
              OR "name" <> ''
              OR "online_at" IS NULL
              OR "online_at" > CURRENT_TIMESTAMP
              OR "lat" NOT BETWEEN 30.70404 AND 30.70676
              OR "lng" NOT BETWEEN 103.85975 AND 103.86665
       ) THEN
        RAISE EXCEPTION 'Chengdu cattle seed does not provide 12 unique, online, inactive campus cattle';
    END IF;
END $$;

UPDATE plugin_sicau_niu_niu
SET "name" = '运营已调整', "status" = 'active'
WHERE "code" = 'NIU-001';

\ir ../../../manifest/sql/013-sicau-niu-chengdu-cattle-seed.sql

DO $$
BEGIN
    IF (SELECT count(*) FROM plugin_sicau_niu_niu) <> 12
       OR NOT EXISTS (
           SELECT 1 FROM plugin_sicau_niu_niu
           WHERE "code" = 'NIU-001' AND "name" = '运营已调整' AND "status" = 'active'
       ) THEN
        RAISE EXCEPTION 'Chengdu cattle seed is not idempotent or overwrote an operator edit';
    END IF;
END $$;

DELETE FROM plugin_sicau_niu_niu
WHERE "code" IN ('NIU-011', 'NIU-012');

\ir ../../../manifest/sql/013-sicau-niu-chengdu-cattle-seed.sql

DO $$
BEGIN
    IF (SELECT count(*) FROM plugin_sicau_niu_niu) <> 12
       OR (SELECT count(*) FROM plugin_sicau_niu_niu WHERE "code" IN ('NIU-011', 'NIU-012') AND "status" = 'inactive') <> 2
       OR NOT EXISTS (
           SELECT 1 FROM plugin_sicau_niu_niu
           WHERE "code" = 'NIU-001' AND "name" = '运营已调整' AND "status" = 'active'
       ) THEN
        RAISE EXCEPTION 'Chengdu cattle seed did not fill only the missing cattle';
    END IF;
END $$;

ROLLBACK;
