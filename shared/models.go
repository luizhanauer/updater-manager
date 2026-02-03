package shared

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	SocketPath    = "/run/linux-updater/service.sock"
	GoalsFilePath = "/var/lib/updater/goals.json"
)

// --- ENUMS ---
const (
	GoalActive = "active"
	GoalAbsent = "absent"

	StateIdle        = "idle"
	StateDownloading = "downloading"
	StateInstalling  = "installing"
	StateRemoving    = "removing"
	StateError       = "error"
)

// --- DTOs (Data Transfer Objects) ---

// Goals: O que o usuário QUER (Persistido)
type Goals struct {
	Apps map[string]GoalApp `json:"apps"`
}

type GoalApp struct {
	DesiredState string `json:"desired_state"`
	AutoUpdate   bool   `json:"auto_update"`
}

// Status: O que o sistema É (Volátil/Memória)
type Status struct {
	UpdatedAt        time.Time            `json:"updated_at"`
	CatalogCheckedAt time.Time            `json:"catalog_checked_at"` // Campo Novo
	Apps             map[string]AppStatus `json:"apps"`
}

type AppStatus struct {
	State          string `json:"state"`
	Message        string `json:"message"`
	Progress       int    `json:"progress"`
	DownloadedSize int64  `json:"downloaded_size"`
	TotalSize      int64  `json:"total_size"`
	CurrentVersion string `json:"current_version"`
	LatestVersion  string `json:"latest_version"`
	Error          string `json:"error,omitempty"`
	AutoUpdate     bool   `json:"auto_update"`
}

// SSEEvent: Evento enviado via Socket para a UI
type SSEEvent struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// --- CATALOG ENTITIES ---

type RegistryCatalog struct {
	LastUpdated time.Time      `json:"last_updated"`
	Apps        map[string]App `json:"apps"`
}

type App struct {
	ID             string      `json:"id"`
	Name           string      `json:"name"`
	Description    string      `json:"description"`
	Category       string      `json:"category"`
	IconURL        string      `json:"icon_url"`
	PackageName    string      `json:"package_name"`
	InstallType    string      `json:"install_type"`
	CurrentRelease ReleaseInfo `json:"current_release"`
}

type ReleaseInfo struct {
	Version     string `json:"version"`
	DownloadURL string `json:"download_url"`
	Checksum    string `json:"checksum"`
	Size        int64  `json:"size"`
}

// --- REGRAS DE DOMÍNIO ---

func IsVersionCompatible(local, remote string) bool {
	if local == "" || remote == "" {
		return false
	}
	l := strings.TrimPrefix(strings.ToLower(local), "v")
	r := strings.TrimPrefix(strings.ToLower(remote), "v")
	if l == r {
		return true
	}
	if strings.HasPrefix(l, r) {
		return true
	}
	return false
}

// Helpers de Disco para Goals
func LoadGoalsDisk() Goals {
	f, err := os.Open(GoalsFilePath)
	if err != nil {
		return Goals{Apps: make(map[string]GoalApp)}
	}
	defer f.Close()
	var g Goals
	json.NewDecoder(f).Decode(&g)
	if g.Apps == nil {
		g.Apps = make(map[string]GoalApp)
	}
	return g
}

func SaveGoalsDisk(g Goals) error {
	os.MkdirAll(filepath.Dir(GoalsFilePath), 0755)
	f, err := os.Create(GoalsFilePath)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(g)
}
