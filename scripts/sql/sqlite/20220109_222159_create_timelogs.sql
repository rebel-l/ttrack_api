-- up
CREATE TABLE IF NOT EXISTS timelogs (
    id CHAR(36) NOT NULL PRIMARY KEY,
    start DATETIME NOT NULL,
    stop DATETIME,
    reason VARCHAR(30) NOT NULL DEFAULT 'work',
    location VARCHAR(20) NOT NULL DEFAULT 'home',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modified_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER IF NOT EXISTS timelogs_after_update AFTER UPDATE ON timelogs BEGIN
    UPDATE timelogs SET modified_at = DATETIME('now') WHERE id = NEW.id;
end;


-- down
DROP TRIGGER IF EXISTS timelogs_after_update;

DROP TABLE IF EXISTS timelogs;
