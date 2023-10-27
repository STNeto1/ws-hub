package pkg

import "encoding/json"

type Log struct {
	ID        int32           `db:"id" json:"id"`
	Topic     string          `db:"topic" json:"topic"`
	Message   json.RawMessage `db:"message" json:"message"`
	CreatedAt string          `db:"created_at" json:"created_at"`
}

func LogPredicate(s1 Log, s2 Log) bool {
	return s1.ID == s2.ID
}
