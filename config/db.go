package config

import (
	"fmt"
	"log"
	"time"
	_ "time/tzdata"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func GetDB() *gorm.DB {
	fmt.Println("Initializing DB")
	sslMode := "disable"
	if DBSslMode == 1 {
		sslMode = "enable"
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
		DBHost, DBUser, DBPassword, DBName, DBPort, sslMode, DBTimezone,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{TranslateError: true})
	if err != nil {
		log.Fatalf("Failed to connect database. Error: %s", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to create connection pool. Error: %s ", err.Error())
	}

	sqlDB.SetMaxIdleConns(DBPoolMaxIdleConn)

	sqlDB.SetMaxOpenConns(DBPoolMaxOpenConn)

	sqlDB.SetConnMaxLifetime(time.Hour)

	return db
}
