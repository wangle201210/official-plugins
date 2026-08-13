-- ------------------------------------------------------------
-- 012 sicau-niu runtime hardening SQL file
-- Purpose: Reset legacy write facts when any replay snapshot is unavailable,
--          then persist first successful feeding, steal, gift, team creation
--          and team join responses so retries never apply a write twice.
-- Dialect: PostgreSQL. Idempotent: safe to re-run.
-- ------------------------------------------------------------

DO $$
BEGIN
    IF to_regclass('plugin_sicau_niu_grass_account') IS NOT NULL
       AND to_regclass('plugin_sicau_niu_grass_txn') IS NOT NULL
       AND to_regclass('plugin_sicau_niu_checkin') IS NOT NULL
       AND to_regclass('plugin_sicau_niu_feeding') IS NOT NULL
       AND to_regclass('plugin_sicau_niu_steal') IS NOT NULL
       AND to_regclass('plugin_sicau_niu_gift') IS NOT NULL
       AND to_regclass('plugin_sicau_niu_inbox_msg') IS NOT NULL
       AND (
            NOT EXISTS (
                SELECT 1 FROM pg_attribute
                WHERE attrelid = to_regclass('plugin_sicau_niu_feeding')
                  AND attname = 'response_json'
                  AND NOT attisdropped
            )
            OR NOT EXISTS (
                SELECT 1 FROM pg_attribute
                WHERE attrelid = to_regclass('plugin_sicau_niu_steal')
                  AND attname = 'result_balance'
                  AND NOT attisdropped
            )
            OR NOT EXISTS (
                SELECT 1 FROM pg_attribute
                WHERE attrelid = to_regclass('plugin_sicau_niu_gift')
                  AND attname = 'result_balance'
                  AND NOT attisdropped
            )
       ) THEN
        TRUNCATE TABLE
            plugin_sicau_niu_inbox_msg,
            plugin_sicau_niu_gift,
            plugin_sicau_niu_steal,
            plugin_sicau_niu_feeding,
            plugin_sicau_niu_checkin,
            plugin_sicau_niu_grass_txn,
            plugin_sicau_niu_grass_account;
    END IF;
END $$;

DO $$
BEGIN
    IF to_regclass('plugin_sicau_niu_feeding') IS NOT NULL THEN
        ALTER TABLE plugin_sicau_niu_feeding
            ADD COLUMN IF NOT EXISTS "response_json" TEXT NOT NULL DEFAULT '';
        IF EXISTS (
            SELECT 1 FROM pg_attribute
            WHERE attrelid = to_regclass('plugin_sicau_niu_feeding')
              AND attname = 'request_id'
              AND NOT attisdropped
        ) THEN
            COMMENT ON COLUMN plugin_sicau_niu_feeding."request_id"
                IS 'Required player-scoped idempotency key for stable replay';
        END IF;
        COMMENT ON COLUMN plugin_sicau_niu_feeding."response_json"
            IS 'Stable first successful feeding response for idempotent replay';
    END IF;
END $$;

DO $$
BEGIN
    IF to_regclass('plugin_sicau_niu_transport_team') IS NOT NULL
       AND to_regclass('plugin_sicau_niu_transport_member') IS NOT NULL
       AND to_regclass('plugin_sicau_niu_transport_report') IS NOT NULL
       AND (
            NOT EXISTS (
                SELECT 1 FROM pg_attribute
                WHERE attrelid = to_regclass('plugin_sicau_niu_transport_team')
                  AND attname = 'create_response_json'
                  AND NOT attisdropped
            )
            OR NOT EXISTS (
                SELECT 1 FROM pg_attribute
                WHERE attrelid = to_regclass('plugin_sicau_niu_transport_member')
                  AND attname = 'join_response_json'
                  AND NOT attisdropped
            )
       ) THEN
        TRUNCATE TABLE
            plugin_sicau_niu_transport_report,
            plugin_sicau_niu_transport_member,
            plugin_sicau_niu_transport_team
        RESTART IDENTITY;
    END IF;
END $$;

DO $$
BEGIN
    IF to_regclass('plugin_sicau_niu_transport_team') IS NOT NULL THEN
        ALTER TABLE plugin_sicau_niu_transport_team
            ADD COLUMN IF NOT EXISTS "create_response_json" TEXT NOT NULL DEFAULT '';
        IF EXISTS (
            SELECT 1 FROM pg_attribute
            WHERE attrelid = to_regclass('plugin_sicau_niu_transport_team')
              AND attname = 'create_request_id'
              AND NOT attisdropped
        ) THEN
            COMMENT ON COLUMN plugin_sicau_niu_transport_team."create_request_id"
                IS 'Required creator-scoped idempotency key for stable replay';
        END IF;
        COMMENT ON COLUMN plugin_sicau_niu_transport_team."create_response_json"
            IS 'Stable first successful team creation response for idempotent replay';
    END IF;
END $$;

DO $$
BEGIN
    IF to_regclass('plugin_sicau_niu_transport_member') IS NOT NULL THEN
        ALTER TABLE plugin_sicau_niu_transport_member
            ADD COLUMN IF NOT EXISTS "join_response_json" TEXT NOT NULL DEFAULT '';
        IF EXISTS (
            SELECT 1 FROM pg_attribute
            WHERE attrelid = to_regclass('plugin_sicau_niu_transport_member')
              AND attname = 'join_request_id'
              AND NOT attisdropped
        ) THEN
            COMMENT ON COLUMN plugin_sicau_niu_transport_member."join_request_id"
                IS 'Required joiner-scoped idempotency key; empty for creator membership';
        END IF;
        COMMENT ON COLUMN plugin_sicau_niu_transport_member."join_response_json"
            IS 'Stable first successful team join response for idempotent replay';
    END IF;
END $$;

DO $$
BEGIN
    IF to_regclass('plugin_sicau_niu_steal') IS NOT NULL THEN
        ALTER TABLE plugin_sicau_niu_steal
            ADD COLUMN IF NOT EXISTS "result_balance" BIGINT NOT NULL DEFAULT 0;
        IF EXISTS (
            SELECT 1 FROM pg_attribute
            WHERE attrelid = to_regclass('plugin_sicau_niu_steal')
              AND attname = 'request_id'
              AND NOT attisdropped
        ) THEN
            COMMENT ON COLUMN plugin_sicau_niu_steal."request_id"
                IS 'Required player-scoped idempotency key for stable replay';
        END IF;
        COMMENT ON COLUMN plugin_sicau_niu_steal."result_balance"
            IS 'Actor balance returned by the first successful steal request';
    END IF;
END $$;

DO $$
BEGIN
    IF to_regclass('plugin_sicau_niu_gift') IS NOT NULL THEN
        ALTER TABLE plugin_sicau_niu_gift
            ADD COLUMN IF NOT EXISTS "result_balance" BIGINT NOT NULL DEFAULT 0;
        IF EXISTS (
            SELECT 1 FROM pg_attribute
            WHERE attrelid = to_regclass('plugin_sicau_niu_gift')
              AND attname = 'request_id'
              AND NOT attisdropped
        ) THEN
            COMMENT ON COLUMN plugin_sicau_niu_gift."request_id"
                IS 'Required player-scoped idempotency key for stable replay';
        END IF;
        COMMENT ON COLUMN plugin_sicau_niu_gift."result_balance"
            IS 'Giver balance returned by the first successful gift request';
    END IF;
END $$;
