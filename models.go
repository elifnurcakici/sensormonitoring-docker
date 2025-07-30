package main

import "time"

type User struct {
	ID       uint   `gorm:"primaryKey"` // bu alan otomatik artan şekilde olacak
	Username string `gorm:"unique;not null"`
	Password string `gorm:"not null"`
	Role     string `gorm:"default:user"` // admin veya user
}

type TemperatureReading struct {
	ID        uint      `gorm:"primaryKey"`
	Value     float64   `gorm:"not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

type HumidityReading struct {
	ID        uint      `gorm:"primaryKey"`
	Value     float64   `gorm:"not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

type PressureReading struct {
	ID        uint      `gorm:"primaryKey"`
	Value     float64   `gorm:"not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}
