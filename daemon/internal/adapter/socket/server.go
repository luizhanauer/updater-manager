package socket

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"

	"github.com/luizhanauer/linux-updater/daemon/internal/domain"
	"github.com/luizhanauer/linux-updater/shared"
)

type Server struct {
	stateRepo       domain.StateRepository
	clients         map[chan shared.SSEEvent]bool
	lock            sync.RWMutex
	triggerChan     chan struct{}
	refreshCallback func()
}

func NewServer(state domain.StateRepository, trigger chan struct{}, onRefresh func()) *Server {
	return &Server{
		stateRepo:       state,
		clients:         make(map[chan shared.SSEEvent]bool),
		triggerChan:     trigger,
		refreshCallback: onRefresh,
	}
}

func (s *Server) Start() error {
	os.Remove(shared.SocketPath)
	listener, err := net.Listen("unix", shared.SocketPath)
	if err != nil {
		return err
	}
	if err := os.Chmod(shared.SocketPath, 0660); err != nil {
		return err
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/status", s.handleGetStatus)
	mux.HandleFunc("/goals", s.handlePostGoals)
	mux.HandleFunc("/events", s.handleSSE)
	mux.HandleFunc("/catalog/refresh", s.handleCatalogRefresh)

	return http.Serve(listener, mux)
}

func (s *Server) Broadcast(event shared.SSEEvent) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	for ch := range s.clients {
		select {
		case ch <- event:
		default:
		}
	}
}

func (s *Server) handleSSE(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	events := make(chan shared.SSEEvent, 10)
	s.lock.Lock()
	s.clients[events] = true
	s.lock.Unlock()

	defer func() {
		s.lock.Lock()
		delete(s.clients, events)
		s.lock.Unlock()
		close(events)
	}()

	json.NewEncoder(w).Encode(shared.SSEEvent{Type: "init", Payload: s.stateRepo.GetStatus()})
	w.(http.Flusher).Flush()

	for {
		select {
		case event := <-events:
			data, _ := json.Marshal(event)
			fmt.Fprintf(w, "data: %s\n\n", data)
			w.(http.Flusher).Flush()
		case <-r.Context().Done():
			return
		}
	}
}

func (s *Server) handleGetStatus(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(s.stateRepo.GetStatus())
}

func (s *Server) handlePostGoals(w http.ResponseWriter, r *http.Request) {
	var newGoals shared.Goals
	if err := json.NewDecoder(r.Body).Decode(&newGoals); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	currentGoals := shared.LoadGoalsDisk()
	for k, v := range newGoals.Apps {
		currentGoals.Apps[k] = v
	}
	shared.SaveGoalsDisk(currentGoals)

	// Acorda o Daemon
	select {
	case s.triggerChan <- struct{}{}:
	default:
	}

	w.WriteHeader(200)
}

func (s *Server) handleCatalogRefresh(w http.ResponseWriter, r *http.Request) {
	if s.refreshCallback != nil {
		go s.refreshCallback()
	}
	w.WriteHeader(200)
}
