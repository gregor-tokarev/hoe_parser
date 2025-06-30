package request_client

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// ProxyClient represents an HTTP client with round-robin proxy support
type ProxyClient struct {
	proxies    []string
	currentIdx int
	mutex      sync.Mutex
	timeout    time.Duration
	maxRetries int
	fallbackOK bool // whether to allow requests without proxy if all proxies fail
}

// NewProxyClient creates a new proxy client with round-robin selection
func NewProxyClient(proxies []string, timeout time.Duration) *ProxyClient {
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &ProxyClient{
		proxies:    proxies,
		currentIdx: 0,
		timeout:    timeout,
		maxRetries: 3,
		fallbackOK: false, // Allow fallback to no proxy if all proxies fail
	}
}

// SetMaxRetries sets the maximum number of retries per request
func (pc *ProxyClient) SetMaxRetries(retries int) {
	pc.maxRetries = retries
}

// SetFallbackAllowed sets whether requests can be made without proxy if all proxies fail
func (pc *ProxyClient) SetFallbackAllowed(allowed bool) {
	pc.fallbackOK = allowed
}

// getNextProxy returns the next proxy in round-robin fashion
func (pc *ProxyClient) getNextProxy() string {
	pc.mutex.Lock()
	defer pc.mutex.Unlock()

	if len(pc.proxies) == 0 {
		return ""
	}

	proxy := pc.proxies[pc.currentIdx]
	pc.currentIdx = (pc.currentIdx + 1) % len(pc.proxies)
	return proxy
}

// getNextProxyIndex returns the next proxy index in round-robin fashion and advances it
func (pc *ProxyClient) getNextProxyIndex() int {
	pc.mutex.Lock()
	defer pc.mutex.Unlock()

	if len(pc.proxies) == 0 {
		return 0
	}

	idx := pc.currentIdx
	pc.currentIdx = (pc.currentIdx + 1) % len(pc.proxies)
	return idx
}

// createClient creates an HTTP client with the specified proxy
func (pc *ProxyClient) createClient(proxyURL string) (*http.Client, error) {
	if proxyURL == "" {
		// No proxy
		return &http.Client{
			Timeout: pc.timeout,
		}, nil
	}

	proxyParsed, err := url.Parse(proxyURL)
	if err != nil {
		return nil, fmt.Errorf("invalid proxy URL %s: %w", proxyURL, err)
	}

	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyParsed),
	}

	return &http.Client{
		Transport: transport,
		Timeout:   pc.timeout,
	}, nil
}

// Get performs a GET request with proxy round-robin
func (pc *ProxyClient) Get(url string) (*http.Response, error) {
	return pc.Do("GET", url, nil, nil)
}

// Post performs a POST request with proxy round-robin
func (pc *ProxyClient) Post(url, contentType string, body io.Reader) (*http.Response, error) {
	headers := map[string]string{
		"Content-Type": contentType,
	}
	return pc.Do("POST", url, body, headers)
}

// Do performs an HTTP request with proxy round-robin and retry logic
func (pc *ProxyClient) Do(method, url string, body io.Reader, headers map[string]string) (*http.Response, error) {
	var lastErr error

	// Try with proxies first - try each proxy exactly once without skipping any
	if len(pc.proxies) > 0 {
		// Get starting index for this request (advances round-robin for next request)
		startIdx := pc.getNextProxyIndex()

		// Try all proxies starting from the selected index
		for i := 0; i < len(pc.proxies); i++ {
			proxyIdx := (startIdx + i) % len(pc.proxies)
			proxy := pc.proxies[proxyIdx]

			resp, err := pc.doRequestWithProxy(method, url, body, headers, proxy)
			if err == nil {
				return resp, nil
			}
			lastErr = err
		}
	}

	// If all proxies failed and fallback is allowed, try without proxy
	if pc.fallbackOK {
		resp, err := pc.doRequestWithProxy(method, url, body, headers, "")
		if err == nil {
			return resp, nil
		}
		lastErr = err
	}

	if lastErr != nil {
		return nil, fmt.Errorf("all proxy attempts failed, last error: %w", lastErr)
	}

	return nil, fmt.Errorf("no working proxy found and fallback disabled")
}

// doRequestWithProxy performs a single HTTP request with the specified proxy
func (pc *ProxyClient) doRequestWithProxy(method, url string, body io.Reader, headers map[string]string, proxyURL string) (*http.Response, error) {
	client, err := pc.createClient(proxyURL)
	if err != nil {
		return nil, err
	}

	// If body is provided, we need to handle it carefully for retries
	var bodyBytes []byte
	if body != nil {
		bodyBytes, err = io.ReadAll(body)
		if err != nil {
			return nil, fmt.Errorf("failed to read request body: %w", err)
		}
	}

	for attempt := 0; attempt < pc.maxRetries; attempt++ {
		var reqBody io.Reader
		if bodyBytes != nil {
			reqBody = strings.NewReader(string(bodyBytes))
		}

		req, err := http.NewRequest(method, url, reqBody)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		// Add headers
		for key, value := range headers {
			req.Header.Set(key, value)
		}

		// Add default headers
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

		resp, err := client.Do(req)
		if err == nil {
			return resp, nil
		}

		// If this is the last attempt, return the error
		if attempt == pc.maxRetries-1 {
			proxyInfo := "no proxy"
			if proxyURL != "" {
				proxyInfo = fmt.Sprintf("proxy %s", proxyURL)
			}
			return nil, fmt.Errorf("request failed with %s after %d attempts: %w", proxyInfo, pc.maxRetries, err)
		}
	}

	return nil, fmt.Errorf("unexpected end of retry loop")
}

// GetProxyCount returns the number of configured proxies
func (pc *ProxyClient) GetProxyCount() int {
	return len(pc.proxies)
}

// GetCurrentProxyIndex returns the current proxy index
func (pc *ProxyClient) GetCurrentProxyIndex() int {
	pc.mutex.Lock()
	defer pc.mutex.Unlock()
	return pc.currentIdx
}

// ListProxies returns a copy of the proxy list
func (pc *ProxyClient) ListProxies() []string {
	result := make([]string, len(pc.proxies))
	copy(result, pc.proxies)
	return result
}
