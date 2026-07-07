package model

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

const monthYearLayout = "01-2006"

// MonthYear хранит дату в формате "MM-YYYY" (день всегда 01).
type MonthYear struct {
	time.Time
}

func (m MonthYear) MarshalJSON() ([]byte, error) {
	return []byte(`"` + m.Format(monthYearLayout) + `"`), nil
}

func (m *MonthYear) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), `"`)
	if s == "" || s == "null" {
		return nil
	}

	t, err := time.Parse(monthYearLayout, s)
	if err != nil {
		return fmt.Errorf("invalid date %q, expected format MM-YYYY: %w", s, err)
	}

	m.Time = t
	return nil
}

// Subscription — запись о подписке пользователя на сервис.
type Subscription struct {
	ID          uuid.UUID  `json:"id"`
	ServiceName string     `json:"service_name"`
	Price       int        `json:"price"`
	UserID      uuid.UUID  `json:"user_id"`
	StartDate   MonthYear  `json:"start_date"`
	EndDate     *MonthYear `json:"end_date,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}
