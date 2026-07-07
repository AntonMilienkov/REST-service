package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/AntonMilienkov/REST-service/internal/model"
)

// PostgresSubscriptionRepository — реализация SubscriptionRepository поверх pgx.
type PostgresSubscriptionRepository struct {
	pool *pgxpool.Pool
}

var _ SubscriptionRepository = (*PostgresSubscriptionRepository)(nil)

func NewSubscriptionRepository(pool *pgxpool.Pool) *PostgresSubscriptionRepository {
	return &PostgresSubscriptionRepository{pool: pool}
}

func (r *PostgresSubscriptionRepository) Create(ctx context.Context, sub *model.Subscription) error {
	const query = `
		INSERT INTO subscriptions (id, service_name, price, user_id, start_date, end_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, now(), now())
		RETURNING created_at, updated_at`

	err := r.pool.QueryRow(ctx, query,
		sub.ID, sub.ServiceName, sub.Price, sub.UserID, sub.StartDate.Time, monthYearToTime(sub.EndDate),
	).Scan(&sub.CreatedAt, &sub.UpdatedAt)
	if err != nil {
		return fmt.Errorf("insert subscription: %w", err)
	}

	return nil
}

func (r *PostgresSubscriptionRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Subscription, error) {
	const query = `
		SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
		FROM subscriptions
		WHERE id = $1`

	var sub model.Subscription
	var endDate *time.Time

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&sub.ID, &sub.ServiceName, &sub.Price, &sub.UserID, &sub.StartDate.Time, &endDate, &sub.CreatedAt, &sub.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get subscription: %w", err)
	}

	sub.EndDate = timeToMonthYear(endDate)
	return &sub, nil
}

func (r *PostgresSubscriptionRepository) List(ctx context.Context) ([]model.Subscription, error) {
	const query = `
		SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
		FROM subscriptions
		ORDER BY created_at`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list subscriptions: %w", err)
	}
	defer rows.Close()

	var subs []model.Subscription
	for rows.Next() {
		var sub model.Subscription
		var endDate *time.Time

		if err := rows.Scan(
			&sub.ID, &sub.ServiceName, &sub.Price, &sub.UserID, &sub.StartDate.Time, &endDate, &sub.CreatedAt, &sub.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan subscription: %w", err)
		}

		sub.EndDate = timeToMonthYear(endDate)
		subs = append(subs, sub)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list subscriptions: %w", err)
	}

	return subs, nil
}

func (r *PostgresSubscriptionRepository) Update(ctx context.Context, sub *model.Subscription) error {
	const query = `
		UPDATE subscriptions
		SET service_name = $1, price = $2, user_id = $3, start_date = $4, end_date = $5, updated_at = now()
		WHERE id = $6
		RETURNING updated_at`

	err := r.pool.QueryRow(ctx, query,
		sub.ServiceName, sub.Price, sub.UserID, sub.StartDate.Time, monthYearToTime(sub.EndDate), sub.ID,
	).Scan(&sub.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("update subscription: %w", err)
	}

	return nil
}

func (r *PostgresSubscriptionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM subscriptions WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete subscription: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func monthYearToTime(m *model.MonthYear) *time.Time {
	if m == nil {
		return nil
	}
	t := m.Time
	return &t
}

func timeToMonthYear(t *time.Time) *model.MonthYear {
	if t == nil {
		return nil
	}
	return &model.MonthYear{Time: *t}
}
