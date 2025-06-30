# Request Client Module

This module provides HTTP client functionality with round-robin proxy support for the hoe_parser application.

## Features

- **Round-robin proxy selection**: Automatically cycles through configured proxies
- **Retry logic**: Configurable retry attempts for failed requests
- **Fallback support**: Can fall back to direct connection if all proxies fail
- **Multiple proxy types**: Supports HTTP, HTTPS, and SOCKS5 proxies
- **Thread-safe**: Safe for concurrent use across goroutines
- **Global instance**: Centralized client configuration via environment variables

## Configuration

Set the `PROXIES` environment variable with comma-separated proxy URLs:

```bash
export PROXIES="http://proxy1:8080,http://user:pass@proxy2:3128,socks5://proxy3:1080"
```

### Supported Proxy Formats

- HTTP: `http://proxy.example.com:8080`
- HTTP with auth: `http://username:password@proxy.example.com:8080`
- SOCKS5: `socks5://proxy.example.com:1080`
- SOCKS5 with auth: `socks5://username:password@proxy.example.com:1080`

## Usage

### Global Client (Recommended)

The application automatically initializes a global client based on configuration:

```go
import "github.com/gregor-tokarev/hoe_parser/internal/modules/request_client"

// The global client is automatically used by service functions
client := request_client.GetGlobalClient()
resp, err := client.Get("https://example.com")
```

### Manual Client Creation

```go
proxies := []string{
    "http://proxy1:8080",
    "http://proxy2:3128",
}

client := request_client.NewProxyClient(proxies, 30*time.Second)
client.SetMaxRetries(3)
client.SetFallbackAllowed(true)

resp, err := client.Get("https://example.com")
```

### Available Methods

- `Get(url string)` - HTTP GET request
- `Post(url, contentType string, body io.Reader)` - HTTP POST request
- `Do(method, url string, body io.Reader, headers map[string]string)` - Custom HTTP request

## Integration

The module is automatically integrated into:

- `internal/service/fetch_and_parse.go` - For all page fetching operations
- All scrapers use this indirectly through the service layer

## Behavior

1. **Round-robin selection**: Each request uses the next proxy in the list
2. **Automatic retry**: Failed requests are retried with the same proxy
3. **Proxy fallthrough**: If a proxy fails, the next proxy is tried
4. **Direct fallback**: If all proxies fail and fallback is enabled, requests go direct
5. **User-Agent**: Automatically sets a realistic browser User-Agent header

## Error Handling

The client provides detailed error messages indicating:
- Which proxy was used
- How many attempts were made
- Whether fallback was attempted
- The underlying error cause

## Performance Considerations

- Concurrent requests automatically distribute across different proxies
- Failed proxies are temporarily skipped in the rotation
- Connection pooling is handled by Go's standard HTTP transport
- Timeout settings prevent hanging requests 