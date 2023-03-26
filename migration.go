package main

import (
	"gorm.io/gorm"
	"time"
)

type Log struct {
	ID                uint      `gorm:"primaryKey"`
	Timestamp         time.Time `gorm:"index"`
	HttpMethod        string
	RequestUrl        string
	RequestBody       string
	RequestHeaders    string
	ResponseBody      string
	ResponseHeaders   string
	StatusCode        int
	ErrorMessage      string
	UserAgent         string
	IPAddress         string
	Duration          time.Duration
	RequestTimestamp  time.Time
	ResponseTimestamp time.Time
	RequestSize       int64
	ResponseSize      int64
	RequestId         string
}

func (Log) TableName() string {
	return "logs"
}

func Up(db *gorm.DB) error {
	err := db.AutoMigrate(&Log{})
	if err != nil {
		return err
	}
	return nil
}
