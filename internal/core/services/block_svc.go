package services

import (
	"cmp"
	"context"
	"encoding/json"
	"field-scheduler-backend/internal/core/models"
	db "field-scheduler-backend/internal/db/sqlc"
	"fmt"
	"slices"
	"strconv"
	"unicode"
)

// NB: geometry types are 'Polygon' for blocks and 'LineString' for roads.

type BlockService interface {
	ListCards(ctx context.Context) ([]models.Card, error)
	ListBlocks(ctx context.Context) ([]models.Block, error)
	CreateBlock(ctx context.Context, name string, geometryType models.BlockGeometryType, coordinates [][]float64, cardName *string) (models.Block, error)
	DeleteBlockByID(ctx context.Context, id string) error
	DeleteBlocksByCard(ctx context.Context, cardName string) error
}

type blockService struct {
	repo db.Querier
}

func NewBlockService(repo db.Querier) BlockService {
	return &blockService{
		repo: repo,
	}
}

func (b *blockService) ListCards(ctx context.Context) ([]models.Card, error) {
	dbCards, err := b.repo.ListCards(ctx)
	if err != nil {
		return nil, err
	}
	var cards []models.Card
	for _, dbCard := range dbCards {
		card := models.Card{
			ID:   dbCard.CardName,
			Name: dbCard.CardName,
		}
		cards = append(cards, card)
	}
	slices.SortFunc(cards, func(a, b models.Card) int {
		valA, _ := strconv.Atoi(a.Name)
		valB, _ := strconv.Atoi(b.Name)
		return valA - valB
	})
	return cards, nil
}

func cardIDToNameMap(cards []db.Card) map[string]string {
	m := make(map[string]string)
	for _, card := range cards {
		m[card.ID.String()] = card.CardName
	}
	return m
}

// Helper to split "10A" into (10, "A")
func splitBlockName(s string) (int, string) {
	var i int
	for i < len(s) && unicode.IsDigit(rune(s[i])) {
		i++
	}
	num, _ := strconv.Atoi(s[:i])
	return num, s[i:]
}

func (b *blockService) ListBlocks(ctx context.Context) ([]models.Block, error) {
	dbBlocks, err := b.repo.ListBlocks(ctx)
	if err != nil {
		return nil, err
	}
	dbCards, err := b.repo.ListCards(ctx)
	if err != nil {
		return nil, err
	}
	var cardIDToName map[string]string = cardIDToNameMap(dbCards)

	var blocks []models.Block

	for _, dbBlock := range dbBlocks {
		var c [][]float64
		if err := json.Unmarshal(dbBlock.Coordinates, &c); err != nil {
			return nil, fmt.Errorf("error unmarshalling block coordinates: %w", err)
		}
		var blockCardID *string
		if dbBlock.CardID.Valid {
			cardName := cardIDToName[dbBlock.CardID.String()]
			blockCardID = &cardName
		} else {
			blockCardID = nil
		}
		block := models.Block{
			ID:           dbBlock.BlockName,
			Name:         dbBlock.BlockName,
			GeometryType: models.BlockGeometryType(dbBlock.GeometryType),
			Coordinates:  c, // Assuming this is stored as JSONB in the DB and scanned into [][]float64
			CardID:       blockCardID,
		}
		blocks = append(blocks, block)
	}
	slices.SortFunc(blocks, func(a, b models.Block) int {
		// Split name into numeric part and letter part
		numA, letA := splitBlockName(a.Name)
		numB, letB := splitBlockName(b.Name)

		// 1. Compare the numbers first
		if numA != numB {
			return cmp.Compare(numA, numB)
		}
		// 2. If numbers are equal, compare the letters (e.g., 10A vs. 10B)
		return cmp.Compare(letA, letB)
	})
	return blocks, nil
}

// CreateBlock creates a new block with the given name, geometry type, and coordinates. It validates the geometry type and stores the coordinates as JSONB in the database.
func (b *blockService) CreateBlock(ctx context.Context, name string, geometryType models.BlockGeometryType, coordinates [][]float64, cardName *string) (models.Block, error) {
	// Validate geometry type
	if !geometryType.IsValid() {
		return models.Block{}, fmt.Errorf("invalid geometry type: %s", geometryType)
	}

	// Convert coordinates to JSONB format for storage
	coordsJSON, err := json.Marshal(coordinates)
	if err != nil {
		return models.Block{}, fmt.Errorf("error marshalling coordinates: %w", err)
	}

	var dbCard db.Card
	if cardName != nil {
		dbCard, err = b.repo.GetOrCreateCard(ctx, *cardName)
		if err != nil {
			return models.Block{}, fmt.Errorf("error creating card in DB: %w", err)
		}
	}

	// Call the repository layer to create the block
	params := db.CreateBlockParams{
		BlockName:    name,
		GeometryType: string(geometryType),
		Coordinates:  coordsJSON,
		CardID:       dbCard.ID,
	}
	dbBlock, err := b.repo.CreateBlock(ctx, params)
	if err != nil {
		return models.Block{}, fmt.Errorf("error creating block in DB: %w", err)
	}

	return models.Block{
		ID:           dbBlock.ID.String(),
		Name:         dbBlock.BlockName,
		GeometryType: models.BlockGeometryType(dbBlock.GeometryType),
		Coordinates:  coordinates,
	}, nil
}

func (b *blockService) DeleteBlockByID(ctx context.Context, id string) error {
	//var uuid pgtype.UUID
	//err := uuid.Scan(id)
	//if err != nil {
	//	return fmt.Errorf("error scanning UUID: %w", err)
	//}
	err := b.repo.DeleteBlockByID(ctx, id)
	if err != nil {
		return fmt.Errorf("error deleting block from DB: %w", err)
	}
	return nil
}

func (b *blockService) DeleteBlocksByCard(ctx context.Context, cardName string) error {
	err := b.repo.DeleteBlocksByCardName(ctx, cardName)
	if err != nil {
		return fmt.Errorf("error deleting block from DB: %w", err)
	}
	return nil
}
