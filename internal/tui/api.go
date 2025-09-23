package tui

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/zacksfF/evm-tvl-aggregator/internal/tui/models"
)

// APIClient handles communication with the TVL Aggregator API
type APIClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewAPIClient creates a new API client
func NewAPIClient(baseURL string) (*APIClient, error) {
	if baseURL == "" {
		return nil, fmt.Errorf("base URL cannot be empty")
	}

	return &APIClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// get performs a GET request to the specified endpoint
func (c *APIClient) get(endpoint string) ([]byte, error) {
	url := fmt.Sprintf("%s%s", c.baseURL, endpoint)
	
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return body, nil
}

// GetHealth checks the API health
func (c *APIClient) GetHealth() (*models.HealthResponse, error) {
	data, err := c.get("/api/v1/health")
	if err != nil {
		return nil, err
	}

	var health models.HealthResponse
	if err := json.Unmarshal(data, &health); err != nil {
		return nil, fmt.Errorf("failed to unmarshal health response: %w", err)
	}

	return &health, nil
}

// GetTotalTVL retrieves the total TVL across all protocols
func (c *APIClient) GetTotalTVL() (*models.TVLResponse, error) {
	data, err := c.get("/api/v1/tvl")
	if err != nil {
		return nil, err
	}

	var tvl models.TVLResponse
	if err := json.Unmarshal(data, &tvl); err != nil {
		return nil, fmt.Errorf("failed to unmarshal TVL response: %w", err)
	}

	return &tvl, nil
}

// GetProtocolTVL retrieves TVL for a specific protocol
func (c *APIClient) GetProtocolTVL(protocol string) (*models.ProtocolTVLResponse, error) {
	endpoint := fmt.Sprintf("/api/v1/tvl/%s", protocol)
	data, err := c.get(endpoint)
	if err != nil {
		return nil, err
	}

	var protocolTVL models.ProtocolTVLResponse
	if err := json.Unmarshal(data, &protocolTVL); err != nil {
		return nil, fmt.Errorf("failed to unmarshal protocol TVL response: %w", err)
	}

	return &protocolTVL, nil
}

// GetProtocols retrieves the list of available protocols
func (c *APIClient) GetProtocols() (*models.ProtocolsResponse, error) {
	data, err := c.get("/api/v1/protocols")
	if err != nil {
		return nil, err
	}

	var protocols models.ProtocolsResponse
	if err := json.Unmarshal(data, &protocols); err != nil {
		return nil, fmt.Errorf("failed to unmarshal protocols response: %w", err)
	}

	return &protocols, nil
}

// GetChains retrieves the list of supported chains
func (c *APIClient) GetChains() (*models.ChainsResponse, error) {
	data, err := c.get("/api/v1/chains")
	if err != nil {
		return nil, err
	}

	var chains models.ChainsResponse
	if err := json.Unmarshal(data, &chains); err != nil {
		return nil, fmt.Errorf("failed to unmarshal chains response: %w", err)
	}

	return &chains, nil
}

// GetStats retrieves system statistics
func (c *APIClient) GetStats() (*models.StatsResponse, error) {
	data, err := c.get("/api/v1/stats")
	if err != nil {
		return nil, err
	}

	var stats models.StatsResponse
	if err := json.Unmarshal(data, &stats); err != nil {
		return nil, fmt.Errorf("failed to unmarshal stats response: %w", err)
	}

	return &stats, nil
}

// GetHistoricalTVL retrieves historical TVL data for a protocol
func (c *APIClient) GetHistoricalTVL(protocol, chain, period string) (*models.HistoricalTVLResponse, error) {
	endpoint := fmt.Sprintf("/api/v1/tvl/%s/history", protocol)
	
	// Add query parameters
	if chain != "" || period != "" {
		endpoint += "?"
		if chain != "" {
			endpoint += fmt.Sprintf("chain=%s", chain)
		}
		if period != "" {
			if chain != "" {
				endpoint += "&"
			}
			endpoint += fmt.Sprintf("period=%s", period)
		}
	}

	data, err := c.get(endpoint)
	if err != nil {
		return nil, err
	}

	var historical models.HistoricalTVLResponse
	if err := json.Unmarshal(data, &historical); err != nil {
		return nil, fmt.Errorf("failed to unmarshal historical TVL response: %w", err)
	}

	return &historical, nil
}