package deb

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
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
	cmd := exec.Command("apt",
		"-o", "Dpkg::Lock::Timeout=60",
		"purge",
		"-y",
		pkgName,
	)
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

	// ETAPA DE SIMULAÇÃO: Verifica se a instalação é viável ANTES de tentar.
	if simErr := simulateInstall(tmpPath); simErr != nil {
		// O erro de simulação já é bem descritivo, então o retornamos diretamente.
		// A camada de serviço (usecase) pode usar esse erro para definir o estado de erro permanente.
		return simErr
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

// SimulationError é um erro estruturado para falhas na simulação de instalação.
// Isso permite que a camada de serviço (usecase) tome decisões inteligentes,
// como não tentar instalar novamente se o erro for de dependência (permanente).
type SimulationError struct {
	// Code é um identificador da falha (ex: "DEP_ERROR", "LOCK_ERROR").
	Code string
	// Message é a mensagem para o usuário.
	Message string
	// Details contém informações adicionais, como a lista de dependências.
	Details []string
}

func (e *SimulationError) Error() string {
	if len(e.Details) > 0 {
		return fmt.Sprintf("%s: %s", e.Message, strings.Join(e.Details, ", "))
	}
	return e.Message
}

// simulateInstall executa 'apt-get install -s' para prever o sucesso da instalação.
func simulateInstall(pkgPath string) *SimulationError {
	// Usamos apt-get aqui por ser mais estável para scripting e parsing de output.
	// O daemon já roda como root, então 'sudo' não é necessário.
	cmd := exec.Command("apt-get", "install", "-s", "-y", "--allow-downgrades", pkgPath)
	cmd.Env = append(os.Environ(), "LANG=C", "LC_ALL=C") // Força inglês para parsing

	outputBytes, err := cmd.CombinedOutput()
	output := string(outputBytes)

	// Cenário 1: Sucesso claro (já instalado ou será instalado)
	isAlreadyInstalled := strings.Contains(output, "is already the newest version")
	willBeInstalled := strings.Contains(output, "The following NEW packages will be installed:") || strings.Contains(output, "The following packages will be upgraded:") || strings.Contains(output, "The following packages will be DOWNGRADED:")

	if isAlreadyInstalled || (willBeInstalled && err == nil) {
		return nil // Simulação bem-sucedida, nenhum erro.
	}

	// Cenário 2: Falha por dependências (Erro Permanente)
	if strings.Contains(output, "Unmet dependencies") || strings.Contains(output, "broken packages") {
		deps := extractMissingDeps(output)
		return &SimulationError{
			Code:    "DEP_ERROR",
			Message: "Dependências não resolvidas",
			Details: deps,
		}
	}

	// Cenário 3: Falha por lock (Erro Temporário)
	if strings.Contains(output, "Could not get lock") {
		return &SimulationError{Code: "LOCK_ERROR", Message: "Gerenciador de pacotes está ocupado"}
	}

	// Cenário 4: Pacote não encontrado
	if strings.Contains(output, "Unable to locate package") {
		return &SimulationError{Code: "NOT_FOUND", Message: "Arquivo do pacote não encontrado ou inválido"}
	}

	// Cenário 5: Falha genérica (baseada no código de saída)
	if err != nil {
		return &SimulationError{Code: "GENERIC_ERROR", Message: "Falha na simulação da instalação"}
	}

	// Se chegou aqui, é uma situação não prevista, mas a simulação não indicou sucesso.
	return &SimulationError{Code: "UNKNOWN_ERROR", Message: "Falha desconhecida na pré-verificação"}
}

// extractMissingDeps usa regex para extrair nomes de dependências ausentes do output do apt.
func extractMissingDeps(output string) []string {
	re := regexp.MustCompile(`Depends:\s+([a-zA-Z0-9\-\.:]+)`)
	matches := re.FindAllStringSubmatch(output, -1)
	deps := make(map[string]bool)
	for _, match := range matches {
		deps[match[1]] = true
	}
	result := make([]string, 0, len(deps))
	for dep := range deps {
		result = append(result, dep)
	}
	return result
}

// Helpers
func downloadFile(path, url string, progress func(int64, int64)) error {
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	// Adiciona um timeout ao cliente HTTP para evitar que o download trave indefinidamente.
	client := http.Client{
		Timeout: 30 * time.Minute, // Timeout generoso para downloads grandes
	}

	req, err := http.NewRequestWithContext(context.Background(), "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
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
