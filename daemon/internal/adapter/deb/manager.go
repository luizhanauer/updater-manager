package deb

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type AptManager struct{}

func NewAptManager() *AptManager {
	return &AptManager{}
}

func (m *AptManager) GetInstalledVersion(pkgName string) (string, error) {
	// A mágica: ${db:Status-Status} retorna 'installed' ou 'config-files' (rc)
	// Formato saída: STATUS|VERSAO
	cmd := exec.Command("dpkg-query", "-W", "-f=${db:Status-Status}|${Version}", pkgName)
	out, err := cmd.Output()

	if err != nil {
		return "", nil
	} // Não encontrado

	parts := strings.Split(strings.TrimSpace(string(out)), "|")
	if len(parts) != 2 {
		return "", nil
	}

	status := parts[0]
	version := parts[1]

	// Só consideramos instalado se o status for explicitamente 'installed'
	if status == "installed" {
		return version, nil
	}

	// Se for 'config-files' (rc) ou qualquer outra coisa, retorna vazio (não instalado)
	return "", nil
}

func (m *AptManager) Remove(pkgName string) error {
	// --purge remove até os arquivos de configuração (limpa o estado rc)
	cmd := exec.Command("apt", "purge", "-y", pkgName)
	cmd.Env = append(os.Environ(), "DEBIAN_FRONTEND=noninteractive")
	return cmd.Run()
}

func (m *AptManager) Install(appName, url, checksum string, progress func(int64, int64)) error {
	tmpPath := filepath.Join(os.TempDir(), appName+".deb")
	defer os.Remove(tmpPath)

	if err := downloadFile(tmpPath, url, progress); err != nil {
		return err
	}

	if checksum != "" {
		if err := verifyChecksum(tmpPath, checksum); err != nil {
			return err
		}
	}

	cmd := exec.Command("apt",
		"-o", "Dpkg::Lock::Timeout=60",
		"install", "-y",
		"--allow-downgrades",
		tmpPath,
	)
	cmd.Env = append(os.Environ(), "DEBIAN_FRONTEND=noninteractive")

	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("apt failed: %v | Log: %s", err, string(out))
	}
	return nil
}

// Helpers
func downloadFile(path, url string, progress func(int64, int64)) error {
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	counter := &writeCounter{Total: resp.ContentLength, OnProgress: progress}
	if _, err = io.Copy(out, io.TeeReader(resp.Body, counter)); err != nil {
		return err
	}
	return nil
}

type writeCounter struct {
	Total      int64
	Downloaded int64
	OnProgress func(int64, int64)
}

func (wc *writeCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.Downloaded += int64(n)
	if wc.OnProgress != nil {
		wc.OnProgress(wc.Downloaded, wc.Total)
	}
	return n, nil
}

func verifyChecksum(path, expected string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}
	if hex.EncodeToString(h.Sum(nil)) != strings.TrimPrefix(expected, "sha256:") {
		return fmt.Errorf("checksum mismatch")
	}
	return nil
}
