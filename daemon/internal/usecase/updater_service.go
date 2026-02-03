package usecase

import (
	"runtime/debug"
	"time"

	"github.com/luizhanauer/linux-updater/daemon/internal/domain"
	"github.com/luizhanauer/linux-updater/shared"
)

type UpdaterService struct {
	pkgManager   domain.PackageManager
	stateRepo    domain.StateRepository
	eventBus     domain.EventBus
	installQueue chan string
}

func NewUpdaterService(pkg domain.PackageManager, state domain.StateRepository, bus domain.EventBus) *UpdaterService {
	svc := &UpdaterService{
		pkgManager:   pkg,
		stateRepo:    state,
		eventBus:     bus,
		installQueue: make(chan string, 100),
	}
	return svc
}

func (s *UpdaterService) ForceCatalogRefresh() {
	s.Reconcile(true)
}

func (s *UpdaterService) Reconcile(forceCatalog bool) {
	goals := shared.LoadGoalsDisk()
	catalog := shared.GetCatalog(forceCatalog)

	s.stateRepo.SetCatalogCheckTime(shared.GetLastFetch())

	for appID, appInfo := range catalog {
		pkgName := appInfo.PackageName
		if pkgName == "" {
			pkgName = appID
		}

		// 1. Atualiza Status
		ver, _ := s.pkgManager.GetInstalledVersion(pkgName)

		s.stateRepo.UpdateAppStatus(appID, func(st *shared.AppStatus) {
			st.CurrentVersion = ver
			st.LatestVersion = appInfo.CurrentRelease.Version

			if g, ok := goals.Apps[appID]; ok {
				st.AutoUpdate = g.AutoUpdate
			}

			isActive := st.State == shared.StateDownloading || st.State == shared.StateInstalling
			if !isActive {
				if ver != "" {
					st.State = shared.StateIdle
					st.Message = "Instalado"
				} else {
					st.State = "not_installed"
					st.Message = ""
				}
			}
		})

		goal, hasGoal := goals.Apps[appID]
		if !hasGoal {
			continue
		}

		// 2. Lógica
		status := s.stateRepo.GetStatus().Apps[appID]
		isActive := status.State == shared.StateDownloading || status.State == shared.StateInstalling

		if goal.DesiredState == shared.GoalActive {
			needsUpdate := !shared.IsVersionCompatible(ver, appInfo.CurrentRelease.Version)

			if (ver == "" || (needsUpdate && goal.AutoUpdate)) && !isActive {
				go s.startInstall(appID, appInfo)
			}
		} else if goal.DesiredState == shared.GoalAbsent && ver != "" {
			if !isActive && status.State != shared.StateRemoving {
				s.stateRepo.UpdateAppStatus(appID, func(st *shared.AppStatus) {
					st.State = shared.StateRemoving
					st.Message = "Removendo..."
				})
				s.eventBus.Broadcast(shared.SSEEvent{Type: "status_update", Payload: s.stateRepo.GetStatus()})

				s.pkgManager.Remove(pkgName)

				delete(goals.Apps, appID)
				shared.SaveGoalsDisk(goals)

				s.stateRepo.UpdateAppStatus(appID, func(st *shared.AppStatus) {
					st.State = "not_installed"
					st.Message = "Removido"
					st.CurrentVersion = ""
					st.AutoUpdate = false
				})
				s.eventBus.Broadcast(shared.SSEEvent{Type: "status_update", Payload: s.stateRepo.GetStatus()})

				debug.FreeOSMemory()
			}
		}
	}
	s.eventBus.Broadcast(shared.SSEEvent{Type: "status_update", Payload: s.stateRepo.GetStatus()})
}

func (s *UpdaterService) startInstall(appID string, app shared.App) {
	defer debug.FreeOSMemory()

	s.stateRepo.UpdateAppStatus(appID, func(st *shared.AppStatus) {
		st.State = shared.StateDownloading
		st.Message = "Baixando..."
	})
	s.eventBus.Broadcast(shared.SSEEvent{Type: "status_update", Payload: s.stateRepo.GetStatus()})

	// CORREÇÃO: Throttling de eventos
	var lastUpdate time.Time

	progress := func(cur, total int64) {
		s.stateRepo.UpdateAppStatus(appID, func(st *shared.AppStatus) {
			st.DownloadedSize = cur
			st.TotalSize = total
			if total > 0 {
				st.Progress = int(float64(cur) / float64(total) * 100)
			}
		})

		// Só envia evento a cada 200ms para não engasgar a UI
		if time.Since(lastUpdate) > 1000*time.Millisecond {
			s.eventBus.Broadcast(shared.SSEEvent{Type: "status_update", Payload: s.stateRepo.GetStatus()})
			lastUpdate = time.Now()
		}
	}

	err := s.pkgManager.Install(appID, app.CurrentRelease.DownloadURL, app.CurrentRelease.Checksum, progress)

	s.stateRepo.UpdateAppStatus(appID, func(st *shared.AppStatus) {
		if err != nil {
			st.State = shared.StateError
			st.Error = err.Error()
		} else {
			st.State = shared.StateIdle
			st.Message = "Sucesso"
			st.CurrentVersion = app.CurrentRelease.Version
		}
	})
	// Envia evento final garantido
	s.eventBus.Broadcast(shared.SSEEvent{Type: "status_update", Payload: s.stateRepo.GetStatus()})
}
