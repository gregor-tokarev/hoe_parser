package main

import (
	"fmt"
	"log"

	"github.com/gregor-tokarev/hoe_parser/internal/scraper"
	"google.golang.org/protobuf/encoding/protojson"
)

func main() {
	fmt.Println("Intimcity Scraper Example")

	// Create scraper
	scraperInstance := scraper.NewIntimcityScraper()

	// Scrape the listing
	url := "https://a.intimcity.gold/indi/anketa675508.htm"
	listing, err := scraperInstance.ScrapeListing(url)
	if err != nil {
		log.Fatalf("Failed to scrape listing: %v", err)
	}

	// Convert to JSON for pretty printing
	jsonData, err := protojson.MarshalOptions{
		Multiline: true,
		Indent:    "  ",
	}.Marshal(listing)
	if err != nil {
		log.Fatalf("Failed to marshal to JSON: %v", err)
	}

	fmt.Println("Scraped Listing Data:")
	fmt.Println(string(jsonData))

	// Example of accessing specific fields
	fmt.Printf("\nListing ID: %s\n", listing.Id)
	fmt.Printf("Phone: %s\n", listing.ContactInfo.Phone)
	fmt.Printf("Age: %d\n", listing.PersonalInfo.Age)
	fmt.Printf("Available Services: %v\n", listing.ServiceInfo.AvailableServices)
	fmt.Printf("Metro Stations: %v\n", listing.LocationInfo.MetroStations)

	if len(listing.PricingInfo.DurationPrices) > 0 {
		fmt.Println("Pricing:")
		for duration, price := range listing.PricingInfo.DurationPrices {
			fmt.Printf("  %s: %d %s\n", duration, price, listing.PricingInfo.Currency)
		}
	}
}
