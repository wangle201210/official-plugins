-- ------------------------------------------------------------
-- 008 sicau-niu anti-cheat and idempotency SQL file
-- 008 sicau-niu 防作弊与写接口幂等 SQL 文件
-- Purpose: Frontend checklist hardening — client request_id idempotency for
--          feeding/steal/gift writes, the speed_anomaly activation attempt
--          result, and the new activation anti-cheat rule seeds.
-- Dialect: PostgreSQL. Idempotent. Table-existence guards keep this file safe
--          on partial schemas used by per-capability test harnesses.
-- ------------------------------------------------------------

DO $$
BEGIN
    IF to_regclass('plugin_sicau_niu_feeding') IS NOT NULL THEN
        ALTER TABLE plugin_sicau_niu_feeding
            ADD COLUMN IF NOT EXISTS "request_id" VARCHAR(64) NOT NULL DEFAULT '';
        COMMENT ON COLUMN plugin_sicau_niu_feeding."request_id"
            IS 'Client idempotency key deduplicating network retries; empty when not provided';
        CREATE UNIQUE INDEX IF NOT EXISTS uk_sicau_niu_feeding_user_request
            ON plugin_sicau_niu_feeding ("user_id", "request_id")
            WHERE "request_id" <> '' AND "deleted_at" IS NULL;
    END IF;
END $$;

DO $$
BEGIN
    IF to_regclass('plugin_sicau_niu_steal') IS NOT NULL THEN
        ALTER TABLE plugin_sicau_niu_steal
            ADD COLUMN IF NOT EXISTS "request_id" VARCHAR(64) NOT NULL DEFAULT '';
        COMMENT ON COLUMN plugin_sicau_niu_steal."request_id"
            IS 'Client idempotency key deduplicating network retries; empty when not provided';
        CREATE UNIQUE INDEX IF NOT EXISTS uk_sicau_niu_steal_actor_request
            ON plugin_sicau_niu_steal ("actor_user_id", "request_id")
            WHERE "request_id" <> '' AND "deleted_at" IS NULL;
    END IF;
END $$;

DO $$
BEGIN
    IF to_regclass('plugin_sicau_niu_gift') IS NOT NULL THEN
        ALTER TABLE plugin_sicau_niu_gift
            ADD COLUMN IF NOT EXISTS "request_id" VARCHAR(64) NOT NULL DEFAULT '';
        COMMENT ON COLUMN plugin_sicau_niu_gift."request_id"
            IS 'Client idempotency key deduplicating network retries; empty when not provided';
        CREATE UNIQUE INDEX IF NOT EXISTS uk_sicau_niu_gift_from_request
            ON plugin_sicau_niu_gift ("from_user_id", "request_id")
            WHERE "request_id" <> '' AND "deleted_at" IS NULL;
    END IF;
END $$;

DO $$
BEGIN
    IF to_regclass('plugin_sicau_niu_activation_attempt') IS NOT NULL THEN
        COMMENT ON COLUMN plugin_sicau_niu_activation_attempt."result"
            IS 'Attempt result: success, no_nearby, out_of_range, speed_anomaly';
    END IF;
END $$;

DO $$
BEGIN
    IF to_regclass('plugin_sicau_niu_rule_config') IS NOT NULL THEN
        INSERT INTO plugin_sicau_niu_rule_config ("config_key", "config_value", "remark")
        SELECT 'activation.dailyAttemptLimit', '20', '每日激活尝试次数上限（含失败）'
        WHERE NOT EXISTS (
            SELECT 1 FROM plugin_sicau_niu_rule_config
            WHERE "config_key" = 'activation.dailyAttemptLimit' AND "deleted_at" IS NULL
        );

        INSERT INTO plugin_sicau_niu_rule_config ("config_key", "config_value", "remark")
        SELECT 'activation.maxSpeedMps', '25', '激活定位移动速度上限（米/秒）'
        WHERE NOT EXISTS (
            SELECT 1 FROM plugin_sicau_niu_rule_config
            WHERE "config_key" = 'activation.maxSpeedMps' AND "deleted_at" IS NULL
        );
    END IF;
END $$;
