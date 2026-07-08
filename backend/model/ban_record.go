package model

import "time"

type BanRecord struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	UUID       string    `json:"uuid" gorm:"size:64;index;not null"`
	PlayerName string    `json:"player_name" gorm:"size:64"`
	Reason     string    `json:"reason" gorm:"size:512"`
	CreatedAt  time.Time `json:"created_at"`
}

func (BanRecord) TableName() string {
	return "ban_records"
}
