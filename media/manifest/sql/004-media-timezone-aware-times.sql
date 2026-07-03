-- Align media time point columns with the host schema's timezone-aware timestamp convention.
-- Column names are preserved for media compatibility, including create_time and update_time.
DO $$
DECLARE
    target record;
BEGIN
    FOR target IN
        SELECT c.table_name, c.column_name
        FROM information_schema.columns AS c
        WHERE c.table_schema = current_schema()
          AND c.table_name IN (
              'media_strategy',
              'media_stream_alias',
              'media_tenant_white',
              'media_tenant_stream_config',
              'media_node',
              'media_report_node',
              'media_report_node_snapshot',
              'media_report_instance',
              'media_report_stream',
              'media_report_session'
          )
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
