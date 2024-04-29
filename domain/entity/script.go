package entity

import "time"

type Script struct {
	ID        int       `db:"id"`
	Command   string    `db:"command"`
	Output    string    `db:"output"`
	IsRunning bool      `db:"is_running"`
	PID       int       `db:"pid"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
