package request_client

import (
	"sync"
	"time"

	"github.com/gregor-tokarev/hoe_parser/internal/config"
)

var (
	globalClient *ProxyClient
	once         sync.Once
)

// InitGlobalClient initializes the global proxy client with configuration
func InitGlobalClient(cfg *config.Config) {
	once.Do(func() {
		globalClient = NewProxyClient(cfg.Proxies, 30*time.Second)
	})
}

// GetGlobalClient returns the global proxy client instance
// If not initialized, it returns a client with no proxies
func GetGlobalClient() *ProxyClient {
	if globalClient == nil {
		// Fallback to no-proxy client if not initialized
		return NewProxyClient([]string{}, 30*time.Second)
	}
	return globalClient
}

// ResetGlobalClient resets the global client (useful for testing)
func ResetGlobalClient() {
	globalClient = nil
	once = sync.Once{}
}
