// Package hyperliquid - API client functionality
package hyperliquid

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/hyperliquid-go/hyperliquid-go/hyperliquid/utils"
)

// API represents the HTTP API client for Hyperliquid
type API struct {
	baseURL    string
	client     *http.Client
	timeout    time.Duration
	logger     *log.Logger
}

// NewAPI creates a new API client instance
func NewAPI(baseURL string, timeout time.Duration) *API {
	if baseURL == "" {
		baseURL = utils.MainnetAPIURL
	}
	
	client := &http.Client{
		Timeout: timeout,
	}
	
	return &API{
		baseURL: baseURL,
		client:  client,
		timeout: timeout,
		logger:  log.New(log.Writer(), "[API] ", log.LstdFlags),
	}
}

// NewAPIWithClient creates a new API client with a custom HTTP client
func NewAPIWithClient(baseURL string, client *http.Client) *API {
	if baseURL == "" {
		baseURL = utils.MainnetAPIURL
	}
	
	return &API{
		baseURL: baseURL,
		client:  client,
		logger:  log.New(log.Writer(), "[API] ", log.LstdFlags),
	}
}

// Post sends a POST request to the specified URL path with the given payload
func (a *API) Post(urlPath string, payload interface{}) (interface{}, error) {
	return a.PostWithContext(context.Background(), urlPath, payload)
}

// PostWithContext sends a POST request with context support
func (a *API) PostWithContext(ctx context.Context, urlPath string, payload interface{}) (interface{}, error) {
	if payload == nil {
		payload = map[string]interface{}{}
	}
	
	url := a.baseURL + urlPath
	
	// Marshal payload to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}
	
	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	// Set headers
	req.Header.Set("Content-Type", "application/json")
	
	// Send request
	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	
	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	
	// Handle HTTP errors
	if err := a.handleException(resp, body); err != nil {
		return nil, err
	}
	
	// Parse JSON response
	var result interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return map[string]interface{}{
			"error": fmt.Sprintf("Could not parse JSON: %s", string(body)),
		}, nil
	}
	
	return result, nil
}

// handleException processes HTTP response errors and returns appropriate Go errors
func (a *API) handleException(resp *http.Response, body []byte) error {
	statusCode := resp.StatusCode
	
	// Success status codes
	if statusCode < 400 {
		return nil
	}
	
	// Client errors (4xx)
	if statusCode >= 400 && statusCode < 500 {
		var errorResponse map[string]interface{}
		if err := json.Unmarshal(body, &errorResponse); err != nil {
			// Could not parse JSON error response
			return &utils.ClientError{
				StatusCode:   statusCode,
				ErrorCode:    "",
				ErrorMessage: string(body),
				Header:       resp.Header,
				ErrorData:    nil,
			}
		}
		
		if errorResponse == nil {
			return &utils.ClientError{
				StatusCode:   statusCode,
				ErrorCode:    "",
				ErrorMessage: string(body),
				Header:       resp.Header,
				ErrorData:    nil,
			}
		}
		
		// Extract error details
		errorCode := ""
		errorMessage := ""
		errorData := errorResponse["data"]
		
		if code, ok := errorResponse["code"].(string); ok {
			errorCode = code
		}
		if msg, ok := errorResponse["msg"].(string); ok {
			errorMessage = msg
		}
		
		return &utils.ClientError{
			StatusCode:   statusCode,
			ErrorCode:    errorCode,
			ErrorMessage: errorMessage,
			Header:       resp.Header,
			ErrorData:    errorData,
		}
	}
	
	// Server errors (5xx)
	return &utils.ServerError{
		StatusCode: statusCode,
		Message:    string(body),
	}
}

// SetTimeout updates the client timeout
func (a *API) SetTimeout(timeout time.Duration) {
	a.timeout = timeout
	a.client.Timeout = timeout
}

// GetBaseURL returns the base URL being used
func (a *API) GetBaseURL() string {
	return a.baseURL
}

// SetBaseURL updates the base URL
func (a *API) SetBaseURL(baseURL string) {
	a.baseURL = baseURL
}
