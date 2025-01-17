package database

import (
	"fmt"
	"sync"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	DB   *gorm.DB
	once sync.Once
)

var config = struct {
	User     string
	Password string
	Host     string
	Port     string
	Database string
}{
	User:     "root",
	Password: "",
	Host:     "127.0.0.1",
	Port:     "3306",
	Database: "ac",
}

// InitMySQL initializes the MySQL database connection using GORM
func InitMySQL() error {
	var initErr error
	once.Do(func() {
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=UTC",
			config.User,
			config.Password,
			config.Host,
			config.Port,
			config.Database,
		)

		db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err != nil {
			initErr = fmt.Errorf("failed to connect to MySQL, err: %w", err)
			return
		}
		DB = db
	})
	return initErr
}
