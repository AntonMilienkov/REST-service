package model

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// MonthYearLayout — формат "MM-YYYY", используется и в JSON, и в query-параметрах периода.
const MonthYearLayout = "01-2006"

// MonthYear хранит дату в формате "MM-YYYY" (день всегда 01).
type MonthYear struct {
	time.Time
}

// ParseMonthYear парсит строку вида "MM-YYYY" в time.Time.
func ParseMonthYear(s string) (time.Time, error) {
	t, err := time.Parse(MonthYearLayout, s)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date %q, expected format MM-YYYY: %w", s, err)
	}
	return t, nil
}

func (m MonthYear) MarshalJSON() ([]byte, error) {
	return []byte(`"` + m.Format(MonthYearLayout) + `"`), nil
}

func (m *MonthYear) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), `"`)
	if s == "" || s == "null" {
		return nil
	}

	t, err := ParseMonthYear(s)
	if err != nil {
		return err
	}

	m.Time = t
	return nil
}

// Subscription — запись о подписке пользователя на сервис.
type Subscription struct {
	ID          uuid.UUID  `json:"id"`
	ServiceName string     `json:"service_name" example:"Yandex Plus"`
	Price       int        `json:"price" example:"400"`
	UserID      uuid.UUID  `json:"user_id"`
	StartDate   MonthYear  `json:"start_date" swaggertype:"string" example:"07-2025"`
	EndDate     *MonthYear `json:"end_date,omitempty" swaggertype:"string" example:"12-2025"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}
