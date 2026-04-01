package service

import (
	"github.com/sky-night-net/snet/database"
	"github.com/sky-night-net/snet/database/model"
	"github.com/sky-night-net/snet/util/crypto"
)

type UserService struct {
}

func (s *UserService) Login(username string, password string) (*model.User, error) {
	db := database.GetDB()
	var user model.User
	err := db.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	if !crypto.CheckPasswordHash(user.Password, password) {
		return nil, nil
	}
	return &user, nil
}

func (s *UserService) GetFirstUser() (*model.User, error) {
	db := database.GetDB()
	var user model.User
	err := db.First(&user).Error
	return &user, err
}

func (s *UserService) UpdateFirstUser(username string, password string) error {
	db := database.GetDB()
	var user model.User
	if err := db.First(&user).Error; err != nil {
		return err
	}
	if username != "" {
		user.Username = username
	}
	if password != "" {
		hashed, err := crypto.HashPasswordAsBcrypt(password)
		if err != nil {
			return err
		}
		user.Password = hashed
	}
	return db.Save(&user).Error
}
