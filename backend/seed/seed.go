package seed

import (
	"errors"

	"gorm.io/gorm"

	"minecraft-manager/model"
	"minecraft-manager/service"
)

const (
	defaultAdminUsername = "admin"
	defaultAdminPassword = "admin123"
	// legacyBadHash was incorrectly documented as admin123 but never matched it.
	legacyBadHash = "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"
)

// DefaultAdmin ensures the default admin account exists with the documented password.
func DefaultAdmin(db *gorm.DB) error {
	hash, err := service.HashPassword(defaultAdminPassword)
	if err != nil {
		return err
	}

	var user model.User
	err = db.Where("username = ?", defaultAdminUsername).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return db.Create(&model.User{
			Username: defaultAdminUsername,
			Password: hash,
			Role:     "admin",
		}).Error
	}
	if err != nil {
		return err
	}

	if user.Password == legacyBadHash {
		user.Password = hash
		return db.Save(&user).Error
	}

	return nil
}
