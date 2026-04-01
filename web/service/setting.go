package service

import (
	"strconv"
	"time"

	"github.com/sky-night-net/snet/database"
	"github.com/sky-night-net/snet/database/model"
)

type SettingService struct {
}

func (s *SettingService) GetSetting(key string) (string, error) {
	db := database.GetDB()
	var setting model.Setting
	err := db.Where("key = ?", key).First(&setting).Error
	if err != nil {
		return "", err
	}
	return setting.Value, nil
}

func (s *SettingService) SetSetting(key string, value string) error {
	db := database.GetDB()
	return db.Where("key = ?", key).Assign(model.Setting{Value: value}).FirstOrCreate(&model.Setting{Key: key}).Error
}

func (s *SettingService) GetPort() (int, error) {
	val, err := s.GetSetting("web_port")
	if err != nil {
		return 2053, nil
	}
	return strconv.Atoi(val)
}

func (s *SettingService) GetBasePath() (string, error) {
	val, err := s.GetSetting("web_base_path")
	if err != nil {
		return "/", nil
	}
	return val, nil
}

func (s *SettingService) GetSecret() ([]byte, error) {
	val, err := s.GetSetting("web_secret")
	if err != nil {
		return []byte("snet-default-secret-key-12345"), nil
	}
	return []byte(val), nil
}

func (s *SettingService) GetSessionMaxAge() (int, error) {
	val, err := s.GetSetting("session_max_age")
	if err != nil {
		return 3600, nil
	}
	return strconv.Atoi(val)
}

func (s *SettingService) GetTimeLocation() (*time.Location, error) {
	// For now default to UTC, can be extended to settings
	return time.UTC, nil
}
