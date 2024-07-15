ALTER TABLE users ADD COLUMN temp_finish_year INTEGER;

UPDATE users
SET temp_finish_year = CASE
    WHEN finish_year = '' THEN -1
    ELSE finish_year::INTEGER
END;

ALTER TABLE users DROP COLUMN finish_year;
ALTER TABLE users RENAME COLUMN temp_finish_year TO finish_year;

ALTER TABLE users ALTER COLUMN finish_year SET NOT NULL;
