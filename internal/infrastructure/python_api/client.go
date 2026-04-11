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
