package service

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/AntonMilienkov/REST-service/internal/model"
	"github.com/AntonMilienkov/REST-service/internal/repository"
)

// fakeRepository — ручной фейк repository.SubscriptionRepository для тестов сервисного слоя.
type fakeRepository struct {
	createFunc      func(ctx context.Context, sub *model.Subscription) error
	getByIDFunc     func(ctx context.Context, id uuid.UUID) (*model.Subscription, error)
	listFunc        func(ctx context.Context) ([]model.Subscription, error)
	updateFunc      func(ctx context.Context, sub *model.Subscription) error
	deleteFunc      func(ctx context.Context, id uuid.UUID) error
	sumByPeriodFunc func(ctx context.Context, filter repository.SumFilter) (int, error)
}

var _ repository.SubscriptionRepository = (*fakeRepository)(nil)

func (f *fakeRepository) Create(ctx context.Context, sub *model.Subscription) error {
	return f.createFunc(ctx, sub)
}

func (f *fakeRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Subscription, error) {
	return f.getByIDFunc(ctx, id)
}

func (f *fakeRepository) List(ctx context.Context) ([]model.Subscription, error) {
	return f.listFunc(ctx)
}

func (f *fakeRepository) Update(ctx context.Context, sub *model.Subscription) error {
	return f.updateFunc(ctx, sub)
}

func (f *fakeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return f.deleteFunc(ctx, id)
}

func (f *fakeRepository) SumByPeriod(ctx context.Context, filter repository.SumFilter) (int, error) {
	return f.sumByPeriodFunc(ctx, filter)
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func validSubscription() *model.Subscription {
	return &model.Subscription{
		ServiceName: "Yandex Plus",
		Price:       400,
		UserID:      uuid.New(),
		StartDate:   model.MonthYear{Time: time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC)},
	}
}

func TestSubscriptionService_Create_Validation(t *testing.T) {
	tests := []struct {
		name string
		sub  func() *model.Subscription
	}{
		{
			name: "empty service_name",
			sub: func() *model.Subscription {
				sub := validSubscription()
				sub.ServiceName = "   "
				return sub
			},
		},
		{
			name: "negative price",
			sub: func() *model.Subscription {
				sub := validSubscription()
				sub.Price = -1
				return sub
			},
		},
		{
			name: "nil user_id",
			sub: func() *model.Subscription {
				sub := validSubscription()
				sub.UserID = uuid.Nil
				return sub
			},
		},
		{
			name: "zero start_date",
			sub: func() *model.Subscription {
				sub := validSubscription()
				sub.StartDate = model.MonthYear{}
				return sub
			},
		},
		{
			name: "end_date before start_date",
			sub: func() *model.Subscription {
				sub := validSubscription()
				end := model.MonthYear{Time: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)}
				sub.EndDate = &end
				return sub
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeRepository{
				createFunc: func(ctx context.Context, sub *model.Subscription) error {
					t.Fatal("repo.Create must not be called for invalid input")
					return nil
				},
			}
			svc := NewSubscriptionService(repo, discardLogger())

			err := svc.Create(context.Background(), tt.sub())
			if !errors.Is(err, ErrValidation) {
				t.Fatalf("expected ErrValidation, got %v", err)
			}
		})
	}
}

func TestSubscriptionService_Create_Success(t *testing.T) {
	var created *model.Subscription
	repo := &fakeRepository{
		createFunc: func(ctx context.Context, sub *model.Subscription) error {
			created = sub
			return nil
		},
	}
	svc := NewSubscriptionService(repo, discardLogger())

	sub := validSubscription()
	if err := svc.Create(context.Background(), sub); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if sub.ID == uuid.Nil {
		t.Fatal("expected service to generate an ID")
	}
	if created != sub {
		t.Fatal("expected repo.Create to receive the same subscription")
	}
}

func TestSubscriptionService_Create_RepoError(t *testing.T) {
	repoErr := errors.New("db is down")
	repo := &fakeRepository{
		createFunc: func(ctx context.Context, sub *model.Subscription) error {
			return repoErr
		},
	}
	svc := NewSubscriptionService(repo, discardLogger())

	err := svc.Create(context.Background(), validSubscription())
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo error to propagate, got %v", err)
	}
}

func TestSubscriptionService_Update_Validation(t *testing.T) {
	repo := &fakeRepository{
		updateFunc: func(ctx context.Context, sub *model.Subscription) error {
			t.Fatal("repo.Update must not be called for invalid input")
			return nil
		},
	}
	svc := NewSubscriptionService(repo, discardLogger())

	sub := validSubscription()
	sub.Price = -100

	err := svc.Update(context.Background(), sub)
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected ErrValidation, got %v", err)
	}
}

func TestSubscriptionService_Update_Success(t *testing.T) {
	repo := &fakeRepository{
		updateFunc: func(ctx context.Context, sub *model.Subscription) error {
			return nil
		},
	}
	svc := NewSubscriptionService(repo, discardLogger())

	if err := svc.Update(context.Background(), validSubscription()); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestSubscriptionService_Delete_Passthrough(t *testing.T) {
	id := uuid.New()
	repo := &fakeRepository{
		deleteFunc: func(ctx context.Context, gotID uuid.UUID) error {
			if gotID != id {
				t.Fatalf("expected id %v, got %v", id, gotID)
			}
			return repository.ErrNotFound
		},
	}
	svc := NewSubscriptionService(repo, discardLogger())

	if err := svc.Delete(context.Background(), id); !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("expected ErrNotFound to propagate, got %v", err)
	}
}

func TestSubscriptionService_GetByID_Passthrough(t *testing.T) {
	want := validSubscription()
	repo := &fakeRepository{
		getByIDFunc: func(ctx context.Context, id uuid.UUID) (*model.Subscription, error) {
			return want, nil
		},
	}
	svc := NewSubscriptionService(repo, discardLogger())

	got, err := svc.GetByID(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got != want {
		t.Fatal("expected the same subscription returned by repo")
	}
}

func TestSubscriptionService_List_Passthrough(t *testing.T) {
	want := []model.Subscription{*validSubscription(), *validSubscription()}
	repo := &fakeRepository{
		listFunc: func(ctx context.Context) ([]model.Subscription, error) {
			return want, nil
		},
	}
	svc := NewSubscriptionService(repo, discardLogger())

	got, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("expected %d subscriptions, got %d", len(want), len(got))
	}
}

func TestSubscriptionService_TotalCost_Validation(t *testing.T) {
	from := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name   string
		filter TotalCostFilter
	}{
		{name: "missing period_from", filter: TotalCostFilter{PeriodTo: to}},
		{name: "missing period_to", filter: TotalCostFilter{PeriodFrom: from}},
		{name: "period_to before period_from", filter: TotalCostFilter{PeriodFrom: to, PeriodTo: from}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeRepository{
				sumByPeriodFunc: func(ctx context.Context, filter repository.SumFilter) (int, error) {
					t.Fatal("repo.SumByPeriod must not be called for invalid filter")
					return 0, nil
				},
			}
			svc := NewSubscriptionService(repo, discardLogger())

			_, err := svc.TotalCost(context.Background(), tt.filter)
			if !errors.Is(err, ErrValidation) {
				t.Fatalf("expected ErrValidation, got %v", err)
			}
		})
	}
}

func TestSubscriptionService_TotalCost_Success(t *testing.T) {
	from := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC)
	userID := uuid.New()
	serviceName := "Netflix"

	var gotFilter repository.SumFilter
	repo := &fakeRepository{
		sumByPeriodFunc: func(ctx context.Context, filter repository.SumFilter) (int, error) {
			gotFilter = filter
			return 1000, nil
		},
	}
	svc := NewSubscriptionService(repo, discardLogger())

	total, err := svc.TotalCost(context.Background(), TotalCostFilter{
		UserID:      &userID,
		ServiceName: &serviceName,
		PeriodFrom:  from,
		PeriodTo:    to,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if total != 1000 {
		t.Fatalf("expected total 1000, got %d", total)
	}
	if gotFilter.UserID == nil || *gotFilter.UserID != userID {
		t.Fatal("expected user_id filter to be passed through")
	}
	if gotFilter.ServiceName == nil || *gotFilter.ServiceName != serviceName {
		t.Fatal("expected service_name filter to be passed through")
	}
	if !gotFilter.PeriodFrom.Equal(from) || !gotFilter.PeriodTo.Equal(to) {
		t.Fatal("expected period to be passed through")
	}
}
