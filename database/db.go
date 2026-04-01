// Package database handles persistence for SNET.
package database

import (
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/sky-night-net/snet/config"
	"github.com/sky-night-net/snet/database/model"
	"github.com/sky-night-net/snet/logger"
	"github.com/sky-night-net/snet/util/crypto"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	glogger "gorm.io/gorm/logger"
)

var (
	db   *gorm.DB
	once sync.Once
)

func InitDB() error {
	var err error
	once.Do(func() {
		dbPath := config.GetDBPath()
		dbDir := filepath.Dir(dbPath)
		if _, err := os.Stat(dbDir); os.IsNotExist(err) {
			os.MkdirAll(dbDir, 0755)
		}

		gconfig := &gorm.Config{}
		if !config.IsDebug() {
			gconfig.Logger = glogger.Default.LogMode(glogger.Silent)
		}

		db, err = gorm.Open(sqlite.Open(dbPath), gconfig)
		if err != nil {
			logger.Errorf("Failed to connect to database: %v", err)
			return
		}

		err = db.AutoMigrate(
			&model.User{},
			&model.Inbound{},
			&model.Client{},
			&model.Setting{},
		)
		if err != nil {
			logger.Errorf("Failed to auto migrate database: %v", err)
			return
		}

		err = initDefaultData()
		if err != nil {
			logger.Errorf("Failed to init default data: %v", err)
			return
		}
	})
	return err
}

func GetDB() *gorm.DB {
	return db
}

func initDefaultData() error {
	var count int64
	db.Model(&model.User{}).Count(&count)
	if count == 0 {
		password, _ := crypto.HashPasswordAsBcrypt("admin")
		user := &model.User{
			Username: "admin",
			Password: password,
		}
		if err := db.Create(user).Error; err != nil {
			return err
		}
		logger.Info("Default admin user created: admin / admin")
	}

	// Initialize default settings if needed
	port := "2053"
	if envPort := config.GetEnvPanelPort(); envPort > 0 {
		port = strconv.Itoa(envPort)
	}

	defaults := map[string]string{
		"web_port":        port,
		"web_base_path":   "/",
		"session_max_age": "3600",
		"xray_bin_path":   config.GetBinFolderPath() + "/xray",
		"server_ip":       config.GetEnvServerIP(),
	}

	for k, v := range defaults {
		var setting model.Setting
		if err := db.Where("key = ?", k).First(&setting).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				db.Create(&model.Setting{Key: k, Value: v})
			}
		}
	}

	return nil
}
