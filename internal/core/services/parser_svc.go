package services

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"field-scheduler-backend/internal/core/models"
	db "field-scheduler-backend/internal/db/sqlc"
	python_api "field-scheduler-backend/internal/infrastructure/python_api"
	"fmt"
	"sync"
	"time"
)

type ParserService struct {
	client   *python_api.Client
	mu       sync.RWMutex
	pdfCache map[string][]models.ParsedEntry
	mapCache *models.MapCacheWrapper
	repo     db.Querier
}

func NewParserService(client *python_api.Client, repo db.Querier) *ParserService {
	svc := &ParserService{
		client:   client,
		pdfCache: make(map[string][]models.ParsedEntry),
		repo:     repo,
	}
	// svc.tryLoadPersistentCache()

	return svc
}

func (s *ParserService) GetMapBlocks(ctx context.Context) (*models.MapBlocksResponse, error) {
	s.mu.RLock()
	if s.mapCache != nil && time.Since(s.mapCache.LastUpdated) < 14*24*time.Hour {
		defer s.mu.RUnlock()
		// time.Sleep(3 * time.Second) // Simulate processing delay for demonstration TODO: Remove this in production
		return &s.mapCache.Data, nil
	}
	s.mu.RUnlock()

	resp, err := s.client.CallFetchBlocks(ctx)
	if err != nil {
		return nil, err
	}

	// If response is nil, we should not update the cache
	if resp == nil {
		return nil, err
	}

	s.mu.Lock()
	s.mapCache = &models.MapCacheWrapper{
		Data:        *resp,
		LastUpdated: time.Now(),
	}
	s.mu.Unlock()
	return resp, nil
}

func (s *ParserService) GetParsedPDF(ctx context.Context, file []byte) ([]models.ParsedEntry, error) {
	return []models.ParsedEntry{}, nil
}

func (s *ParserService) SyncBlocksFromCSV(ctx context.Context, csvData []byte) error {
	reader := csv.NewReader(bytes.NewReader(csvData))
	records, err := reader.ReadAll()
	if err != nil {
		return err
	}

	// Maps to group streets by Card -> Block
	type blockData struct {
		CardName string
		Streets  []string
	}
	// key: block_name (e.g. "1A")
	blocksToProcess := make(map[string]*blockData)

	var currentCard string
	var currentBlock string

	// 1. Parse Jagged CSV
	// Assume row 0-2 are headers based on your excerpt
	for i, row := range records {
		if i < 3 || len(row) < 3 {
			continue
		}

		// Logic for merged cells in CSV
		if row[0] != "" {
			currentCard = row[0]
		}
		if row[1] != "" {
			currentBlock = row[1]
		}
		streetName := row[2]

		if currentBlock == "" || streetName == "" {
			continue
		}

		if _, ok := blocksToProcess[currentBlock]; !ok {
			blocksToProcess[currentBlock] = &blockData{
				CardName: currentCard,
				Streets:  []string{},
			}
		}
		blocksToProcess[currentBlock].Streets = append(blocksToProcess[currentBlock].Streets, streetName)
	}

	// 2. Process each block
	for blockName, data := range blocksToProcess {
		// a. Get Coordinates from Python
		geo, err := s.client.CallFetchBlockByStreets(ctx, data.Streets)
		if err != nil {
			fmt.Printf("Skipping block %s: %v\n", blockName, err)
			continue
		}

		// b. Database Transaction (Simplified)
		card, err := s.repo.GetOrCreateCard(ctx, data.CardName)
		if err != nil {
			return err
		}

		coordsJSON, _ := json.Marshal(geo.Coordinates)
		// cardId := uuid.NullUUID{UUID: card.ID.Bytes, Valid: true}

		_, err = s.repo.CreateBlock(ctx, db.CreateBlockParams{
			BlockName:    blockName,
			GeometryType: "Polygon",
			Coordinates:  coordsJSON,
			CardID:       card.ID,
		})
		if err != nil {
			return err
		}
	}

	return nil
}
