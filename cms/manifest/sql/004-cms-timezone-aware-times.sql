-- Align CMS time point columns with the host schema's timezone-aware timestamp convention.
DO $$
DECLARE
    target record;
BEGIN
    FOR target IN
        SELECT c.table_name, c.column_name
        FROM information_schema.columns AS c
        WHERE c.table_schema = current_schema()
          AND c.table_name LIKE 'plugin_cms_%'
          AND c.data_type = 'timestamp without time zone'
    LOOP
        EXECUTE format(
            'ALTER TABLE %I ALTER COLUMN %I TYPE TIMESTAMPTZ USING %I AT TIME ZONE current_setting(''TIMEZONE'')',
            target.table_name,
            target.column_name,
            target.column_name
        );
    END LOOP;
END $$;
