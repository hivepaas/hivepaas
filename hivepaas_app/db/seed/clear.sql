-- Clear all current tables
DO $$ DECLARE
r RECORD;
BEGIN
    -- if the schema you operate on is not "current", you will want to
    -- replace current_schema() in query with 'schematodeletetablesfrom'
    -- *and* update the generate 'DROP...' accordingly.
FOR r IN (SELECT tablename FROM pg_tables WHERE schemaname = current_schema()) LOOP
    IF r.tablename NOT IN ('spatial_ref_sys') THEN -- these tables used by pg_search
        EXECUTE 'DROP TABLE IF EXISTS ' || quote_ident(r.tablename) || ' CASCADE';
    END IF;
END LOOP;
END $$;
