package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"
)

type Station struct {
	Name        string `json:"name"`
	URL         string `json:"url"`
	Country     string `json:"country"`
	Codec       string `json:"codec"`
	Bitrate     int    `json:"bitrate"`
	Tags        string `json:"tags"`
	Language    string `json:"language"`
	Homepage    string `json:"homepage"`
	Favicon     string `json:"favicon"`
	ClickCount  int    `json:"clickcount"`
	LastCheckOK int    `json:"lastcheckok"`
}

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string, timeout time.Duration) *Client {
	if baseURL == "" {
		baseURL = "https://de1.api.radio-browser.info/json"
	}

	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout:   5 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext,
				MaxIdleConns:          100,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   5 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
		},
	}
}

func (c *Client) SearchStations(ctx context.Context, query string) ([]Station, error) {
	if query == "" {
		return nil, errors.New("query must not be empty")
	}

	endpoint := fmt.Sprintf("%s/stations/search?name=%s", c.baseURL, url.QueryEscape(query))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("User-Agent", "RadioTerminal/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("performing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	var stations []Station
	if err := json.NewDecoder(resp.Body).Decode(&stations); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return stations, nil
}
