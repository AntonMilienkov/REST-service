package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/AntonMilienkov/REST-service/internal/model"
	"github.com/AntonMilienkov/REST-service/internal/repository"
	"github.com/AntonMilienkov/REST-service/internal/service"
)

// SubscriptionHandler отдаёт HTTP CRUD поверх SubscriptionService.
type SubscriptionHandler struct {
	svc *service.SubscriptionService
}

func NewSubscriptionHandler(svc *service.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{svc: svc}
}

// errorResponse — тело ответа об ошибке.
type errorResponse struct {
	Error string `json:"error"`
}

// totalCostResponse — тело ответа ручки подсчёта суммарной стоимости.
type totalCostResponse struct {
	Total int `json:"total"`
}

// Create godoc
// @Summary Создать подписку
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param input body model.Subscription true "Данные подписки"
// @Success 201 {object} model.Subscription
// @Failure 400 {object} errorResponse
// @Router /subscriptions [post]
func (h *SubscriptionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var sub model.Subscription
	if err := json.NewDecoder(r.Body).Decode(&sub); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.svc.Create(r.Context(), &sub); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, sub)
}

// Get godoc
// @Summary Получить подписку по id
// @Tags subscriptions
// @Produce json
// @Param id path string true "ID подписки"
// @Success 200 {object} model.Subscription
// @Failure 400 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Router /subscriptions/{id} [get]
func (h *SubscriptionHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	sub, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, sub)
}

// List godoc
// @Summary Список всех подписок
// @Tags subscriptions
// @Produce json
// @Success 200 {array} model.Subscription
// @Router /subscriptions [get]
func (h *SubscriptionHandler) List(w http.ResponseWriter, r *http.Request) {
	subs, err := h.svc.List(r.Context())
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, subs)
}

// Update godoc
// @Summary Обновить подписку
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path string true "ID подписки"
// @Param input body model.Subscription true "Данные подписки"
// @Success 200 {object} model.Subscription
// @Failure 400 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Router /subscriptions/{id} [put]
func (h *SubscriptionHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var sub model.Subscription
	if err := json.NewDecoder(r.Body).Decode(&sub); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	sub.ID = id

	if err := h.svc.Update(r.Context(), &sub); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, sub)
}

// TotalCost godoc
// @Summary Суммарная стоимость подписок за период
// @Tags subscriptions
// @Produce json
// @Param period_from query string true "Начало периода, формат MM-YYYY"
// @Param period_to query string true "Конец периода, формат MM-YYYY"
// @Param user_id query string false "Фильтр по ID пользователя"
// @Param service_name query string false "Фильтр по названию сервиса"
// @Success 200 {object} totalCostResponse
// @Failure 400 {object} errorResponse
// @Router /subscriptions/total-cost [get]
func (h *SubscriptionHandler) TotalCost(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	periodFromStr := q.Get("period_from")
	periodToStr := q.Get("period_to")
	if periodFromStr == "" || periodToStr == "" {
		writeError(w, http.StatusBadRequest, "period_from and period_to are required")
		return
	}

	periodFrom, err := model.ParseMonthYear(periodFromStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	periodTo, err := model.ParseMonthYear(periodToStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	filter := service.TotalCostFilter{
		PeriodFrom: periodFrom,
		PeriodTo:   periodTo,
	}

	if userIDStr := q.Get("user_id"); userIDStr != "" {
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid user_id")
			return
		}
		filter.UserID = &userID
	}

	if serviceName := q.Get("service_name"); serviceName != "" {
		filter.ServiceName = &serviceName
	}

	total, err := h.svc.TotalCost(r.Context(), filter)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, totalCostResponse{Total: total})
}

// Delete godoc
// @Summary Удалить подписку
// @Tags subscriptions
// @Param id path string true "ID подписки"
// @Success 204
// @Failure 400 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Router /subscriptions/{id} [delete]
func (h *SubscriptionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	if err := h.svc.Delete(r.Context(), id); err != nil {
		handleServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func handleServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrValidation):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, repository.ErrNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "internal error")
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorResponse{Error: message})
}
