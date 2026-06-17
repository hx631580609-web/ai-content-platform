package database

import (
	"fmt"
	"log"

	"ai-content-platform/config"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

var DB *gorm.DB

func ConnectDatabase() {
	cfg := config.LoadConfig()
	
	var err error
	
	// 首先尝试连接MySQL
	mysqlDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)
	
	DB, err = gorm.Open("mysql", mysqlDSN)
	if err != nil {
		// 如果MySQL连接失败，尝试PostgreSQL
		log.Printf("MySQL连接失败: %v, 尝试PostgreSQL", err)
		postgresDSN := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)
		DB, err = gorm.Open("postgres", postgresDSN)
	}
	
	if err != nil {
		// 如果仍然失败，尝试SQLite
		log.Printf("PostgreSQL连接失败: %v, 尝试SQLite", err)
		DB, err = gorm.Open("sqlite3", "./example_db.db")
	}
	
	if err != nil {
		// 如果仍然失败，使用内存存储
		log.Printf("数据库连接失败: %v", err)
		log.Println("注意：由于数据库连接失败，系统将使用内存存储模式运行")
		log.Println("数据仅保存在内存中，重启后会丢失")
		MemoryStore = NewInMemoryStore()
		UseMemory = true
		return
	}

	// Test the connection
	if err := DB.DB().Ping(); err != nil {
		log.Printf("数据库ping失败: %v, 使用内存存储模式", err)
		MemoryStore = NewInMemoryStore()
		UseMemory = true
		return
	}

	log.Println("Successfully connected to database")
}

func CloseDatabase() {
	if DB != nil {
		DB.Close()
	}
}