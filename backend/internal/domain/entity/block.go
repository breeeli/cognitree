package entity

import "time"

type BlockType string

const (
	BlockTypeParagraph BlockType = "paragraph"
	BlockTypeList      BlockType = "list"
	BlockTypeCode      BlockType = "code"
	BlockTypeQuote     BlockType = "quote"
	BlockTypeHeading   BlockType = "heading"
)

type Block struct {
	ID        string
	QAPairID  string
	Type      BlockType
	Content   string
	CreatedAt time.Time
}
