
CREATE TABLE IF NOT EXISTS devices (
    id         INTEGER     PRIMARY KEY,
    vendor     TEXT        NOT NULL,
    ip         TEXT        NOT NULL,
    port       INTEGER     NOT NULL,
    name       TEXT        NOT NULL,
    UNIQUE (ip, port)
);
