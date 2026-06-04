-- ------------------------------------------------------------
-- 006 sicau-niu settlement SQL file
-- 006 sicau-niu 结算 SQL 文件
-- Purpose: Operator settlement archive snapshots for C7 niu-settlement. The
--          dashboard, export, batch certificate issuance and risk-alert views are
--          derived from C1-C5 tables and own no table; only the archive snapshot
--          is persisted here.
-- Dialect: PostgreSQL. Idempotent.
-- ------------------------------------------------------------

CREATE TABLE IF NOT EXISTS plugin_sicau_niu_settlement (
    "id"          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "title"       VARCHAR(128) NOT NULL DEFAULT '',
    "snapshot"    TEXT NOT NULL DEFAULT '',
    "operator_id" BIGINT NOT NULL DEFAULT 0,
    "archived_at" TIMESTAMP,
    "created_at"  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "updated_at"  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "deleted_at"  TIMESTAMP
);

COMMENT ON TABLE plugin_sicau_niu_settlement IS 'sicau-niu settlement archive snapshot table (C7)';
COMMENT ON COLUMN plugin_sicau_niu_settlement."title" IS 'Archive title given by the operator';
COMMENT ON COLUMN plugin_sicau_niu_settlement."snapshot" IS 'Frozen dashboard metrics serialized as JSON text';
COMMENT ON COLUMN plugin_sicau_niu_settlement."operator_id" IS 'Host operator user ID who created the archive';
COMMENT ON COLUMN plugin_sicau_niu_settlement."archived_at" IS 'Archive time, set explicitly by the settlement service';
COMMENT ON COLUMN plugin_sicau_niu_settlement."created_at" IS 'Creation time';
COMMENT ON COLUMN plugin_sicau_niu_settlement."updated_at" IS 'Update time';
COMMENT ON COLUMN plugin_sicau_niu_settlement."deleted_at" IS 'Soft-delete time, NULL means active';

-- Index supporting the bounded archive list ordered by archive time descending.
CREATE INDEX IF NOT EXISTS idx_sicau_niu_settlement_archived
    ON plugin_sicau_niu_settlement ("archived_at" DESC) WHERE "deleted_at" IS NULL;

-- Index supporting the C7 shared-device risk grouping (and the C1 one-device-many-
-- accounts audit): the risk view groups players on a non-empty device fingerprint.
-- Added here as the C7 iteration SQL; idempotent so re-runs are safe.
CREATE INDEX IF NOT EXISTS idx_sicau_niu_user_device_fingerprint
    ON plugin_sicau_niu_user ("device_fingerprint")
    WHERE "deleted_at" IS NULL AND "device_fingerprint" <> '';
