package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type HTTPClient struct {
	client *http.Client
}

type HTTPRequest struct {
	Method  string
	URL     string
	Headers map[string]string
	Body    interface{}
	Timeout time.Duration
}

type HTTPResponse struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
}

func NewHTTPClient(timeout time.Duration) *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 20,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}
}

func (hc *HTTPClient) Do(req *HTTPRequest) (*HTTPResponse, error) {
	var body io.Reader
	if req.Body != nil {
		jsonData, err := json.Marshal(req.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		body = bytes.NewBuffer(jsonData)
	}

	httpReq, err := http.NewRequest(req.Method, req.URL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	if req.Body != nil {
		httpReq.Header.Set("Content-Type", "application/json")
	}

	ctx := context.Background()
	if req.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), req.Timeout)
		defer cancel()
	}
	httpReq = httpReq.WithContext(ctx)

	resp, err := hc.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return &HTTPResponse{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       responseBody,
	}, nil
}

func (hc *HTTPClient) Get(url string, headers map[string]string, timeout time.Duration) (*HTTPResponse, error) {
	req := &HTTPRequest{
		Method:  "GET",
		URL:     url,
		Headers: headers,
		Timeout: timeout,
	}
	return hc.Do(req)
}

func (hc *HTTPClient) Post(url string, body interface{}, headers map[string]string, timeout time.Duration) (*HTTPResponse, error) {
	req := &HTTPRequest{
		Method:  "POST",
		URL:     url,
		Headers: headers,
		Body:    body,
		Timeout: timeout,
	}
	return hc.Do(req)
}

func (hc *HTTPClient) Put(url string, body interface{}, headers map[string]string, timeout time.Duration) (*HTTPResponse, error) {
	req := &HTTPRequest{
		Method:  "PUT",
		URL:     url,
		Headers: headers,
		Body:    body,
		Timeout: timeout,
	}
	return hc.Do(req)
}

func (hc *HTTPClient) Delete(url string, headers map[string]string, timeout time.Duration) (*HTTPResponse, error) {
	req := &HTTPRequest{
		Method:  "DELETE",
		URL:     url,
		Headers: headers,
		Timeout: timeout,
	}
	return hc.Do(req)
}

func (hr *HTTPResponse) JSON(v interface{}) error {
	return json.Unmarshal(hr.Body, v)
}

func (hr *HTTPResponse) String() string {
	return string(hr.Body)
}

func (hr *HTTPResponse) IsSuccess() bool {
	return hr.StatusCode >= 200 && hr.StatusCode < 300
}

func (hr *HTTPResponse) IsClientError() bool {
	return hr.StatusCode >= 400 && hr.StatusCode < 500
}

func (hr *HTTPResponse) IsServerError() bool {
	return hr.StatusCode >= 500
}

type RetryableHTTPClient struct {
	*HTTPClient
	MaxRetries int
	Backoff    time.Duration
}

func NewRetryableHTTPClient(timeout time.Duration, maxRetries int, backoff time.Duration) *RetryableHTTPClient {
	return &RetryableHTTPClient{
		HTTPClient: NewHTTPClient(timeout),
		MaxRetries: maxRetries,
		Backoff:    backoff,
	}
}

func (rhc *RetryableHTTPClient) DoWithRetry(req *HTTPRequest) (*HTTPResponse, error) {
	var lastErr error
	
	for i := 0; i <= rhc.MaxRetries; i++ {
		resp, err := rhc.Do(req)
		if err == nil && resp.IsSuccess() {
			return resp, nil
		}
		
		if err != nil {
			lastErr = err
		} else if resp.IsServerError() {
			lastErr = fmt.Errorf("server error: %d", resp.StatusCode)
		} else {
			return resp, nil
		}
		
		if i < rhc.MaxRetries {
			time.Sleep(rhc.Backoff * time.Duration(i+1))
		}
	}
	
	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}