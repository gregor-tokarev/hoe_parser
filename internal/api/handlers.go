package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gregor-tokarev/hoe_parser/internal/config"
	"github.com/gregor-tokarev/hoe_parser/internal/scraper"
	"google.golang.org/protobuf/encoding/protojson"
)

// Handlers struct holds the API handlers
type Handlers struct {
	config  *config.Config
	scraper *scraper.IntimcityScraper
}

// NewHandlers creates a new handlers instance
func NewHandlers(cfg *config.Config) *Handlers {
	return &Handlers{
		config:  cfg,
		scraper: scraper.NewIntimcityScraper(),
	}
}

// ScrapeRequest represents the request payload for scraping
type ScrapeRequest struct {
	URL string `json:"url"`
}

// ScrapeResponse represents the response for scraping
type ScrapeResponse struct {
	Success bool   `json:"success"`
	Data    string `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

// HealthHandler handles health check requests
func (h *Handlers) HealthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := map[string]interface{}{
		"status":    "ok",
		"service":   "hoe_parser",
		"version":   "1.0.0",
		"timestamp": "2025-06-26T07:00:00Z",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ScrapeHandler handles scraping requests
func (h *Handlers) ScrapeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ScrapeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response := ScrapeResponse{
			Success: false,
			Error:   "Invalid request body",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	if req.URL == "" {
		response := ScrapeResponse{
			Success: false,
			Error:   "URL is required",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Scrape the listing
	listing, err := h.scraper.ScrapeListing(req.URL)
	if err != nil {
		response := ScrapeResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to scrape listing: %v", err),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Convert to JSON
	jsonData, err := protojson.MarshalOptions{
		Multiline: false,
		Indent:    "",
	}.Marshal(listing)
	if err != nil {
		response := ScrapeResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to marshal data: %v", err),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := ScrapeResponse{
		Success: true,
		Data:    string(jsonData),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// SetupRoutes sets up HTTP routes
func (h *Handlers) SetupRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", h.HealthHandler)
	mux.HandleFunc("/api/v1/health", h.HealthHandler)
	mux.HandleFunc("/api/v1/scrape", h.ScrapeHandler)

	return mux
}
