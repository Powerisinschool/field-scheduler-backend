package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

// The structure matching the Python Pandas JSON output
type ParsedEntry struct {
	Day       string `json:"Day"`
	Date      string `json:"Date"`      // e.g. "1-Mar"
	Time      string `json:"Time"`      // e.g. "10:55"
	Venue     string `json:"Venue"`     // e.g. "Kingdom Hall"
	Conductor string `json:"Conductor"` // e.g. "Chinuji Ogwu"
	Remark    string `json:"Remark"`    // e.g. "House-to-House"
}

type ParserClient interface {
	ParsePDF(ctx context.Context, fileData []byte) ([]ParsedEntry, error)
	FetchMapBlocks(ctx context.Context) (*MapBlocksResponse, error)
}

type parserClient struct {
	pythonURL  string
	httpClient *http.Client
}

func NewParserClient(pythonURL string) ParserClient {
	return &parserClient{
		pythonURL:  pythonURL,
		httpClient: &http.Client{},
	}
}

func (p *parserClient) ParsePDF(ctx context.Context, fileData []byte) ([]ParsedEntry, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Create the file form field
	part, err := writer.CreateFormFile("file", "schedule.pdf")
	if err != nil {
		return nil, err
	}
	part.Write(fileData)
	writer.Close()

	// Make the request to the Python service (assume endpoint is /parse/pdf)
	req, err := http.NewRequestWithContext(ctx, "POST", p.pythonURL+"/parse/pdf", body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to reach python service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("python service returned %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var entries []ParsedEntry
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		return nil, fmt.Errorf("failed to decode python response: %w", err)
	}

	return entries, nil
}

// The structure matching the python GeoJson output
type MapBlocksResponse struct {
	Address     string        `json:"address"`
	BlocksCount int           `json:"blocks_count"`
	Polygons    [][][]float64 `json:"polygons"` // Nested array: [Polygon][Point][Lat/Lon]
}

func (p *parserClient) FetchMapBlocks(ctx context.Context) (*MapBlocksResponse, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", p.pythonURL+"/fetch/map-blocks", nil)
	if err != nil {
		return nil, err
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to reach python service: %w", err)
	}
	defer resp.Body.Close()

	var result MapBlocksResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode python response: %w", err)
	}

	return &result, nil
}
