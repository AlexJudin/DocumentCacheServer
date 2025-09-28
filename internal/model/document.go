package model

import (
	"time"

	"github.com/lib/pq"
)

const (
	MongoDbName         = "documents"
	MongoCollectionName = "json_files"
)

type MetaDocument struct {
	ID        uint           `gorm:"primarykey" json:"-"`
	UUID      string         `json:"-"`
	CreatedAt time.Time      `json:"-"`
	Name      string         `json:"name"`
	File      bool           `json:"file"`
	Public    bool           `json:"public"`
	Mime      string         `json:"mime"`
	Grant     pq.StringArray `gorm:"type:text[]" json:"grant"`
	FilePath  string         `json:"-"`
}
