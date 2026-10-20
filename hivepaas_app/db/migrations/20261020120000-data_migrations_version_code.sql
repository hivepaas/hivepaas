-- +migrate Up
-- The row the table was made with names v000001's data. As v1-20251019 it sorted
-- after every version code, and data migration never ran.
UPDATE data_migrations SET id = 'v000001' WHERE id = 'v1-20251019';

-- +migrate Down
UPDATE data_migrations SET id = 'v1-20251019' WHERE id = 'v000001';
