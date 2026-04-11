package models

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// ToPgTime converts a string (HH:MM:SS) to pgtype.Time
func ToPgTime(value string) (pgtype.Time, error) {
	t, err := time.Parse("15:04:05", value)
	if err != nil {
		return pgtype.Time{}, err
	}

	// Calculate Microseconds since midnight
	usec := t.Sub(time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())).Microseconds()
	return pgtype.Time{Microseconds: usec, Valid: true}, nil
}

// ToPgDate converts a string (YYYY-MM-DD) to pgtype.Date
func ToPgDate(value string) (pgtype.Date, error) {
	t, err := time.Parse("2006-01-02", value)
	if err != nil {
		return pgtype.Date{}, err
	}

	return pgtype.Date{Time: t, Valid: true}, nil
}
