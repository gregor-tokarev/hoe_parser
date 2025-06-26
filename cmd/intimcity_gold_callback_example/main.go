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
	fmt.Println("Starting Intimcity Gold continuous scraper with callback...")

	// Create a new scraper instance
	goldScraper := scraper.NewIntimcityGoldScraper()

	// Create a channel to handle shutdown signals
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// Start continuous monitoring with callback in a goroutine
	go func() {
		linkCount := 0
		fmt.Println("🔄 Starting continuous monitoring with callback...")

		err := goldScraper.StartContinuousMonitoringWithCallback(func(link string) {
			linkCount++
			fmt.Printf("🔗 Callback received link #%d: %s\n", linkCount, link)

			// Example: Process the link immediately
			// You could:
			// - Parse the individual listing
			// - Save to database
			// - Send to message queue
			// - etc.

			// Example of using the existing scraper to parse the link
			if linkCount <= 5 { // Only process first 5 links as example
				fmt.Printf("  📄 Processing listing details for link #%d...\n", linkCount)
				intimcityScraper := scraper.NewIntimcityScraper()
				listing, err := intimcityScraper.ScrapeListing(link)
				if err != nil {
					fmt.Printf("  ❌ Failed to scrape listing: %v\n", err)
				} else {
					fmt.Printf("  ✅ Successfully parsed listing ID: %s\n", listing.Id)
					if listing.PersonalInfo != nil {
						fmt.Printf("     Age: %d, Height: %d, Weight: %d\n",
							listing.PersonalInfo.Age,
							listing.PersonalInfo.Height,
							listing.PersonalInfo.Weight)
					}
				}
			}
		})

		if err != nil {
			log.Printf("❌ Continuous monitoring failed: %v", err)
		}
	}()

	// Wait for shutdown signal
	fmt.Println("🚀 Scraper is running with callback. Press Ctrl+C to stop...")
	<-signalChan

	fmt.Println("\n🛑 Shutdown signal received. Stopping scraper...")
}
