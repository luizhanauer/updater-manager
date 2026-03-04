package shared

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"
)

const CatalogURL = "https://raw.githubusercontent.com/luizhanauer/updater-registry/refs/heads/main/api/catalog.json"

var (
	cachedCatalog map[string]App
	lastFetch     time.Time
	CacheDuration = 24 * time.Hour
	catalogLock   sync.RWMutex
)

func GetLastFetch() time.Time {
	catalogLock.RLock()
	defer catalogLock.RUnlock()
	return lastFetch
}

func GetCatalog(forceRefresh bool) map[string]App {
	catalogLock.RLock()
	// Se tem cache, não forçou, e está dentro das 24h: usa cache
	if !forceRefresh && cachedCatalog != nil && time.Since(lastFetch) < CacheDuration {
		defer catalogLock.RUnlock()
		return cachedCatalog
	}
	catalogLock.RUnlock()

	// A partir daqui, precisamos de um lock de escrita
	catalogLock.Lock()
	defer catalogLock.Unlock()

	log.Println("🔄 Baixando catálogo remoto...")
	client := http.Client{Timeout: 15 * time.Second} // Aumentado um pouco o timeout
	resp, err := client.Get(CatalogURL)
	if err != nil {
		log.Printf("Erro infra: %v", err)
		if cachedCatalog != nil {
			return cachedCatalog
		}
		return map[string]App{}
	}
	defer resp.Body.Close()

	var wrapper RegistryCatalog
	if err := json.NewDecoder(resp.Body).Decode(&wrapper); err != nil {
		log.Printf("Erro parse: %v", err)
		if cachedCatalog != nil {
			return cachedCatalog
		}
		return map[string]App{}
	}

	cachedCatalog = wrapper.Apps
	lastFetch = time.Now()
	return cachedCatalog
}
