BEGIN;

CREATE TEMP TABLE plugin_sicau_niu_iron (
    "id" BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "code" VARCHAR(64) NOT NULL,
    "name" VARCHAR(64) NOT NULL,
    "last_lat" DOUBLE PRECISION,
    "last_lng" DOUBLE PRECISION,
    "located_at" TIMESTAMPTZ,
    "deleted_at" TIMESTAMP
);

CREATE UNIQUE INDEX uk_sicau_niu_iron_code
    ON plugin_sicau_niu_iron ("code") WHERE "deleted_at" IS NULL;

\ir ../../../manifest/sql/014-sicau-niu-iron-registry-seed.sql

DO $$
BEGIN
    IF (SELECT count(*) FROM plugin_sicau_niu_iron) <> 21
       OR EXISTS (
           (SELECT "code" FROM plugin_sicau_niu_iron)
           EXCEPT
           (VALUES
               ('50272057919'), ('50271994245'), ('50272012203'),
               ('50272012237'), ('50272012427'), ('50272030338'),
               ('50272030619'), ('50272040220'), ('50272042176'),
               ('50272043158'), ('50272043299'), ('50272055459'),
               ('50272055541'), ('50272056069'), ('50272056267'),
               ('50272056481'), ('50272056762'), ('50272057158'),
               ('50272057836'), ('50283597560'), ('50275156712')
           )
       )
       OR EXISTS (
           (VALUES
               ('50272057919'), ('50271994245'), ('50272012203'),
               ('50272012237'), ('50272012427'), ('50272030338'),
               ('50272030619'), ('50272040220'), ('50272042176'),
               ('50272043158'), ('50272043299'), ('50272055459'),
               ('50272055541'), ('50272056069'), ('50272056267'),
               ('50272056481'), ('50272056762'), ('50272057158'),
               ('50272057836'), ('50283597560'), ('50275156712')
           )
           EXCEPT
           (SELECT "code" FROM plugin_sicau_niu_iron)
       )
       OR EXISTS (
           SELECT 1 FROM plugin_sicau_niu_iron
           WHERE "name" <> "code" OR "last_lat" IS NOT NULL
              OR "last_lng" IS NOT NULL OR "located_at" IS NOT NULL
       ) THEN
        RAISE EXCEPTION 'Iron registry seed differs from the 21 locator identifiers or has fabricated positions';
    END IF;
END $$;

UPDATE plugin_sicau_niu_iron
SET "name" = '运营名称', "last_lat" = 30.7054, "last_lng" = 103.8632,
    "located_at" = CURRENT_TIMESTAMP
WHERE "code" = '50275156712';

\ir ../../../manifest/sql/014-sicau-niu-iron-registry-seed.sql

DO $$
BEGIN
    IF (SELECT count(*) FROM plugin_sicau_niu_iron) <> 21
       OR NOT EXISTS (
           SELECT 1 FROM plugin_sicau_niu_iron
           WHERE "code" = '50275156712' AND "name" = '运营名称'
             AND "last_lat" = 30.7054 AND "last_lng" = 103.8632
             AND "located_at" IS NOT NULL
       ) THEN
        RAISE EXCEPTION 'Iron registry seed overwrote an existing locator';
    END IF;
END $$;

DELETE FROM plugin_sicau_niu_iron WHERE "code" = '50272057919';

\ir ../../../manifest/sql/014-sicau-niu-iron-registry-seed.sql

DO $$
BEGIN
    IF (SELECT count(*) FROM plugin_sicau_niu_iron) <> 21
       OR NOT EXISTS (
           SELECT 1 FROM plugin_sicau_niu_iron
           WHERE "code" = '50272057919' AND "name" = "code" AND "last_lat" IS NULL
       ) THEN
        RAISE EXCEPTION 'Iron registry seed did not restore only the missing locator';
    END IF;
END $$;

ROLLBACK;
