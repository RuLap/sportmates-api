package profile

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/RuLap/sportmates-api/internal/pkg/errors"
	validation "github.com/RuLap/sportmates-api/internal/pkg/validator"
	"github.com/darahayes/go-boom"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handler struct {
	log     *slog.Logger
	service Service
}

func NewHandler(log *slog.Logger, service Service) *Handler {
	return &Handler{log: log, service: service}
}

func (h *Handler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	id, err := h.getUrlParamUuid(r, "id")
	if err != nil {
		boom.BadRequest(w, err)
		return
	}

	response, err := h.service.GetUserByID(r.Context(), *id)
	if err != nil {
		boom.Internal(w, err)
		return
	}

	h.sendJSON(w, response, http.StatusOK)
}

func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	var req SaveProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		boom.BadRequest(w, "неверный формат JSON")
		return
	}

	if errors := validation.ValidateStruct(req); errors != nil {
		boom.BadRequest(w, "ошибки валидации", errors)
		return
	}

	userID, err := h.getUserIDFromContext(r.Context())
	if err != nil {
		boom.BadRequest(w, err)
	}

	response, err := h.service.SaveProfile(r.Context(), &req, userID)
	if err != nil {
		boom.Internal(w, err)
		return
	}

	h.sendJSON(w, response, http.StatusOK)
}

func (h *Handler) GetAvatarUploadURL(w http.ResponseWriter, r *http.Request) {
	id, err := h.getUserIDFromContext(r.Context())
	if err != nil {
		boom.BadRequest(w, err)
	}

	response, err := h.service.GetAvatarUploadURL(r.Context(), *id)
	if err != nil {
		boom.Internal(w, err)
		return
	}

	h.sendJSON(w, response, http.StatusOK)
}

func (h *Handler) ConfirmAvatarUpload(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromContext(r.Context())
	if err != nil {
		boom.BadRequest(w, err)
	}

	response, err := h.service.ConfirmAvatarUpload(r.Context(), *userID)
	if err != nil {
		boom.Internal(w, err)
		return
	}

	h.sendJSON(w, response, http.StatusOK)
}

func (h *Handler) getUrlParamUuid(r *http.Request, param string) (*uuid.UUID, error) {
	str := chi.URLParam(r, param)
	if str == "" {
		err := fmt.Errorf("параметр %s необходим", param)
		h.log.Error("Incorrect ID in URL", param, str, "error", err.Error())
		return nil, err
	}

	uid, err := uuid.Parse(str)
	if err != nil {
		err := fmt.Errorf("неверный формат параметра %s", param)
		h.log.Error("Incorrect ID in URL", param, str, "error", err.Error())
		return nil, err
	}

	return &uid, nil
}

func (h *Handler) sendJSON(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) getUserIDFromContext(ctx context.Context) (*uuid.UUID, error) {
	userIDStr, ok := ctx.Value("user_id").(string)
	if !ok {
		h.log.Error("Incorrect ID in context", "userID", userIDStr)
		return nil, fmt.Errorf(errors.ErrCommon)
	}

	id, err := uuid.Parse(userIDStr)
	if err != nil {
		h.log.Error("failed to parse userID from context", "userID", userIDStr, "error", err)
		return nil, fmt.Errorf(errors.ErrCommon)
	}

	return &id, nil
}
