DO $$
DECLARE
    previous_default TEXT;
BEGIN
    SELECT pg_get_expr(default_value.adbin, default_value.adrelid)
    INTO previous_default
    FROM pg_attribute AS column_meta
    JOIN pg_attrdef AS default_value
      ON default_value.adrelid = column_meta.attrelid
     AND default_value.adnum = column_meta.attnum
    WHERE column_meta.attrelid = 'plugin_sicau_niu_iron'::regclass
      AND column_meta.attname = 'bonus_enabled';

    EXECUTE 'ALTER TABLE plugin_sicau_niu_iron ALTER COLUMN bonus_enabled SET DEFAULT TRUE';

    IF previous_default = 'false' THEN
        UPDATE plugin_sicau_niu_iron
        SET bonus_enabled = TRUE
        WHERE bonus_enabled = FALSE AND deleted_at IS NULL;
    END IF;
END $$;
