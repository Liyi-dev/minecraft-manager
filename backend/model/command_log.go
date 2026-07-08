package model

import "time"

type CommandLog struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id" gorm:"index;not null"`
	Command   string    `json:"command" gorm:"size:1024;not null"`
	Result    string    `json:"result" gorm:"type:text"`
	CreatedAt time.Time `json:"created_at"`
}

func (CommandLog) TableName() string {
	return "command_logs"
}
