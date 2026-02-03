package mem

import (
	"sync"
	"time"

	"github.com/luizhanauer/linux-updater/shared"
)

type StateRepository struct {
	status shared.Status
	mu     sync.RWMutex
}

func NewStateRepository() *StateRepository {
	return &StateRepository{
		status: shared.Status{Apps: make(map[string]shared.AppStatus)},
	}
}

func (r *StateRepository) GetStatus() shared.Status {
	r.mu.RLock()
	defer r.mu.RUnlock()
	// Deep copy para segurança
	cp := shared.Status{
		UpdatedAt:        r.status.UpdatedAt,
		CatalogCheckedAt: r.status.CatalogCheckedAt,
		Apps:             make(map[string]shared.AppStatus),
	}
	for k, v := range r.status.Apps {
		cp.Apps[k] = v
	}
	return cp
}

func (r *StateRepository) UpdateAppStatus(appID string, mod func(*shared.AppStatus)) {
	r.mu.Lock()
	defer r.mu.Unlock()
	s := r.status.Apps[appID]
	mod(&s)
	r.status.Apps[appID] = s
	r.status.UpdatedAt = time.Now()
}

func (r *StateRepository) SetFullStatus(s shared.Status) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.status = s
}

func (r *StateRepository) SetCatalogCheckTime(t time.Time) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.status.CatalogCheckedAt = t
	// Não alteramos UpdatedAt aqui para não disparar reatividade desnecessária nos apps
}
