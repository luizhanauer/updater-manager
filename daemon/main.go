package main

import (
	"log"
	"runtime/debug"
	"time"

	"github.com/luizhanauer/linux-updater/daemon/internal/adapter/deb"
	"github.com/luizhanauer/linux-updater/daemon/internal/adapter/mem"
	"github.com/luizhanauer/linux-updater/daemon/internal/adapter/socket"
	"github.com/luizhanauer/linux-updater/daemon/internal/usecase"
	"github.com/luizhanauer/linux-updater/shared"
)

func main() {
	log.Println(">>> Updater Daemon v3.2 (Cache + Refresh) Iniciado")

	if err := shared.SaveGoalsDisk(shared.LoadGoalsDisk()); err == nil {
	}

	trigger := make(chan struct{}, 1)

	// Adapters
	aptManager := deb.NewAptManager()
	stateRepo := mem.NewStateRepository()

	// Declaração prévia para usar na closure
	var updater *usecase.UpdaterService

	// Callback que o Server chamará quando receber POST /catalog/refresh
	onRefresh := func() {
		if updater != nil {
			log.Println("⚡ Atualização manual solicitada!")
			updater.ForceCatalogRefresh()
		}
	}

	server := socket.NewServer(stateRepo, trigger, onRefresh)
	updater = usecase.NewUpdaterService(aptManager, stateRepo, server)

	go func() {
		if err := server.Start(); err != nil {
			log.Fatalf("Socket falhou: %v", err)
		}
	}()

	ticker := time.NewTicker(30 * time.Second)

	// Boot: Força atualização do catálogo (true)
	updater.Reconcile(true)

	for {
		select {
		case <-ticker.C:
			// Loop: Usa cache se disponível (false)
			updater.Reconcile(false)
			debug.FreeOSMemory()
		case <-trigger:
			// Trigger de Ação: Usa cache se disponível (false)
			log.Println("⚡ Trigger de ação recebido")
			updater.Reconcile(false)
			debug.FreeOSMemory()
		}
	}
}
