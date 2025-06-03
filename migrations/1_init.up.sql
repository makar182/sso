CREATE TABLE IF NOT EXISTS public.apps
(
    id        SERIAL PRIMARY KEY,
    name      TEXT      NOT NULL UNIQUE,
    secret    TEXT      NOT NULL,
    timestamp TIMESTAMP NOT NULL
);
CREATE TABLE IF NOT EXISTS public.users
(
    id        SERIAL PRIMARY KEY,
    email     TEXT      NOT NULL UNIQUE,
    pass_hash TEXT      NOT NULL,
    is_admin  BOOLEAN   NOT NULL DEFAULT FALSE,
    timestamp TIMESTAMP NOT NULL
);

INSERT INTO apps (name, secret, timestamp)
VALUES ('default', 'default_secret', NOW())