-- Align CAS time point columns with the host schema's timezone-aware timestamp convention.
-- birthday is date-like profile data and is intentionally left unchanged.
DO $$
DECLARE
    target record;
BEGIN
    FOR target IN
        SELECT c.table_name, c.column_name
        FROM information_schema.columns AS c
        WHERE c.table_schema = current_schema()
          AND c.table_name LIKE 'plugin_linapro_uidentity_cas_%'
          AND c.data_type = 'timestamp without time zone'
          AND c.column_name <> 'birthday'
    LOOP
        EXECUTE format(
            'ALTER TABLE %I ALTER COLUMN %I TYPE TIMESTAMPTZ USING %I AT TIME ZONE current_setting(''TIMEZONE'')',
            target.table_name,
            target.column_name,
            target.column_name
        );
    END LOOP;
END $$;
