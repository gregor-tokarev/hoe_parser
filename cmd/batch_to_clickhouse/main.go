package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gregor-tokarev/hoe_parser/internal/clickhouse"
	"github.com/gregor-tokarev/hoe_parser/internal/config"
	"github.com/gregor-tokarev/hoe_parser/internal/scraper"
	listing "github.com/gregor-tokarev/hoe_parser/proto"
)

func main() {
	fmt.Println("Starting Batch ClickHouse Processing Example...")

	// Load configuration from environment variables
	cfg := config.Load()
	fmt.Printf("Loaded configuration: ClickHouse Host=%s, Port=%d, Database=%s\n",
		cfg.ClickHouse.Host, cfg.ClickHouse.Port, cfg.ClickHouse.Database)

	// Create ClickHouse adapter using configuration
	chConfig := clickhouse.FromMainConfig(cfg, false) // Set debug to false for batch processing

	adapter, err := clickhouse.NewAdapter(chConfig)
	if err != nil {
		log.Fatalf("Failed to create ClickHouse adapter: %v", err)
	}
	defer adapter.Close()

	fmt.Println("✅ Connected to ClickHouse successfully!")

	// Print initial stats
	printStats(adapter)

	// Example 1: Get listing links and batch process them
	fmt.Println("\n🔍 Getting listing links from gold scraper...")
	goldScraper := scraper.NewIntimcityGoldScraper()

	// Get all listing links (this will scrape all pages - use with caution)
	// For demonstration, we'll limit the results afterwards
	fmt.Println("⚠️  This will scrape a few pages - please be patient...")
	allLinks, err := goldScraper.ScrapeAllListingLinks()
	if err != nil {
		log.Fatalf("Failed to get listing links: %v", err)
	}

	fmt.Printf("Found %d total listing links\n", len(allLinks))

	// Limit to first 5 links for demonstration
	maxLinks := 5
	var links []scraper.ListingLink
	if len(allLinks) > maxLinks {
		links = allLinks[:maxLinks]
		fmt.Printf("Limited to first %d links for demonstration\n", maxLinks)
	} else {
		links = allLinks
	}

	// Scrape individual listings
	fmt.Println("\n📄 Scraping individual listings...")
	intimcityScraper := scraper.NewIntimcityScraper()

	var listings []*listing.Listing
	var sourceURLs []string

	for i, link := range links {
		fmt.Printf("Scraping %d/%d: %s\n", i+1, len(links), link.URL)

		scrapedListing, err := intimcityScraper.ScrapeListing(link.URL)
		if err != nil {
			log.Printf("❌ Failed to scrape %s: %v", link.URL, err)
			continue
		}

		listings = append(listings, scrapedListing)
		sourceURLs = append(sourceURLs, link.URL)
	}

	fmt.Printf("✅ Successfully scraped %d listings\n", len(listings))

	// Batch insert into ClickHouse
	if len(listings) > 0 {
		fmt.Println("\n💾 Batch inserting listings into ClickHouse...")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		start := time.Now()
		err = adapter.BatchInsertListings(ctx, listings, sourceURLs)
		if err != nil {
			log.Fatalf("Failed to batch insert listings: %v", err)
		}

		elapsed := time.Since(start)
		fmt.Printf("✅ Successfully batch inserted %d listings in %v\n", len(listings), elapsed)

		// Log changes for each listing
		for _, listing := range listings {
			err = adapter.LogChange(ctx, listing.Id, "created", "", "batch_inserted", "scraper", "batch_processor")
			if err != nil {
				log.Printf("⚠️ Failed to log change for listing %s: %v", listing.Id, err)
			}
		}
	}

	// Print final stats
	printStats(adapter)

	// Example 2: Demonstrate individual operations
	fmt.Println("\n🔍 Demonstrating individual operations...")

	if len(listings) > 0 {
		firstListing := listings[0]

		// Get listing by ID
		ctx := context.Background()
		retrieved, err := adapter.GetListingByID(ctx, firstListing.Id)
		if err != nil {
			log.Printf("❌ Failed to get listing by ID: %v", err)
		} else {
			fmt.Printf("✅ Retrieved listing: ID=%s, Age=%d, Price=%d RUB, City=%s\n",
				retrieved.ID, retrieved.PersonalAge, retrieved.PriceHour, retrieved.LocationCity)
		}

		// Update the listing (demonstrate upsert behavior)
		fmt.Println("📝 Updating listing...")
		err = adapter.UpdateListing(ctx, firstListing, sourceURLs[0])
		if err != nil {
			log.Printf("❌ Failed to update listing: %v", err)
		} else {
			fmt.Printf("✅ Updated listing %s\n", firstListing.Id)
		}
	}

	fmt.Println("\n🎉 Batch processing example completed!")
}

func printStats(adapter *clickhouse.Adapter) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stats, err := adapter.GetStats(ctx)
	if err != nil {
		log.Printf("❌ Failed to get stats: %v", err)
		return
	}

	fmt.Println("\n📊 Database Statistics:")
	fmt.Printf("  Total listings: %v\n", stats["total_listings"])
	fmt.Printf("  Listings with age: %v\n", stats["listings_with_age"])
	fmt.Printf("  Listings with price: %v\n", stats["listings_with_price"])
	fmt.Printf("  Listings with phone: %v\n", stats["listings_with_phone"])
	fmt.Printf("  Listings with photos: %v\n", stats["listings_with_photos"])
	if avgAge, ok := stats["avg_age"].(float64); ok && avgAge > 0 {
		fmt.Printf("  Average age: %.1f years\n", avgAge)
	}
	if avgPrice, ok := stats["avg_price_hour"].(float64); ok && avgPrice > 0 {
		fmt.Printf("  Average hourly price: %.0f RUB\n", avgPrice)
	}
	fmt.Printf("  Unique cities: %v\n", stats["unique_cities"])
}
