package main

import "time"

type User struct {
	ID          uint   `gorm:"primaryKey"`
	Username    string `gorm:"unique;not null"`
	Password    string `gorm:"not null"`
	Role        string `gorm:"default:user"`
	WorkspaceID uint   `gorm:"not null"`
}

type Sensor struct {
	ID          uint   `gorm:"primaryKey"`
	Name        string `gorm:"unique;not null"`
	Type        string `gorm:"not null"`
	WorkspaceID uint   `gorm:"not null"`
}

type Workspace struct {
	ID      uint `gorm:"primaryKey"`
	Name    string
	Sensors []Sensor `gorm:"foreignKey:WorkspaceID"`
}

type SensorReading struct {
	ID        uint      `gorm:"primaryKey"`
	SensorID  uint      `gorm:"not null;index"`
	Sensor    Sensor    `gorm:"foreignKey:SensorID;references:ID"`
	Value     float64   `gorm:"not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}
type LoginLog struct {
	ID       uint `gorm:"primaryKey"`
	UserID   uint
	User     User `gorm:"foreignKey:UserID"`
	LoginAt  time.Time
	LogoutAt *time.Time // NULL olabilir
}
