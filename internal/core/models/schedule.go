package models

import (
	db "field-scheduler-backend/internal/db/sqlc"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
)

type ScheduleEntry struct {
	ID              string `json:"id"`
	ScheduleDate    string `json:"schedule_date"`
	StartTime       string `json:"start_time"`
	TaskDescription string `json:"task_description"`
	Location        string `json:"location,omitempty"`
	Conductor       string `json:"conductor_name,omitempty"`
	CardID          string `json:"card_id,omitempty"`
}

func pgTimeToString(pt pgtype.Time) string {
	if !pt.Valid {
		return ""
	}

	// Convert microseconds to seconds
	totalSeconds := pt.Microseconds / 1000000
	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60

	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}

func ToScheduleEntryModel(sqlcEntry db.ScheduleEntry, conductors map[pgtype.UUID]string, venues map[pgtype.UUID]string) ScheduleEntry {
	// get conductor from conductor_id if it exists, otherwise return empty string
	conductor := "Unknown Conductor"
	if sqlcEntry.ConductorID.Valid {
		if name, ok := conductors[sqlcEntry.ConductorID]; ok {
			conductor = name
		}
	}

	loc := "Unknown Venue"
	if sqlcEntry.VenueID.Valid {
		if _loc, ok := venues[sqlcEntry.VenueID]; ok {
			loc = _loc
		}
	}

	return ScheduleEntry{
		ID:              sqlcEntry.ID.String(),
		ScheduleDate:    sqlcEntry.ScheduleDate.Time.Format("2006-01-02"),
		StartTime:       pgTimeToString(sqlcEntry.StartTime),
		TaskDescription: sqlcEntry.TaskDescription,
		Location:        loc,
		Conductor:       conductor,
	}
}

func ToScheduleEntryModels(sqlcEntries []db.ScheduleEntry, conductors map[pgtype.UUID]string, venues map[pgtype.UUID]string) []ScheduleEntry {
	entries := make([]ScheduleEntry, len(sqlcEntries))
	for i, e := range sqlcEntries {
		entries[i] = ToScheduleEntryModel(e, conductors, venues)
	}
	return entries
}
