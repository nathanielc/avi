package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/nathanielc/avi"
)

const gameIDLen = 16

type Handler struct {
	r *mux.Router
	h http.Handler

	mu sync.RWMutex

	activeGames map[string]*game
	data        *data
}

func newHandler(d *data) *Handler {
	r := mux.NewRouter()
	lh := handlers.CombinedLoggingHandler(os.Stderr, r)
	return &Handler{
		r:           r,
		h:           lh,
		data:        d,
		activeGames: make(map[string]*game),
	}
}

func (h *Handler) Open() error {
	h.r.HandleFunc("/avi/ping", h.ping).Methods("GET")
	h.r.HandleFunc("/avi/maps", h.getMaps).Methods("GET")
	h.r.HandleFunc("/avi/part_sets", h.getParts).Methods("GET")
	h.r.HandleFunc("/avi/fleets", h.getFleets).Methods("GET")
	h.r.HandleFunc("/avi/games", h.startGame).Methods("POST")
	h.r.HandleFunc("/avi/games", h.getGames).Methods("GET")
	h.r.HandleFunc("/avi/games/{id}", h.streamGame).Methods("GET")
	return nil
}

func (h *Handler) Close() {
	for _, g := range h.activeGames {
		g.Close()
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.h.ServeHTTP(w, r)
}

func (h *Handler) ping(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"pong":true}`))
}

func (h *Handler) getMaps(w http.ResponseWriter, r *http.Request) {
	h.mu.RLock()
	maps, err := h.data.Maps()
	h.mu.RUnlock()
	if err != nil {
		h.error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(maps)
}

func (h *Handler) getParts(w http.ResponseWriter, r *http.Request) {
	h.mu.RLock()
	partSets, err := h.data.PartSets()
	h.mu.RUnlock()
	if err != nil {
		h.error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(partSets)
}
func (h *Handler) getFleets(w http.ResponseWriter, r *http.Request) {
	h.mu.RLock()
	fleets, err := h.data.Fleets()
	h.mu.RUnlock()
	if err != nil {
		h.error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(fleets)
}

type Game struct {
	ID     string    `json:"id"`
	Date   time.Time `json:"date"`
	Active bool      `json:"active"`
}

type gamesResponse struct {
	Games map[string]Game `json:"games"`
}

func (h *Handler) getGames(w http.ResponseWriter, r *http.Request) {
	games := make(map[string]Game)
	n := time.Now()

	h.mu.RLock()
	for id := range h.activeGames {
		games[id] = Game{
			ID:     id,
			Date:   n,
			Active: true,
		}
	}
	replays, err := h.data.Replays()
	h.mu.RUnlock()

	if err != nil {
		h.error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for _, r := range replays {
		if _, ok := games[r.GameID]; !ok {
			games[r.GameID] = Game{
				ID:     r.GameID,
				Date:   r.Date,
				Active: false,
			}
		}
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(gamesResponse{Games: games})
}

type jsonError struct {
	Error string `json:"error"`
}

func (h *Handler) error(w http.ResponseWriter, err string, code int) {
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(jsonError{Error: err})
}

type startGameRequest struct {
	Map     string   `json:"map"`
	PartSet string   `json:"part_set"`
	Fleets  []string `json:"fleets"`
	FPS     int      `json:"fps"`
	MaxTime int64    `json:"max_time"`
}
type startGameResponse struct {
	ID string `json:"id"`
}

func defaultStartGameRequest() startGameRequest {
	return startGameRequest{
		FPS:     60,
		MaxTime: int64(10 * time.Minute),
	}
}

func (h *Handler) startGame(w http.ResponseWriter, r *http.Request) {
	sgr := defaultStartGameRequest()
	if err := json.NewDecoder(r.Body).Decode(&sgr); err != nil {
		h.error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	m, err := h.data.Map(sgr.Map)
	if err != nil {
		h.error(w, fmt.Sprintf("unknown map %q: %v", sgr.Map, err), http.StatusNotFound)
		return
	}

	ps, err := h.data.PartSet(sgr.PartSet)
	if err != nil {
		h.error(w, fmt.Sprintf("unknown part set %q: %v", sgr.PartSet, err), http.StatusNotFound)
		return
	}

	fleets := make([]avi.FleetConf, len(sgr.Fleets))
	for i, f := range sgr.Fleets {
		fleet, err := h.data.Fleet(f)
		if err != nil {
			h.error(w, fmt.Sprintf("unknown fleet %q: %v", f, err), http.StatusNotFound)
			return
		}
		fleets[i] = fleet
	}

	id := randString(gameIDLen)
	replay := h.data.NewReplay(id)
	finished := make(chan struct{})
	g := newGame(id, replay, finished)
	h.activeGames[id] = g

	sim, err := avi.NewSimulation(
		m,
		ps,
		fleets,
		g,
		time.Duration(sgr.MaxTime),
		int64(sgr.FPS),
	)
	if err != nil {
		h.error(w, fmt.Sprintf("failed to create simulation: %v", err), http.StatusNotFound)
		return
	}
	g.sim = sim

	if err := g.Open(); err != nil {
		h.error(w, fmt.Sprintf("failed to start game: %v", err), http.StatusNotFound)
		return
	}

	go func() {
		// Wait till game finishes
		<-finished
		g.Close()
		h.mu.Lock()
		delete(h.activeGames, id)
		h.mu.Unlock()
	}()

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(startGameResponse{ID: id})
}

func (h *Handler) streamGame(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	h.mu.RLock()
	g, ok := h.activeGames[id]
	h.mu.RUnlock()

	if !ok {
		// Check for replay
		h.mu.RLock()
		replay, err := h.data.Replay(id)
		h.mu.RUnlock()
		if err != nil {
			h.error(w, fmt.Sprintf("unknown game %q", id), http.StatusNotFound)
			return
		}
		rc, err := replay.ReadCloser()
		if err != nil {
			h.error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rc.Close()
		w.WriteHeader(http.StatusOK)
		io.Copy(w, rc)
		return
	}

	w.WriteHeader(http.StatusOK)
	g.Stream(w)
}
