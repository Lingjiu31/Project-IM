package repository

import (
	"gorm.io/gorm"
)

type MySQLGroupRepo struct {
	db *gorm.DB
}
