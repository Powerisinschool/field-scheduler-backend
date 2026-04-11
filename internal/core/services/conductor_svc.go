package services

import (
	"context"
	db "field-scheduler-backend/internal/db/sqlc"

	"github.com/jackc/pgx/v5/pgtype"
)

type ConductorService interface {
	ListConductors(ctx context.Context) ([]db.Conductor, error)
	GetConductorsMap(ctx context.Context) (map[pgtype.UUID]string, error) // Returns a map of conductor_id to conductor_name
}

type conductorService struct {
	repo db.Querier
}

func NewConductorService(repo db.Querier) ConductorService {
	return &conductorService{
		repo: repo,
	}
}

func (c *conductorService) ListConductors(ctx context.Context) ([]db.Conductor, error) {
	return c.repo.ListConductors(ctx)
}

func (c *conductorService) GetConductorsMap(ctx context.Context) (map[pgtype.UUID]string, error) {
	conductorList, err := c.repo.ListConductors(ctx)
	if err != nil {
		return nil, err
	}

	conductors := make(map[pgtype.UUID]string, len(conductorList))

	for _, conductor := range conductorList {
		if conductor.ID.Valid {
			conductors[conductor.ID] = conductor.FullName
		}
	}

	return conductors, nil
}
