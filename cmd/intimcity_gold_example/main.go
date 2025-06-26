package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gregor-tokarev/hoe_parser/internal/scraper"
)

func main() {
	fmt.Println("Starting Intimcity Gold continuous scraper...")

	// Create a new scraper instance
	goldScraper := scraper.NewIntimcityGoldScraper()

	// Create a channel to receive new links
	linkChan := make(chan string, 100)

	// Create a channel to handle shutdown signals
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// Start a goroutine to handle incoming links
	go func() {
		linkCount := 0
		for link := range linkChan {
			linkCount++
			fmt.Printf("📥 New link #%d: %s\n", linkCount, link)

			// Here you could add logic to:
			// - Save links to database
			// - Send to another system
			// - Process with existing scraper
			// - etc.
		}
	}()

	// Start continuous monitoring in a goroutine
	go func() {
		fmt.Println("🔄 Starting continuous monitoring of https://a.intimcity.gold/...")
		err := goldScraper.StartContinuousMonitoring(linkChan)
		if err != nil {
			log.Printf("❌ Continuous monitoring failed: %v", err)
		}
	}()

	// Wait for shutdown signal
	fmt.Println("🚀 Scraper is running. Press Ctrl+C to stop...")
	<-signalChan

	fmt.Println("\n🛑 Shutdown signal received. Stopping scraper...")
	close(linkChan)
}
