-- +migrate Up notransaction
-- A task's output becomes JSONB: read by containment - the autoscale's runs, by
-- the apps they scaled - it is not parsed again on every read, and is indexed.
-- JSONB refuses a \u0000 that JSON took: an output holding one has it dropped,
-- and one that would still not convert is let go. Each statement may run again
-- after one that failed: the column is read as text, JSON or JSONB.

-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION hp_text_to_jsonb_lenient(v text) RETURNS jsonb LANGUAGE plpgsql AS $$
BEGIN
    RETURN v::jsonb;
EXCEPTION WHEN OTHERS THEN
    BEGIN
        RETURN regexp_replace(v, '\\u0000', '', 'g')::jsonb;
    EXCEPTION WHEN OTHERS THEN
        RETURN NULL;
    END;
END
$$;
-- +migrate StatementEnd

ALTER TABLE tasks ALTER COLUMN output TYPE JSONB USING hp_text_to_jsonb_lenient(output::text);

DROP FUNCTION hp_text_to_jsonb_lenient(text);

-- The periodic jobs' runs by their output: the autoscale's events of an app
-- (appautoscaleserviceimpl, Events) ask by these conditions, and must.
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_tasks_periodic_output ON tasks USING gin (output jsonb_path_ops)
    WHERE type = 'task:periodic-exec' AND deleted_at IS NULL;

-- +migrate Down notransaction
DROP INDEX CONCURRENTLY IF EXISTS idx_tasks_periodic_output;
ALTER TABLE tasks ALTER COLUMN output TYPE JSON USING output::json;
