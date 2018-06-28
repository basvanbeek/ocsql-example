package server

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// Service implementation
type Service struct {
	repository Repository
}

// New returns an initialized Service instance.
func New(repository Repository) *Service {
	return &Service{
		repository: repository,
	}
}

// ListItems is a HTTP Handler
func (s *Service) ListItems(w http.ResponseWriter, r *http.Request) {
	var (
		ctx    = r.Context()
		vars   = mux.Vars(r)
		userID = vars["user_id"]
	)

	ownerID, err := strconv.ParseInt(userID, 10, 32)
	if err != nil {
		log.Printf("invalid user_id: %v\n", err)
		http.Error(w, "invalid user id provided", http.StatusBadRequest)
		return
	}

	items, err := s.repository.ListItems(ctx, ownerID)
	if err != nil {
		log.Printf("repository error: %v\n", err)
		http.Error(w, "unable to retrieve items", http.StatusInternalServerError)
		return
	}
	if len(items) == 0 {
		http.Error(w, "items not found", http.StatusNotFound)
		return
	}

	w.Header().Set("content", "application/json")
	enc := json.NewEncoder(w)
	if err = enc.Encode(items); err != nil {
		log.Printf("json encode error: %v\n", err)
		http.Error(w, "unable to output items", http.StatusInternalServerError)
	}
}

// Item holds some data
type Item struct {
	ID     int64  `json:"id"`
	UserID int64  `json:"user_id,omitempty"`
	Name   string `json:"name"`
}
