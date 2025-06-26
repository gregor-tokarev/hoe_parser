package scraper

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// IntimcityGoldScraper handles scraping of intimcity.gold listings
type IntimcityGoldScraper struct {
	client  *http.Client
	baseURL string
}

// ListingLink represents a listing link with metadata
type ListingLink struct {
	URL   string
	Title string
	ID    string
}

// NewIntimcityGoldScraper creates a new intimcity gold scraper
func NewIntimcityGoldScraper() *IntimcityGoldScraper {
	return &IntimcityGoldScraper{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: "https://a.intimcity.gold",
	}
}

// ScrapeAllListingLinks scrapes all pages and returns all listing links
func (s *IntimcityGoldScraper) ScrapeAllListingLinks() ([]ListingLink, error) {
	var allLinks []ListingLink

	// First, get the total number of pages
	totalPages, err := s.getTotalPages()
	if err != nil {
		return nil, fmt.Errorf("failed to get total pages: %w", err)
	}

	fmt.Printf("Found %d total pages to scrape\n", totalPages)

	// Loop through all pages
	for page := 1; page <= totalPages; page++ {
		fmt.Printf("Scraping page %d/%d\n", page, totalPages)

		links, err := s.scrapePageLinks(page)
		if err != nil {
			fmt.Printf("Warning: failed to scrape page %d: %v\n", page, err)
			continue
		}

		allLinks = append(allLinks, links...)
	}

	fmt.Printf("Total listing links collected: %d\n", len(allLinks))
	return allLinks, nil
}

// getTotalPages extracts the total number of pages from the main page
func (s *IntimcityGoldScraper) getTotalPages() (int, error) {
	doc, err := FetchAndParsePage(s.baseURL)
	if err != nil {
		return 0, err
	}

	// Look for pagination - search for the pattern "1 2 3 4 5 ... 145 ������"
	var maxPage int

	// Try to find pagination links
	doc.Find("a").Each(func(i int, sel *goquery.Selection) {
		href, exists := sel.Attr("href")
		if !exists {
			return
		}

		// Look for page number patterns in href like "page=123" or similar
		pageNumRegex := regexp.MustCompile(`(?i)(?:page=|p=|страница=)(\d+)`)
		matches := pageNumRegex.FindStringSubmatch(href)
		if len(matches) > 1 {
			if pageNum, err := strconv.Atoi(matches[1]); err == nil && pageNum > maxPage {
				maxPage = pageNum
			}
		}

		// Also check the link text for numbers
		text := strings.TrimSpace(sel.Text())
		if numRegex := regexp.MustCompile(`^\d+$`); numRegex.MatchString(text) {
			if pageNum, err := strconv.Atoi(text); err == nil && pageNum > maxPage {
				maxPage = pageNum
			}
		}
	})

	// If we couldn't find pagination links, look for text patterns
	if maxPage == 0 {
		pageText := doc.Text()
		// Look for patterns like "1 2 3 4 5 ... 145"
		paginationRegex := regexp.MustCompile(`\d+\s+\d+\s+\d+\s+\d+\s+\d+\s*\.\.\.\s*(\d+)`)
		matches := paginationRegex.FindStringSubmatch(pageText)
		if len(matches) > 1 {
			maxPage, _ = strconv.Atoi(matches[1])
		}
	}

	// Fallback: assume at least 1 page if nothing found
	if maxPage == 0 {
		maxPage = 1
	}

	return maxPage, nil
}

// scrapePageLinks extracts listing links from a specific page
func (s *IntimcityGoldScraper) scrapePageLinks(pageNum int) ([]ListingLink, error) {
	var pageURL string
	if pageNum == 1 {
		pageURL = s.baseURL
	} else {
		// Try different pagination URL patterns
		pageURL = fmt.Sprintf("%s/?page=%d", s.baseURL, pageNum)
	}

	doc, err := FetchAndParsePage(pageURL)
	if err != nil {
		// Try alternative pagination format
		pageURL = fmt.Sprintf("%s/p%d", s.baseURL, pageNum)
		doc, err = FetchAndParsePage(pageURL)
		if err != nil {
			return nil, err
		}
	}

	var links []ListingLink

	// Look for listing links - these typically contain profile/listing information
	doc.Find("a").Each(func(i int, sel *goquery.Selection) {
		href, exists := sel.Attr("href")
		if !exists {
			return
		}

		// Clean up the href
		href = strings.TrimSpace(href)
		if href == "" {
			return
		}

		// Make sure it's a full URL
		if strings.HasPrefix(href, "/") {
			href = s.baseURL + href
		} else if !strings.HasPrefix(href, "http") {
			href = s.baseURL + "/" + href
		}

		// Filter for listing links - look for patterns that indicate individual listings
		if s.isListingLink(href) {
			title := strings.TrimSpace(sel.Text())
			if title == "" {
				// If no text, try to get title from attributes
				title, _ = sel.Attr("title")
			}

			// Extract ID from URL
			id := s.extractIDFromURL(href)

			link := ListingLink{
				URL:   href,
				Title: title,
				ID:    id,
			}

			links = append(links, link)
		}
	})

	// Remove duplicates
	links = s.removeDuplicateLinks(links)

	return links, nil
}

// isListingLink determines if a URL is likely a listing page
func (s *IntimcityGoldScraper) isListingLink(url string) bool {
	// Common patterns for listing pages
	listingPatterns := []string{
		`anketa\d+`,  // anketa123.htm
		`profile\d+`, // profile123
		`user\d+`,    // user123
		`girl\d+`,    // girl123
		`id\d+`,      // id123
		`listing\d+`, // listing123
	}

	urlLower := strings.ToLower(url)
	for _, pattern := range listingPatterns {
		if matched, _ := regexp.MatchString(pattern, urlLower); matched {
			return true
		}
	}

	// Also check for paths that look like individual listings
	// Avoid navigation/category links
	excludePatterns := []string{
		`page=`, `p=`, `category`, `search`, `filter`, `sort`,
		`login`, `register`, `contact`, `about`, `help`,
		`javascript:`, `mailto:`, `tel:`, `#`,
	}

	for _, pattern := range excludePatterns {
		if strings.Contains(urlLower, pattern) {
			return false
		}
	}

	return false
}

// extractIDFromURL extracts an ID from the listing URL
func (s *IntimcityGoldScraper) extractIDFromURL(url string) string {
	// Try various ID extraction patterns
	patterns := []string{
		`anketa(\d+)`,
		`profile(\d+)`,
		`user(\d+)`,
		`girl(\d+)`,
		`id(\d+)`,
		`listing(\d+)`,
		`/(\d+)/?$`, // ID at the end of path
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(url)
		if len(matches) > 1 {
			return matches[1]
		}
	}

	return ""
}

// removeDuplicateLinks removes duplicate links based on URL
func (s *IntimcityGoldScraper) removeDuplicateLinks(links []ListingLink) []ListingLink {
	seen := make(map[string]bool)
	var result []ListingLink

	for _, link := range links {
		if !seen[link.URL] {
			seen[link.URL] = true
			result = append(result, link)
		}
	}

	return result
}

// StartContinuousMonitoring starts continuous monitoring of all pages, sending new links to the channel
// It loops through all pages, and when it reaches the last page, it starts over from the first page
func (s *IntimcityGoldScraper) StartContinuousMonitoring(linkChan chan<- string) error {
	// Get total pages once at the start
	totalPages, err := s.getTotalPages()
	if err != nil {
		return fmt.Errorf("failed to get total pages: %w", err)
	}

	fmt.Printf("Starting continuous monitoring of %d pages...\n", totalPages)

	cycleCount := 0

	// Infinite loop through all pages
	for {
		cycleCount++
		fmt.Printf("\n=== Starting cycle %d ===\n", cycleCount)

		// Loop through all pages in this cycle
		for page := 1; page <= totalPages; page++ {
			fmt.Printf("Monitoring page %d/%d (cycle %d)\n", page, totalPages, cycleCount)

			links, err := s.scrapePageLinks(page)
			if err != nil {
				fmt.Printf("Warning: failed to scrape page %d: %v\n", page, err)
				continue
			}

			// Send new links to channel
			for _, link := range links {
				linkChan <- link.URL
			}
		}
	}
}

// StartContinuousMonitoringWithCallback starts continuous monitoring with a callback function for each new link
func (s *IntimcityGoldScraper) StartContinuousMonitoringWithCallback(callback func(string)) error {
	linkChan := make(chan string, 10000) // Buffered channel

	// Start a goroutine to handle incoming links
	go func() {
		for link := range linkChan {
			callback(link)
		}
	}()

	// Start the monitoring (this will block)
	return s.StartContinuousMonitoring(linkChan)
}

// GetListingLinks is a convenience method that returns just the URLs
func (s *IntimcityGoldScraper) GetListingLinks() ([]string, error) {
	links, err := s.ScrapeAllListingLinks()
	if err != nil {
		return nil, err
	}

	urls := make([]string, len(links))
	for i, link := range links {
		urls[i] = link.URL
	}

	return urls, nil
}
