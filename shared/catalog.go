package shared

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

const CatalogURL = "https://raw.githubusercontent.com/luizhanauer/updater-registry/refs/heads/main/api/catalog.json"

var (
	cachedCatalog map[string]App
	lastFetch     time.Time
	CacheDuration = 24 * time.Hour
)

func GetLastFetch() time.Time {
	return lastFetch
}

func GetCatalog(forceRefresh bool) map[string]App {
	// Se tem cache, não forçou, e está dentro das 24h: usa cache
	if !forceRefresh && cachedCatalog != nil && time.Since(lastFetch) < CacheDuration {
		return cachedCatalog
	}

	log.Println("🔄 Baixando catálogo remoto...")
	client := http.Client{Timeout: 10 * time.Second}
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
