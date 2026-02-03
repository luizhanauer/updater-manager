package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/luizhanauer/linux-updater/shared"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx        context.Context
	sockClient *http.Client
}

func NewApp() *App {
	return &App{
		sockClient: &http.Client{
			Timeout: 2 * time.Second,
			Transport: &http.Transport{
				DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
					return net.Dial("unix", shared.SocketPath)
				},
			},
		},
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	go a.listenSSE()
}

func (a *App) listenSSE() {
	for {
		a.connectAndReadSSE()
		time.Sleep(2 * time.Second)
	}
}

func (a *App) connectAndReadSSE() {
	resp, err := a.sockClient.Get("http://unix/events")
	if err != nil {
		return
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data: ") {
			jsonStr := strings.TrimPrefix(line, "data: ")
			var event shared.SSEEvent
			if err := json.Unmarshal([]byte(jsonStr), &event); err == nil {
				runtime.EventsEmit(a.ctx, "daemon_event", event)
			}
		}
	}
}

// --- UI Models ---

type UIApp struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	Category       string `json:"category"`
	IconURL        string `json:"icon_url"`
	Status         string `json:"status"`
	Message        string `json:"message"`
	IsInstalled    bool   `json:"is_installed"`
	AutoUpdate     bool   `json:"auto_update"`
	Progress       int    `json:"progress"`
	DownloadedSize string `json:"downloaded_size"`
	TotalSize      string `json:"total_size"` // Agora sempre preenchido
	LocalVersion   string `json:"local_version"`
	RemoteVersion  string `json:"remote_version"`
}

func (a *App) GetApps() ([]UIApp, error) {
	catalog := shared.GetCatalog(false)

	resp, err := a.sockClient.Get("http://unix/status")
	status := shared.Status{Apps: make(map[string]shared.AppStatus)}
	if err == nil {
		defer resp.Body.Close()
		json.NewDecoder(resp.Body).Decode(&status)
	}

	var result []UIApp
	for id, info := range catalog {
		st := status.Apps[id]

		ui := UIApp{
			ID:            id,
			Name:          info.Name,
			Description:   info.Description,
			Category:      info.Category,
			IconURL:       info.IconURL,
			Status:        st.State,
			Message:       st.Message,
			Progress:      st.Progress,
			RemoteVersion: info.CurrentRelease.Version,
			LocalVersion:  st.CurrentVersion,
			AutoUpdate:    st.AutoUpdate,
			// CORREÇÃO: Usa o tamanho do catálogo por padrão
			TotalSize: formatBytes(info.CurrentRelease.Size),
		}

		// Se estiver baixando, sobrescreve com dados reais do progresso
		if st.TotalSize > 0 {
			ui.DownloadedSize = formatBytes(st.DownloadedSize)
			ui.TotalSize = formatBytes(st.TotalSize)
		}

		if ui.Status == "" {
			ui.Status = "not_installed"
		}
		if st.CurrentVersion != "" {
			ui.IsInstalled = true
		}

		result = append(result, ui)
	}
	return result, nil
}

func (a *App) ToggleInstall(appID string, install bool) string {
	goals := shared.Goals{Apps: map[string]shared.GoalApp{}}
	state := shared.GoalAbsent
	auto := false
	if install {
		state = shared.GoalActive
		auto = true
	}
	goals.Apps[appID] = shared.GoalApp{DesiredState: state, AutoUpdate: auto}
	return a.sendGoals(goals)
}

func (a *App) ToggleAutoUpdate(appID string, enable bool) string {
	goals := shared.Goals{Apps: map[string]shared.GoalApp{}}
	goals.Apps[appID] = shared.GoalApp{DesiredState: shared.GoalActive, AutoUpdate: enable}
	return a.sendGoals(goals)
}

func (a *App) RefreshCatalog() string {
	resp, err := a.sockClient.Post("http://unix/catalog/refresh", "application/json", nil)
	if err != nil {
		return err.Error()
	}
	defer resp.Body.Close()
	return "ok"
}

func (a *App) sendGoals(g shared.Goals) string {
	data, _ := json.Marshal(g)
	resp, err := a.sockClient.Post("http://unix/goals", "application/json", bytes.NewReader(data))
	if err != nil {
		return err.Error()
	}
	defer resp.Body.Close()
	return "ok"
}

// Helper para formatar bytes (KB, MB, GB)
func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
