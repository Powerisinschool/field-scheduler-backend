package services

import (
	"context"
	db "field-scheduler-backend/internal/db/sqlc"

	"github.com/jackc/pgx/v5/pgtype"
)

type VenueService interface {
	ListVenues(ctx context.Context) ([]db.Venue, error)
	GetVenuesMap(ctx context.Context) (map[pgtype.UUID]string, error) // Returns a map of venue_id to venue_name
}

type venueService struct {
	repo db.Querier
}

func NewVenueService(repo db.Querier) VenueService {
	return &venueService{
		repo: repo,
	}
}

func (v *venueService) ListVenues(ctx context.Context) ([]db.Venue, error) {
	return v.repo.ListVenues(ctx)
}

func (v *venueService) GetVenuesMap(ctx context.Context) (map[pgtype.UUID]string, error) {
	venueList, err := v.repo.ListVenues(ctx)
	if err != nil {
		return nil, err
	}

	venues := make(map[pgtype.UUID]string, len(venueList))

	for _, venue := range venueList {
		if venue.ID.Valid {
			venues[venue.ID] = venue.Name
		}
	}

	return venues, nil
}
