package service

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/sky-night-net/snet/logger"
)

type BackupService struct {
	basePath string
}

func NewBackupService() *BackupService {
	return &BackupService{
		basePath: "/etc/snet",
	}
}

func (s *BackupService) CreateBackup() (string, error) {
	timestamp := time.Now().Format("20060102150405")
	backupPath := filepath.Join(os.TempDir(), fmt.Sprintf("snet_backup_%s.zip", timestamp))

	logger.Infof("Creating system backup at %s", backupPath)

	zipFile, err := os.Create(backupPath)
	if err != nil {
		return "", err
	}
	defer zipFile.Close()

	archive := zip.NewWriter(zipFile)
	defer archive.Close()

	// 1. Backup DB
	err = s.addToZip(archive, filepath.Join(s.basePath, "snet.db"), "snet.db")
	if err != nil {
		logger.Warning("Database not found for backup, skipping")
	}

	// 2. Backup Configs
	filepath.Walk(filepath.Join(s.basePath, "amneziawg"), func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			rel, _ := filepath.Rel(s.basePath, path)
			s.addToZip(archive, path, rel)
		}
		return nil
	})

	return backupPath, nil
}

func (s *BackupService) addToZip(w *zip.Writer, srcPath, destPath string) error {
	file, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer file.Close()

	info, _ := file.Stat()
	header, _ := zip.FileInfoHeader(info)
	header.Name = destPath
	header.Method = zip.Deflate

	writer, err := w.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, file)
	return err
}
