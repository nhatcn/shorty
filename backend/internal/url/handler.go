package url

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

type createURLRequest struct {
	OriginalURL string `json:"original_url"`
	UserID      int64  `json:"user_id"` // For demo, in real use from JWT
}

type createURLResponse struct {
	ShortURL string `json:"short_url"`
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.CreateShortURL(w, r)
	case http.MethodGet:
		h.ListURLs(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) CreateShortURL(w http.ResponseWriter, r *http.Request) {
	var req createURLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	shortCode, err := h.service.CreateShortURL(req.UserID, req.OriginalURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	resp := createURLResponse{
		ShortURL: "http://localhost:8080/" + shortCode,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) Redirect(w http.ResponseWriter, r *http.Request) {
	shortCode := r.URL.Path[1:]
	if shortCode == "" {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	originalURL, err := h.service.GetOriginalURL(shortCode)
	if err != nil {
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}
	http.Redirect(w, r, originalURL, http.StatusFound)
}

func (h *Handler) ListURLs(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("user_id")
	userID, _ := strconv.ParseInt(query, 10, 64)
	urls, err := h.service.ListURLs(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(urls)
}
func (h *Handler) UserStats(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("user_id")
	userID, _ := strconv.ParseInt(query, 10, 64)
	stats, err := h.service.GetUserStats(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
