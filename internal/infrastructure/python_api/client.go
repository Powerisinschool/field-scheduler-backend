package pythonapi

import (
	"bytes"
	"context"
	"encoding/json"
	"field-scheduler-backend/internal/core/models"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

type BlockByStreetsRequest struct {
	Streets []string `json:"streets"`
}

type BlockByStreetsResponse struct {
	RequestedStreets []string    `json:"requested_streets"`
	MatchScore       float64     `json:"match_score"`
	Coordinates      [][]float64 `json:"coordinates"` // [[lat, lon], ...]
}

func (c *Client) CallParsePDF(ctx context.Context, data []byte) ([]models.ParsedEntry, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Create the file form field
	part, err := writer.CreateFormFile("file", "schedule.pdf")
	if err != nil {
		return nil, err
	}
	part.Write(data)
	writer.Close()

	// Make the request to the Python service (assume endpoint is /parse/pdf)
	req, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL+"/parse/pdf", body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to reach python service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("python service returned %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var entries []models.ParsedEntry
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		return nil, fmt.Errorf("failed to decode python response: %w", err)
	}

	return entries, nil
}

func (c *Client) CallFetchBlocks(ctx context.Context) (*models.MapBlocksResponse, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.BaseURL+"/fetch/map-blocks", nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to reach python service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("python service returned %d", resp.StatusCode)
	}

	var result models.MapBlocksResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode python response: %w", err)
	}

	return &result, nil
}

func (c *Client) CallFetchBlockByStreets(ctx context.Context, streets []string) (*BlockByStreetsResponse, error) {
	jsonData, _ := json.Marshal(BlockByStreetsRequest{Streets: streets})

	req, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL+"/fetch/block-by-topology", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("python service returned status %d", resp.StatusCode)
	}

	var result BlockByStreetsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode python response: %w", err)
	}

	return &result, nil
}
