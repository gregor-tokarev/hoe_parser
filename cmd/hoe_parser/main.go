package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gregor-tokarev/hoe_parser/internal/clickhouse"
	"github.com/gregor-tokarev/hoe_parser/internal/config"
	"github.com/gregor-tokarev/hoe_parser/internal/scraper"
	listing "github.com/gregor-tokarev/hoe_parser/proto"
)

func main() {
	fmt.Println("Starting ClickHouse Adapter Example...")

	// Load configuration from environment variables
	cfg := config.Load()
	fmt.Printf("Loaded configuration: ClickHouse Host=%s, Port=%d, Database=%s\n",
		cfg.ClickHouse.Host, cfg.ClickHouse.Port, cfg.ClickHouse.Database)

	// Create ClickHouse adapter using configuration
	chConfig := clickhouse.FromMainConfig(cfg, cfg.Debug)

	adapter, err := clickhouse.NewAdapter(chConfig)
	if err != nil {
		log.Fatalf("Failed to create ClickHouse adapter: %v", err)
	}
	defer adapter.Close()

	fmt.Println("✅ Connected to ClickHouse successfully!")

	// Create scrapers
	intimcityScraper := scraper.NewIntimcityScraper()
	goldScraper := scraper.NewIntimcityGoldScraper()

	// Create channel for shutdown signals
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// Channel to receive new listing URLs
	linkChan := make(chan string, 100)

	// Context for the entire application (no timeout)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start gold scraper monitoring in a goroutine
	go func() {
		fmt.Println("🔄 Starting continuous gold scraper monitoring...")
		err := goldScraper.StartContinuousMonitoring(linkChan)
		if err != nil {
			log.Printf("❌ Gold scraper monitoring failed: %v", err)
		}
	}()

	// Function to retry ClickHouse operations
	retryInsert := func(listing *listing.Listing, sourceURL string, maxRetries int) error {
		for attempt := 1; attempt <= maxRetries; attempt++ {
			// Create a context with timeout for this specific operation
			opCtx, opCancel := context.WithTimeout(ctx, 30*time.Second)

			err := adapter.InsertListing(opCtx, listing, sourceURL)
			opCancel()

			if err == nil {
				return nil
			}

			if attempt < maxRetries {
				log.Printf("⚠️  Attempt %d/%d failed for listing %s, retrying in %ds: %v",
					attempt, maxRetries, listing.Id, attempt*2, err)
				time.Sleep(time.Duration(attempt*2) * time.Second)
			} else {
				return fmt.Errorf("failed after %d attempts: %w", maxRetries, err)
			}
		}
		return nil
	}

	// Process incoming links and save to ClickHouse
	go func() {
		for {
			select {
			case link := <-linkChan:
				go func(link string) {
					fmt.Printf("📥 Processing link: %s\n", link)

					// Scrape the individual listing
					listing, err := intimcityScraper.ScrapeListing(link)

					if err != nil {
						log.Printf("❌ Failed to scrape listing %s: %v", link, err)
						return
					}

					// Insert into ClickHouse with retry logic
					err = retryInsert(listing, link, 3)
					if err != nil {
						return
					}
				}(link)

			case <-ctx.Done():
				fmt.Println("🛑 Processing stopped")
				return
			}
		}
	}()

	fmt.Println("🚀 ClickHouse adapter is running. Press Ctrl+C to stop...")
	<-signalChan

	fmt.Println("\n🛑 Shutdown signal received. Stopping...")
	cancel()

	// Give goroutines a moment to clean up
	time.Sleep(2 * time.Second)
	fmt.Println("✅ Shutdown complete")
}
