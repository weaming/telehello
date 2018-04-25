-- +migrate Up
CREATE TABLE IF NOT EXISTS reminders(
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  content TEXT,
  chat_id INTEGER,
  due_dt DATETIME,
  created DATETIME
);

-- +migrate Down
DROP TABLE reminders;