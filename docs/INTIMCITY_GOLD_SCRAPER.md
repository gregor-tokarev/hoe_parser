# Intimcity Gold Scraper

This scraper is designed to collect listing links from https://a.intimcity.gold/ by looping through all pages of the main page. It supports both one-time scraping and continuous monitoring modes.

## Features

- **Page Discovery**: Automatically detects the total number of pages (up to 145+ pages)
- **Link Extraction**: Extracts all listing links from each page
- **Continuous Monitoring**: Loops through all pages continuously, starting over from page 1 after reaching the last page
- **Channel Integration**: Sends new links to Go channels for real-time processing
- **Deduplication**: Removes duplicate links across all cycles
- **Encoding Support**: Handles Russian text and Windows-1251 encoding
- **Export Options**: Saves results in both JSON and text formats (one-time mode)
- **Progress Tracking**: Shows progress while scraping
- **Rate Limiting**: Includes delays to be respectful to the server
- **Graceful Shutdown**: Supports signal handling for clean shutdown

## Usage

### One-Time Scraping (Legacy)

For one-time scraping of all pages:

```bash
# Note: This method is deprecated in favor of continuous monitoring
# But still available for compatibility
```

### Continuous Monitoring (Channel-based)

For continuous monitoring with channel integration:

```bash
# Build the continuous scraper
make gold-scraper

# Run the continuous scraper  
make run-gold-scraper

# Or run directly with Go
go run ./cmd/intimcity_gold_example
```

### Continuous Monitoring (Callback-based)

For continuous monitoring with callback functions:

```bash
# Build the callback scraper
make gold-callback-scraper

# Run the callback scraper
make run-gold-callback-scraper

# Or run directly with Go
go run ./cmd/intimcity_gold_callback_example
```

### Output Files

The scraper generates two output files:

1. **`intimcity_gold_links.json`** - Complete data with URLs, titles, and IDs
2. **`intimcity_gold_urls.txt`** - Simple text file with just URLs

### Example JSON Output

```json
[
  {
    "URL": "https://a.intimcity.gold/anketa12345.htm",
    "Title": "Анкета девушки",
    "ID": "12345"
  }
]
```

## Integration Examples

### Channel-based Continuous Monitoring

```go
package main

import (
    "fmt"
    "github.com/gregor-tokarev/hoe_parser/internal/scraper"
)

func main() {
    goldScraper := scraper.NewIntimcityGoldScraper()
    linkChan := make(chan string, 100)

    // Handle incoming links
    go func() {
        for link := range linkChan {
            fmt.Printf("New link: %s\n", link)
            // Process the link (save to DB, scrape details, etc.)
        }
    }()

    // Start continuous monitoring (blocks)
    err := goldScraper.StartContinuousMonitoring(linkChan)
    if err != nil {
        panic(err)
    }
}
```

### Callback-based Continuous Monitoring

```go
package main

import (
    "fmt"
    "github.com/gregor-tokarev/hoe_parser/internal/scraper"
)

func main() {
    goldScraper := scraper.NewIntimcityGoldScraper()
    
    // Start monitoring with callback (blocks)
    err := goldScraper.StartContinuousMonitoringWithCallback(func(link string) {
        fmt.Printf("Callback received: %s\n", link)
        
        // Example: Scrape the individual listing immediately
        intimcityScraper := scraper.NewIntimcityScraper()
        listing, err := intimcityScraper.ScrapeListing(link)
        if err != nil {
            fmt.Printf("Failed to scrape %s: %v\n", link, err)
            return
        }
        
        // Process the listing data...
        fmt.Printf("Scraped listing ID: %s\n", listing.Id)
    })
    
    if err != nil {
        panic(err)
    }
}
```

### Legacy One-Time Scraping

For backward compatibility, the old one-time scraping method is still available:

```go
package main

import (
    "fmt"
    "github.com/gregor-tokarev/hoe_parser/internal/scraper"
)

func main() {
    // Get listing links (one-time)
    goldScraper := scraper.NewIntimcityGoldScraper()
    links, err := goldScraper.ScrapeAllListingLinks()
    if err != nil {
        panic(err)
    }

    // Process links
    for _, link := range links {
        fmt.Printf("Found link: %s\n", link.URL)
    }
}
```

## API

### IntimcityGoldScraper

#### Methods

- `NewIntimcityGoldScraper() *IntimcityGoldScraper` - Creates a new scraper instance
- `ScrapeAllListingLinks() ([]ListingLink, error)` - Scrapes all pages once and returns listing links (legacy)
- `GetListingLinks() ([]string, error)` - Convenience method that returns just the URLs (legacy)
- `StartContinuousMonitoring(linkChan chan<- string) error` - Starts continuous monitoring, sending new links to channel
- `StartContinuousMonitoringWithCallback(callback func(string)) error` - Starts continuous monitoring with callback function

#### ListingLink Struct

```go
type ListingLink struct {
    URL   string // Full URL to the listing
    Title string // Title/name from the link text
    ID    string // Extracted ID from the URL
}
```

## Configuration

The scraper includes several configurable patterns for:
- **Listing Link Detection**: Patterns like `anketa\d+`, `profile\d+`, etc.
- **Page Number Extraction**: Multiple pagination URL patterns
- **ID Extraction**: Various ID patterns from URLs

## Error Handling

The scraper includes robust error handling:
- Continues scraping other pages if one page fails
- Handles encoding conversion errors gracefully
- Provides detailed error messages for debugging

## Performance

- **Rate Limiting**: 1-second delay between page requests
- **Memory Efficient**: Processes pages one at a time
- **Concurrent Safe**: HTTP client with proper timeouts
- **Encoding Optimized**: Efficient UTF-8 conversion

## Continuous Monitoring Behavior

The continuous monitoring feature works as follows:

1. **Initial Discovery**: Detects the total number of pages (e.g., 145 pages)
2. **Cycle Start**: Begins with page 1 and progresses through all pages
3. **Link Processing**: Sends only NEW links to the channel/callback (duplicates are filtered)
4. **Cycle Completion**: After reaching the last page, starts over from page 1
5. **Infinite Loop**: Continues indefinitely until the program is stopped
6. **Rate Limiting**: Includes 1-second delays between pages and 30-second delays between full cycles

### Example Timeline

```
Cycle 1: Pages 1→2→3→...→145 (finds 2000 new links)
Wait 30 seconds
Cycle 2: Pages 1→2→3→...→145 (finds 50 new links)  
Wait 30 seconds
Cycle 3: Pages 1→2→3→...→145 (finds 12 new links)
...continues forever...
```

## Notes

- The scraper respects the website's structure and includes appropriate delays
- It automatically handles Russian text encoding (Windows-1251 to UTF-8)
- Duplicate links are automatically removed across ALL cycles
- Progress is displayed during the scraping process
- Graceful shutdown is supported via SIGINT/SIGTERM signals
- Memory usage grows slowly as the duplicate tracking map expands 