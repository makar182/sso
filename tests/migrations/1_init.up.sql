INSERT INTO apps (id, name, secret, timestamp)
VALUES (1, 'test', 'test-secret', '2023-10-01 00:00:00')
ON CONFLICT DO NOTHING;
