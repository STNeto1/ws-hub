package pkg

import (
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/libsql/libsql-client-go/libsql"
	_ "modernc.org/sqlite"
)

func CreateDB() *sqlx.DB {
	conn, err := sqlx.Open("libsql", "file:./database.sqlite3")
	if err != nil {
		log.Fatalln("failed to open connection", err)
	}

	if err := conn.Ping(); err != nil {
		log.Fatalln("failed to ping", err)
	}

	if _, err = conn.Exec(__getSql()); err != nil {
		log.Fatalln("failed to create table rooms", err)
	}

	return conn
}

func __getSql() string {
	return `
        CREATE TABLE IF NOT EXISTS logs (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            topic TEXT NOT NULL,
            message JSON NOT NULL,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP
        );

        CREATE INDEX IF NOT EXISTS idx_logs_topic ON logs (topic);
    `
}
