-- ------------------------------------------------------------
-- 004 sicau-niu grass SQL file
-- 004 sicau-niu 草经济与社交 SQL 文件
-- Purpose: Grass account/ledger, check-in, feeding (+iron bonus), steal, gift, inbox for C4 niu-grass.
-- 用途:草账户/流水、签到、喂草(含铁牛加成)、偷草、送草、站内信,C4 niu-grass。
-- Dialect: PostgreSQL. Idempotent: safe to re-run.
-- ------------------------------------------------------------

-- 草账户(账本余额)。
CREATE TABLE IF NOT EXISTS plugin_sicau_niu_grass_account (
    "id"         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "user_id"    BIGINT NOT NULL DEFAULT 0,
    "balance"    BIGINT NOT NULL DEFAULT 0,
    "created_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "deleted_at" TIMESTAMP
);
COMMENT ON TABLE plugin_sicau_niu_grass_account IS 'sicau-niu grass account balance table';
COMMENT ON COLUMN plugin_sicau_niu_grass_account."user_id" IS 'Owning player ID';
COMMENT ON COLUMN plugin_sicau_niu_grass_account."balance" IS 'Current grass balance';
COMMENT ON COLUMN plugin_sicau_niu_grass_account."deleted_at" IS 'Soft-delete time, NULL means active';
CREATE UNIQUE INDEX IF NOT EXISTS uk_sicau_niu_grass_account_user
    ON plugin_sicau_niu_grass_account ("user_id") WHERE "deleted_at" IS NULL;

-- 草流水(追加式账本)。
CREATE TABLE IF NOT EXISTS plugin_sicau_niu_grass_txn (
    "id"         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "user_id"    BIGINT NOT NULL DEFAULT 0,
    "delta"      BIGINT NOT NULL DEFAULT 0,
    "txn_type"   VARCHAR(20) NOT NULL DEFAULT '',
    "ref_id"     BIGINT NOT NULL DEFAULT 0,
    "created_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "deleted_at" TIMESTAMP
);
COMMENT ON TABLE plugin_sicau_niu_grass_txn IS 'sicau-niu grass ledger transaction table';
COMMENT ON COLUMN plugin_sicau_niu_grass_txn."delta" IS 'Signed grass change: positive credit, negative debit';
COMMENT ON COLUMN plugin_sicau_niu_grass_txn."txn_type" IS 'Type: checkin, feed, steal_gain, stolen_loss, gift_out, gift_in';
COMMENT ON COLUMN plugin_sicau_niu_grass_txn."ref_id" IS 'Related business record ID (feeding/steal/gift/checkin)';
CREATE INDEX IF NOT EXISTS idx_sicau_niu_grass_txn_user ON plugin_sicau_niu_grass_txn ("user_id");

-- 每日签到。
CREATE TABLE IF NOT EXISTS plugin_sicau_niu_checkin (
    "id"           BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "user_id"      BIGINT NOT NULL DEFAULT 0,
    "checkin_date" VARCHAR(10) NOT NULL DEFAULT '',
    "amount"       INT NOT NULL DEFAULT 0,
    "created_at"   TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "updated_at"   TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "deleted_at"   TIMESTAMP
);
COMMENT ON TABLE plugin_sicau_niu_checkin IS 'sicau-niu daily check-in table';
COMMENT ON COLUMN plugin_sicau_niu_checkin."checkin_date" IS 'Check-in date YYYY-MM-DD';
COMMENT ON COLUMN plugin_sicau_niu_checkin."amount" IS 'Granted grass amount (20-50)';
CREATE UNIQUE INDEX IF NOT EXISTS uk_sicau_niu_checkin_user_date
    ON plugin_sicau_niu_checkin ("user_id", "checkin_date") WHERE "deleted_at" IS NULL;

-- 喂草记录(含铁牛加成)。
CREATE TABLE IF NOT EXISTS plugin_sicau_niu_feeding (
    "id"                BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "user_id"           BIGINT NOT NULL DEFAULT 0,
    "niu_id"            BIGINT NOT NULL DEFAULT 0,
    "base_amount"       INT NOT NULL DEFAULT 0,
    "coefficient_basis" INT NOT NULL DEFAULT 100,
    "effect_amount"     INT NOT NULL DEFAULT 0,
    "is_iron_bonus"     INT NOT NULL DEFAULT 0,
    "fed_at"            TIMESTAMP,
    "created_at"        TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "updated_at"        TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "deleted_at"        TIMESTAMP
);
COMMENT ON TABLE plugin_sicau_niu_feeding IS 'sicau-niu feeding record table';
COMMENT ON COLUMN plugin_sicau_niu_feeding."base_amount" IS 'Original grass amount fed';
COMMENT ON COLUMN plugin_sicau_niu_feeding."coefficient_basis" IS 'Bonus coefficient in basis of 100: 100=x1.0, 150=x1.5';
COMMENT ON COLUMN plugin_sicau_niu_feeding."effect_amount" IS 'Actual feeding effect = base_amount * coefficient_basis / 100';
COMMENT ON COLUMN plugin_sicau_niu_feeding."is_iron_bonus" IS 'Whether an iron-cow proximity bonus applied: 1 yes, 0 no';
CREATE INDEX IF NOT EXISTS idx_sicau_niu_feeding_user ON plugin_sicau_niu_feeding ("user_id");
CREATE INDEX IF NOT EXISTS idx_sicau_niu_feeding_niu ON plugin_sicau_niu_feeding ("niu_id");

-- 偷草记录。
CREATE TABLE IF NOT EXISTS plugin_sicau_niu_steal (
    "id"             BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "actor_user_id"  BIGINT NOT NULL DEFAULT 0,
    "target_user_id" BIGINT NOT NULL DEFAULT 0,
    "amount"         INT NOT NULL DEFAULT 0,
    "steal_date"     VARCHAR(10) NOT NULL DEFAULT '',
    "created_at"     TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "updated_at"     TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "deleted_at"     TIMESTAMP
);
COMMENT ON TABLE plugin_sicau_niu_steal IS 'sicau-niu steal-grass record table';
COMMENT ON COLUMN plugin_sicau_niu_steal."steal_date" IS 'Steal date YYYY-MM-DD for the daily count limit';
CREATE INDEX IF NOT EXISTS idx_sicau_niu_steal_actor_date ON plugin_sicau_niu_steal ("actor_user_id", "steal_date");

-- 送草记录。
CREATE TABLE IF NOT EXISTS plugin_sicau_niu_gift (
    "id"           BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "from_user_id" BIGINT NOT NULL DEFAULT 0,
    "to_user_id"   BIGINT NOT NULL DEFAULT 0,
    "amount"       INT NOT NULL DEFAULT 0,
    "gift_date"    VARCHAR(10) NOT NULL DEFAULT '',
    "created_at"   TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "updated_at"   TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "deleted_at"   TIMESTAMP
);
COMMENT ON TABLE plugin_sicau_niu_gift IS 'sicau-niu gift-grass record table';
COMMENT ON COLUMN plugin_sicau_niu_gift."gift_date" IS 'Gift date YYYY-MM-DD for the daily count limit';
CREATE INDEX IF NOT EXISTS idx_sicau_niu_gift_from_date ON plugin_sicau_niu_gift ("from_user_id", "gift_date");

-- 站内信。
CREATE TABLE IF NOT EXISTS plugin_sicau_niu_inbox_msg (
    "id"         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "user_id"    BIGINT NOT NULL DEFAULT 0,
    "msg_type"   VARCHAR(20) NOT NULL DEFAULT '',
    "content"    VARCHAR(500) NOT NULL DEFAULT '',
    "is_read"    INT NOT NULL DEFAULT 0,
    "created_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "deleted_at" TIMESTAMP
);
COMMENT ON TABLE plugin_sicau_niu_inbox_msg IS 'sicau-niu in-app message (inbox) table';
COMMENT ON COLUMN plugin_sicau_niu_inbox_msg."msg_type" IS 'Message type: stolen, gift_received';
COMMENT ON COLUMN plugin_sicau_niu_inbox_msg."is_read" IS 'Read flag: 1 read, 0 unread';
CREATE INDEX IF NOT EXISTS idx_sicau_niu_inbox_user ON plugin_sicau_niu_inbox_msg ("user_id");
