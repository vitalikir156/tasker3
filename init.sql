CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    uname TEXT NOT NULL,
    taskread boolean,
    taskwrite boolean,
    userread boolean,
    userwrite boolean
);
CREATE TABLE tasks (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    description TEXT,
    status TEXT CHECK (status IN ('new', 'in_progress', 'done')) DEFAULT 'new',
    created_at TIMESTAMP DEFAULT now()
);

