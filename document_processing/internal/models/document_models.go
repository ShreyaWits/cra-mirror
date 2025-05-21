//go:build !test
// build +test
package models

import "time"

type DocumentData struct {
	Batch_id  string `gorm:"primaryKey"`
	Data      string `gorm:"type:text"`
	Status    string `gorm:"type:text"`
	Message   string `gorm:"type:text"`
	CreatedAt time.Time
}
