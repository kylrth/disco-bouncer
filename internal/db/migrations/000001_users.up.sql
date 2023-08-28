CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    name_key_hash TEXT NOT NULL,
    finish_year INT NOT NULL,
    professor BOOLEAN DEFAULT FALSE,
    ta BOOLEAN DEFAULT FALSE,
    student_leadership BOOLEAN DEFAULT FALSE,
    alumni_board BOOLEAN DEFAULT FALSE
);
