package database

import (
	"gorm.io/gorm"
)

func GormNoLoggerConfig() *gorm.Config {
	return &gorm.Config{
		Logger: nil,
	}
}
