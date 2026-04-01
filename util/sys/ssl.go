package sys

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sky-night-net/snet/logger"
)

type SSLManager struct {
	basePath string
}

func NewSSLManager() *SSLManager {
	return &SSLManager{
		basePath: "/etc/snet/ssl",
	}
}

func (m *SSLManager) IssueCertificate(domain string, email string) error {
	logger.Infof("Requesting SSL certificate for domain %s", domain)
	
	if err := os.MkdirAll(m.basePath, 0755); err != nil {
		return err
	}

	// 1. Install acme.sh if not present
	if _, err := Execute("acme.sh --version"); err != nil {
		logger.Info("Installing acme.sh...")
		Execute(fmt.Sprintf("curl https://get.acme.sh | sh -s email=%s", email))
	}

	// 2. Issue certificate via standalone mode (temporarily stop web server)
	// In a real panel, we would use DNS challenge or --webroot
	cmd := fmt.Sprintf("~/.acme.sh/acme.sh --issue -d %s --standalone --server letsencrypt", domain)
	
	_, err := Execute(cmd)
	if err != nil {
		return fmt.Errorf("certificate issuance failed: %v", err)
	}

	// 3. Install certificate to our path
	installCmd := fmt.Sprintf("~/.acme.sh/acme.sh --install-cert -d %s --key-file %s --fullchain-file %s",
		domain, filepath.Join(m.basePath, "private.key"), filepath.Join(m.basePath, "fullchain.crt"))
	
	_, err = Execute(installCmd)
	return err
}

func (m *SSLManager) IsCertExists() bool {
    _, err := os.Stat(filepath.Join(m.basePath, "fullchain.crt"))
    return err == nil
}

func (m *SSLManager) GetCertPath() (string, string) {
    return filepath.Join(m.basePath, "fullchain.crt"), filepath.Join(m.basePath, "private.key")
}
