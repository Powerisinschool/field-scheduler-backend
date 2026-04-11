package services

import (
	"context"
	"field-scheduler-backend/internal/core/models"
	db "field-scheduler-backend/internal/db/sqlc"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// Define the interface for the handler to rely on
type ScheduleService interface {
	CreateScheduleEntry(ctx context.Context, userID pgtype.UUID, date pgtype.Date, startTime pgtype.Time, task string) (db.ScheduleEntry, error)
	ListScheduleEntries(ctx context.Context, userID *pgtype.UUID, startDate pgtype.Date, endDate pgtype.Date) ([]db.ScheduleEntry, error)
	SyncScheduleFromParsedEntries(ctx context.Context, parsedEntries []models.ParsedEntry, year int) error
}

type scheduleService struct {
	repo db.Querier // sqlc generated interface
}

// NewScheduleService creates a new instance of the service
func NewScheduleService(repo db.Querier) ScheduleService {
	return &scheduleService{
		repo: repo,
	}
}

func (s *scheduleService) CreateScheduleEntry(ctx context.Context, userID pgtype.UUID, date pgtype.Date, startTime pgtype.Time, task string) (db.ScheduleEntry, error) {
	// Business logic / validation would go here
	// userID is temporarily replacing authentication/authorization logic. In deployment, we will have a more robust system for handling user context and permissions.

	// Call the raw sqlc repository layer
	params := db.CreateScheduleEntryParams{
		ScheduleDate:    date,
		StartTime:       startTime,
		TaskDescription: task,
	}

	return s.repo.CreateScheduleEntry(ctx, params)
}

func (s *scheduleService) ListScheduleEntries(ctx context.Context, userID *pgtype.UUID, startDate pgtype.Date, endDate pgtype.Date) ([]db.ScheduleEntry, error) {
	// Business logic / validation would go here
	// userID can be used to filter entries for private entries, if needed. For now, we will ignore it and return all entries in the date range.

	// Call the raw sqlc repository layer
	params := db.ListScheduleEntriesParams{
		StartDate: startDate,
		EndDate:   endDate,
	}

	return s.repo.ListScheduleEntries(ctx, params)
}

func (s *scheduleService) SyncScheduleFromParsedEntries(ctx context.Context, parsedEntries []models.ParsedEntry, year int) error {
	// Default to current year if not provided
	if year <= 0 {
		year = time.Now().Year()
	}

	if len(parsedEntries) == 0 {
		return fmt.Errorf("no entries found in PDF")
	}

	var minDate, maxDate time.Time
	isFirst := true

	// 1. Pre-process the entries and find the date range
	for _, entry := range parsedEntries {
		// Parse Date (e.g., "1-Mar" -> "1-Mar-2026")
		dateStr := fmt.Sprintf("%s-%d", entry.Date, year)
		parsedDate, err := time.Parse("2-Jan-2006", dateStr)
		if err != nil {
			continue // Skip unparseable dates
		}

		if isFirst {
			minDate = parsedDate
			maxDate = parsedDate
			isFirst = false
		} else {
			if parsedDate.Before(minDate) {
				minDate = parsedDate
			}
			if parsedDate.After(maxDate) {
				maxDate = parsedDate
			}
		}
	}

	// 2. Delete existing entries in this authoritative date range
	err := s.repo.DeleteScheduleEntriesByDateRange(ctx, db.DeleteScheduleEntriesByDateRangeParams{
		ScheduleDate:   pgtype.Date{Time: minDate, Valid: true},
		ScheduleDate_2: pgtype.Date{Time: maxDate, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("failed to clear old schedule entries: %w", err)
	}

	// Cache for conductor name to ID mapping to minimize database lookups
	conductorCache := make(map[string]pgtype.UUID)
	venueCache := make(map[string]pgtype.UUID)

	// 3. Insert the new entries
	for _, entry := range parsedEntries {
		// Parse Date & Time
		dateStr := fmt.Sprintf("%s-%d", entry.Date, year)
		parsedDate, err := time.Parse("2-Jan-2006", dateStr)
		if err != nil {
			continue
		}
		parsedTime, err := time.Parse("15:04", entry.Time)
		if err != nil {
			// Fallback to midnight if time is unparseable
			parsedTime, _ = time.Parse("15:04", "00:00")
		}
		// Calculate microseconds for pgtype.Time
		usec := parsedTime.Sub(time.Date(parsedTime.Year(), parsedTime.Month(), parsedTime.Day(), 0, 0, 0, 0, parsedTime.Location())).Microseconds()

		// Attempt to resolve the Conductor ID by Name (Optional)
		conductorID := pgtype.UUID{Valid: false} // Default to NULL
		venueID := pgtype.UUID{Valid: false}     // Default to NULL
		if name := entry.Conductor; name != "" {
			if cachedID, ok := conductorCache[name]; ok {
				conductorID = cachedID
			} else {
				conductor, err := s.repo.GetConductorByName(ctx, name)
				if err != nil {
					// not found is an error, so only create if err != nil
					newConductor, err := s.repo.CreateConductor(ctx, name)
					if err != nil {
						fmt.Printf("Error creating conductor '%s': %v\n", entry.Conductor, err)
						conductorID = pgtype.UUID{Valid: false} // Keep as NULL if creation fails
					} else {
						conductorID = pgtype.UUID{Bytes: newConductor.Bytes, Valid: true}
					}
				} else {
					conductorID = pgtype.UUID{Bytes: conductor.Bytes, Valid: true}
				}
				conductorCache[name] = conductorID // Cache the result (even if not found, to avoid repeated lookups)
			}
		}
		if name := entry.Venue; name != "" {
			if cachedID, ok := venueCache[name]; ok {
				venueID = cachedID
			} else {
				venue, err := s.repo.GetVenueByName(ctx, name)
				if err != nil {
					// not found is an error, so only create if err != nil
					newVenue, err := s.repo.CreateVenue(ctx, db.CreateVenueParams{
						Name:     name,
						Location: pgtype.Text{String: name, Valid: true}, // Default location to name if not provided
					})
					if err != nil {
						fmt.Printf("Error creating venue '%s': %v\n", entry.Venue, err)
						venueID = pgtype.UUID{Valid: false} // Keep as NULL if creation fails
					} else {
						venueID = pgtype.UUID{Bytes: newVenue.Bytes, Valid: true}
					}
				} else {
					venueID = pgtype.UUID{Bytes: venue.Bytes, Valid: true}
				}
				venueCache[name] = venueID // Cache the result (even if not found, to avoid repeated lookups)
			}
		}

		// Insert
		_, err = s.repo.CreateScheduleEntry(ctx, db.CreateScheduleEntryParams{
			ScheduleDate:    pgtype.Date{Time: parsedDate, Valid: true},
			StartTime:       pgtype.Time{Microseconds: usec, Valid: true},
			TaskDescription: entry.Remark,
			ConductorID:     conductorID,
			VenueID:         venueID,
		})

		if err != nil {
			fmt.Printf("Warning: failed to insert entry for %s: %v\n", entry.Date, err)
		}
	}

	return nil
}
