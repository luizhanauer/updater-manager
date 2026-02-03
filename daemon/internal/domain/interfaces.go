package domain

import (
	"time"

	"github.com/luizhanauer/linux-updater/shared"
)

type PackageManager interface {
	GetInstalledVersion(pkgName string) (string, error)
	Install(appName, url, checksum string, progress func(int64, int64)) error
	Remove(pkgName string) error
}

type StateRepository interface {
	GetStatus() shared.Status
	UpdateAppStatus(appID string, modifier func(*shared.AppStatus))
	SetFullStatus(status shared.Status)
	SetCatalogCheckTime(t time.Time)
}

type EventBus interface {
	Broadcast(event shared.SSEEvent)
}
