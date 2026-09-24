ALTER TABLE plugin_sicau_niu_iron
    ADD COLUMN IF NOT EXISTS "bonus_enabled" BOOLEAN NOT NULL DEFAULT FALSE;

COMMENT ON COLUMN plugin_sicau_niu_iron."bonus_enabled" IS 'Whether this iron cow may grant the feeding proximity bonus';
