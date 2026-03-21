CREATE TABLE IF NOT EXISTS devices (
    id         INTEGER     PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    vendor     TEXT        NOT NULL,
    ip         TEXT        NOT NULL,
    port       INTEGER     NOT NULL,
    name       TEXT        NOT NULL DEFAULT '',
    UNIQUE (ip, port)
);
