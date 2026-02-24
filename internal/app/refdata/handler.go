package refdata

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/darahayes/go-boom"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	log     *slog.Logger
	service Service
}

func NewHandler(log *slog.Logger, service Service) *Handler {
	return &Handler{log: log, service: service}
}

func (h *Handler) GetCityByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	idInt, err := strconv.Atoi(id)
	if err != nil {
		boom.BadRequest(w, err)
		return
	}

	result, err := h.service.GetCityByID(r.Context(), idInt)
	if err != nil {
		boom.Internal(w, err)
		return
	}

	h.sendJSON(w, result, http.StatusOK)
}

func (h *Handler) GetCitiesByRegionID(w http.ResponseWriter, r *http.Request) {
	regionID := chi.URLParam(r, "region_id")

	regionIDInt, err := strconv.Atoi(regionID)
	if err != nil {
		boom.BadRequest(w, err)
		return
	}

	result, err := h.service.GetCitiesByRegionID(r.Context(), regionIDInt)
	if err != nil {
		boom.Internal(w, err)
		return
	}

	h.sendJSON(w, result, http.StatusOK)
}

func (h *Handler) GetRegionByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	idInt, err := strconv.Atoi(id)
	if err != nil {
		boom.BadRequest(w, err)
		return
	}

	result, err := h.service.GetRegionByID(r.Context(), idInt)
	if err != nil {
		boom.Internal(w, err)
		return
	}

	h.sendJSON(w, result, http.StatusOK)
}

func (h *Handler) GetAllRegions(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.GetAllRegions(r.Context())
	if err != nil {
		boom.Internal(w, err)
		return
	}

	h.sendJSON(w, result, http.StatusOK)
}

func (h *Handler) GetAllSports(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.GetAllSports(r.Context())
	if err != nil {
		boom.Internal(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

func (h *Handler) GetSportByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		boom.BadRequest(w, "ID is required")
		return
	}

	sport, err := h.service.GetSportByID(r.Context(), id)
	if err != nil {
		boom.Internal(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(sport)
}

func (h *Handler) sendJSON(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}
