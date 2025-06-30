package request_client

import (
	"strings"
	"testing"
	"time"
)

func TestNewProxyClient(t *testing.T) {
	proxies := []string{
		"http://proxy1:8080",
		"http://proxy2:3128",
	}

	client := NewProxyClient(proxies, 10*time.Second)

	if client.GetProxyCount() != 2 {
		t.Errorf("Expected 2 proxies, got %d", client.GetProxyCount())
	}

	if client.GetCurrentProxyIndex() != 0 {
		t.Errorf("Expected initial index 0, got %d", client.GetCurrentProxyIndex())
	}
}

func TestRoundRobinSelection(t *testing.T) {
	proxies := []string{
		"http://proxy1:8080",
		"http://proxy2:3128",
		"http://proxy3:1080",
	}

	client := NewProxyClient(proxies, 10*time.Second)

	// Test round-robin selection
	first := client.getNextProxy()
	if first != "http://proxy1:8080" {
		t.Errorf("Expected first proxy to be proxy1, got %s", first)
	}

	second := client.getNextProxy()
	if second != "http://proxy2:3128" {
		t.Errorf("Expected second proxy to be proxy2, got %s", second)
	}

	third := client.getNextProxy()
	if third != "http://proxy3:1080" {
		t.Errorf("Expected third proxy to be proxy3, got %s", third)
	}

	// Should wrap around
	fourth := client.getNextProxy()
	if fourth != "http://proxy1:8080" {
		t.Errorf("Expected fourth proxy to wrap around to proxy1, got %s", fourth)
	}
}

func TestEmptyProxyList(t *testing.T) {
	client := NewProxyClient([]string{}, 10*time.Second)

	if client.GetProxyCount() != 0 {
		t.Errorf("Expected 0 proxies, got %d", client.GetProxyCount())
	}

	proxy := client.getNextProxy()
	if proxy != "" {
		t.Errorf("Expected empty proxy, got %s", proxy)
	}
}

func TestClientConfiguration(t *testing.T) {
	client := NewProxyClient([]string{"http://proxy:8080"}, 5*time.Second)

	// Test setting max retries
	client.SetMaxRetries(5)
	if client.maxRetries != 5 {
		t.Errorf("Expected max retries 5, got %d", client.maxRetries)
	}

	// Test setting fallback
	client.SetFallbackAllowed(false)
	if client.fallbackOK != false {
		t.Errorf("Expected fallback to be false, got %t", client.fallbackOK)
	}
}

func TestListProxies(t *testing.T) {
	originalProxies := []string{
		"http://proxy1:8080",
		"http://proxy2:3128",
	}

	client := NewProxyClient(originalProxies, 10*time.Second)
	listedProxies := client.ListProxies()

	if len(listedProxies) != len(originalProxies) {
		t.Errorf("Expected %d proxies, got %d", len(originalProxies), len(listedProxies))
	}

	for i, proxy := range originalProxies {
		if listedProxies[i] != proxy {
			t.Errorf("Expected proxy %s at index %d, got %s", proxy, i, listedProxies[i])
		}
	}

	// Ensure it's a copy (modifying returned slice shouldn't affect original)
	listedProxies[0] = "modified"
	if client.proxies[0] == "modified" {
		t.Error("ListProxies should return a copy, not the original slice")
	}
}

func TestGlobalClient(t *testing.T) {
	// Reset global client for testing
	ResetGlobalClient()

	// Should return a fallback client if not initialized
	client := GetGlobalClient()
	if client == nil {
		t.Error("GetGlobalClient should never return nil")
	}

	if client.GetProxyCount() != 0 {
		t.Errorf("Fallback client should have 0 proxies, got %d", client.GetProxyCount())
	}
}

func TestInvalidProxyURL(t *testing.T) {
	client := NewProxyClient([]string{"://invalid"}, 10*time.Second)

	_, err := client.createClient("://invalid")
	if err == nil {
		t.Error("Expected error for invalid proxy URL")
		return
	}

	if !strings.Contains(err.Error(), "invalid proxy URL") {
		t.Errorf("Expected error to mention invalid proxy URL, got: %s", err.Error())
	}
}

func TestProxyTriedOnceOnly(t *testing.T) {
	// Create a client with 3 proxies
	proxies := []string{
		"http://proxy1:8080",
		"http://proxy2:3128",
		"http://proxy3:1080",
	}

	client := NewProxyClient(proxies, 10*time.Second)

	// Test that getNextProxyIndex advances correctly and doesn't repeat
	indices := make([]int, 6) // Get 6 indices (2 full cycles)
	for i := 0; i < 6; i++ {
		indices[i] = client.getNextProxyIndex()
	}

	// Should cycle through 0,1,2,0,1,2
	expected := []int{0, 1, 2, 0, 1, 2}
	for i, expectedIdx := range expected {
		if indices[i] != expectedIdx {
			t.Errorf("Expected index %d at position %d, got %d", expectedIdx, i, indices[i])
		}
	}
}
