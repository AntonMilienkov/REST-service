package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/AntonMilienkov/REST-service/internal/model"
	"github.com/AntonMilienkov/REST-service/internal/repository"
)

// ErrValidation оборачивает ошибки некорректных входных данных.
var ErrValidation = errors.New("validation error")

// SubscriptionService содержит бизнес-логику и валидацию поверх репозитория подписок.
type SubscriptionService struct {
	repo   repository.SubscriptionRepository
	logger *slog.Logger
}

func NewSubscriptionService(repo repository.SubscriptionRepository, logger *slog.Logger) *SubscriptionService {
	return &SubscriptionService{repo: repo, logger: logger}
}

func (s *SubscriptionService) Create(ctx context.Context, sub *model.Subscription) error {
	if err := validate(sub); err != nil {
		s.logger.Warn("create validation failed", "error", err)
		return err
	}

	sub.ID = uuid.New()
	if err := s.repo.Create(ctx, sub); err != nil {
		return err
	}

	s.logger.Info("subscription created", "id", sub.ID, "user_id", sub.UserID, "service_name", sub.ServiceName)
	return nil
}

func (s *SubscriptionService) GetByID(ctx context.Context, id uuid.UUID) (*model.Subscription, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *SubscriptionService) List(ctx context.Context) ([]model.Subscription, error) {
	return s.repo.List(ctx)
}

func (s *SubscriptionService) Update(ctx context.Context, sub *model.Subscription) error {
	if err := validate(sub); err != nil {
		s.logger.Warn("update validation failed", "error", err, "id", sub.ID)
		return err
	}

	if err := s.repo.Update(ctx, sub); err != nil {
		return err
	}

	s.logger.Info("subscription updated", "id", sub.ID)
	return nil
}

func (s *SubscriptionService) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	s.logger.Info("subscription deleted", "id", id)
	return nil
}

// TotalCostFilter задаёт фильтры для подсчёта суммарной стоимости подписок за период.
type TotalCostFilter struct {
	UserID      *uuid.UUID
	ServiceName *string
	PeriodFrom  time.Time
	PeriodTo    time.Time
}

func (s *SubscriptionService) TotalCost(ctx context.Context, filter TotalCostFilter) (int, error) {
	if filter.PeriodFrom.IsZero() || filter.PeriodTo.IsZero() {
		err := fmt.Errorf("%w: period_from and period_to are required", ErrValidation)
		s.logger.Warn("total cost validation failed", "error", err)
		return 0, err
	}

	if filter.PeriodTo.Before(filter.PeriodFrom) {
		err := fmt.Errorf("%w: period_to must not be before period_from", ErrValidation)
		s.logger.Warn("total cost validation failed", "error", err)
		return 0, err
	}

	return s.repo.SumByPeriod(ctx, repository.SumFilter{
		UserID:      filter.UserID,
		ServiceName: filter.ServiceName,
		PeriodFrom:  filter.PeriodFrom,
		PeriodTo:    filter.PeriodTo,
	})
}

func validate(sub *model.Subscription) error {
	if strings.TrimSpace(sub.ServiceName) == "" {
		return fmt.Errorf("%w: service_name is required", ErrValidation)
	}

	if sub.Price < 0 {
		return fmt.Errorf("%w: price must be >= 0", ErrValidation)
	}

	if sub.UserID == uuid.Nil {
		return fmt.Errorf("%w: user_id is required", ErrValidation)
	}

	if sub.StartDate.IsZero() {
		return fmt.Errorf("%w: start_date is required", ErrValidation)
	}

	if sub.EndDate != nil && sub.EndDate.Before(sub.StartDate.Time) {
		return fmt.Errorf("%w: end_date must not be before start_date", ErrValidation)
	}

	return nil
}
