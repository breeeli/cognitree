package entity

import "time"

type QAPair struct {
	ID        string
	NodeID    string
	Question  string
	CreatedAt time.Time
	Blocks    []Block
}
