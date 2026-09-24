INSERT INTO plugin_sicau_niu_niu ("code", "niu_type", "lat", "lng", "online_at", "status")
SELECT
    'NIU-' || lpad(ordinal::text, 3, '0'),
    'common',
    30.70450 + ((ordinal - 1) / 4) * 0.00090,
    103.86110 + ((ordinal - 1) % 4) * 0.00140,
    CURRENT_TIMESTAMP,
    'inactive'
FROM generate_series(1, 12) AS cattle(ordinal)
ON CONFLICT DO NOTHING;
