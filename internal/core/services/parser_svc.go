package services

import (
	"context"
	"field-scheduler-backend/internal/core/models"
	python_api "field-scheduler-backend/internal/infrastructure/python_api"
	"sync"
	"time"
)

type ParserService struct {
	client   *python_api.Client
	mu       sync.RWMutex
	pdfCache map[string][]models.ParsedEntry
	mapCache *models.MapCacheWrapper
}

func NewParserService(client *python_api.Client) *ParserService {
	svc := &ParserService{
		client:   client,
		pdfCache: make(map[string][]models.ParsedEntry),
	}
	// svc.tryLoadPersistentCache()

	return svc
}

func (s *ParserService) GetMapBlocks(ctx context.Context) (*models.MapBlocksResponse, error) {
	s.mu.RLock()
	if s.mapCache != nil && time.Since(s.mapCache.LastUpdated) < 14*24*time.Hour {
		defer s.mu.RUnlock()
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
