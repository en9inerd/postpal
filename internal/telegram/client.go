package telegram

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/en9inerd/go-pkgs/httpclient"
	"github.com/en9inerd/go-pkgs/retry"
	"github.com/en9inerd/go-pkgs/validator"
)

const (
	// BaseURL is the base URL for Telegram Bot API
	BaseURL = "https://api.telegram.org/bot"
)

// Client represents a Telegram Bot API client
type Client struct {
	httpClient *httpclient.Client
	botToken   string
	logger     *slog.Logger
}

// NewClient creates a new Telegram Bot API client
func NewClient(botToken string, logger *slog.Logger) *Client {
	if logger == nil {
		logger = slog.Default()
	}

	baseURL := fmt.Sprintf("%s%s/", BaseURL, botToken)

	return &Client{
		httpClient: httpclient.New().
			WithBaseURL(baseURL).
			WithLogger(logger).
			WithTimeout(30 * time.Second).
			WithHeader("Content-Type", "application/json"),
		botToken: botToken,
		logger:   logger,
	}
}

// WithHTTPClient allows setting a custom HTTP client
func (c *Client) WithHTTPClient(client *httpclient.Client) *Client {
	c.httpClient = client
	return c
}

// WithTimeout sets a custom timeout for HTTP requests
func (c *Client) WithTimeout(timeout time.Duration) *Client {
	c.httpClient = c.httpClient.WithTimeout(timeout)
	return c
}

// validateRequest validates a request if it implements the Validatable interface
func (c *Client) validateRequest(req interface{}) error {
	if validatable, ok := req.(validator.Validatable); ok {
		v := &validator.Validator{}
		validatable.Validate(v)
		if !v.Valid() {
			return fmt.Errorf("validation failed: %s", string(v.JSON()))
		}
	}
	return nil
}

// makeRequest makes an HTTP request to the Telegram Bot API with retry logic
func (c *Client) makeRequest(method string, payload interface{}) (*APIResponse, error) {
	// Validate request if it's validatable
	if err := c.validateRequest(payload); err != nil {
		return nil, err
	}

	// Configure retry strategy for transient failures
	strategy := retry.DefaultStrategy()
	strategy.MaxAttempts = 3
	strategy.InitialDelay = 1 * time.Second
	strategy.MaxDelay = 10 * time.Second
	// Only retry on network errors, not API errors
	strategy.RetryableErrors = retry.IsRetryableError

	var apiResp APIResponse
	err := retry.Do(context.Background(), strategy, func() error {
		c.logger.Debug("making telegram api request", "method", method)

		err := c.httpClient.PostJSON(context.Background(), method, payload, &apiResp)
		if err != nil {
			// Network errors will be retried automatically
			c.logger.Warn("telegram api request failed, retrying", "error", err, "method", method)
			return err
		}

		// Check if Telegram API returned an error
		if !apiResp.OK {
			// API errors are not retryable - return immediately
			return fmt.Errorf("telegram api error: %s (code: %d)",
				apiResp.Description, apiResp.ErrorCode)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("telegram api request failed: %w", err)
	}

	return &apiResp, nil
}
