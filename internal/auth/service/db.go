package service

import (
	"github.com/yourusername/api_ricky_and_morty/internal/auth/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() error {
	db, err := gorm.Open(sqlite.Open("/app/data/users.db"), &gorm.Config{})
	if err != nil {
		return err
	}
	db.AutoMigrate(&models.User{})
	DB = db
	return nil
}

func CreateUser(username, password string) error {
	user := models.User{Username: username, Password: password}
	return DB.Create(&user).Error
}

func GetUserByUsername(username string) (*models.User, error) {
	var user models.User
	err := DB.Where("username = ?", username).First(&user).Error
	return &user, err
}
