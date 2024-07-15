ALTER TABLE users ADD COLUMN temp_finish_year TEXT;

UPDATE users
SET temp_finish_year = CASE
    WHEN finish_year = 0 THEN ''
    WHEN finish_year = -1 THEN ''
    ELSE finish_year::TEXT
END;

ALTER TABLE users DROP COLUMN finish_year;
ALTER TABLE users RENAME COLUMN temp_finish_year TO finish_year;

ALTER TABLE users ALTER COLUMN finish_year SET NOT NULL;
