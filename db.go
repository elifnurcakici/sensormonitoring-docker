package main

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var gormDB *gorm.DB

func connectDB() *gorm.DB {
	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	port := os.Getenv("DB_PORT")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		host, user, password, dbname, port)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{}) // gorm.Open fonksiyonu ile veritabanına bağlanmaya çalışıyor
	if err != nil {
		log.Fatalf("Veritabanına bağlanılamadı: %v", err) // Eğer bağlantı başarısız olursa
	}

	// Tablolar otomatik oluşturulur
	db.AutoMigrate(&User{}, &TemperatureReading{}, &HumidityReading{}, &PressureReading{})
	//Eğer tablo zaten varsa da eksik sütunları tamamlar.

	fmt.Println("PostgreSQL bağlantısı başarılı!")
	return db
}
