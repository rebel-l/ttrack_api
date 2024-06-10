-- up
CREATE TABLE IF NOT EXISTS publicholidays (
    id CHAR(36) NOT NULL PRIMARY KEY,
    day DATETIME NOT NULL,
    name VARCHAR(250) NOT NULL,
    halfday INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modified_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER IF NOT EXISTS publicholidays_after_update AFTER UPDATE ON publicholidays BEGIN
    UPDATE publicholidays SET modified_at = DATETIME('now') WHERE id = NEW.id;
end;


-- down
DROP TRIGGER IF EXISTS publicholidays_after_update;

DROP TABLE IF EXISTS publicholidays;
