package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/AntonMilienkov/REST-service/internal/model"
)

// ErrNotFound возвращается, когда запись с таким id не найдена.
var ErrNotFound = errors.New("subscription not found")

// SubscriptionRepository описывает доступ к хранилищу записей о подписках.
type SubscriptionRepository interface {
	Create(ctx context.Context, sub *model.Subscription) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Subscription, error)
	List(ctx context.Context) ([]model.Subscription, error)
	Update(ctx context.Context, sub *model.Subscription) error
	Delete(ctx context.Context, id uuid.UUID) error
}
