package scraper

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"unicode/utf8"

	"github.com/PuerkitoBio/goquery"
	listing "github.com/gregor-tokarev/hoe_parser/proto"
)

// IntimcityScraper handles scraping of intimcity listings
type IntimcityScraper struct {
}

// NewIntimcityScraper creates a new intimcity scraper
func NewIntimcityScraper() *IntimcityScraper {
	return &IntimcityScraper{}
}

// ScrapeListing scrapes a single listing from intimcity and returns protobuf model
func (s *IntimcityScraper) ScrapeListing(url string) (*listing.Listing, error) {
	doc, err := FetchAndParsePage(url)

	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Extract listing ID from URL
	listingID := s.extractListingID(url)

	// Create the listing object
	listingObj := &listing.Listing{
		Id:           listingID,
		PersonalInfo: s.extractPersonalInfo(doc),
		ContactInfo:  s.extractContactInfo(doc),
		PricingInfo:  s.extractPricingInfo(doc),
		ServiceInfo:  s.extractServiceInfo(doc),
		LocationInfo: s.extractLocationInfo(doc),
		Description:  s.extractDescription(doc),
		LastUpdated:  s.extractLastUpdated(doc),
		Photos:       s.extractPhotos(doc),
	}

	return listingObj, nil
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// cleanString ensures the string is valid UTF-8 and safe for protobuf
func cleanString(s string) string {
	if !utf8.ValidString(s) {
		s = strings.ToValidUTF8(s, "")
	}
	return strings.TrimSpace(s)
}

// extractListingID extracts the listing ID from URL
func (s *IntimcityScraper) extractListingID(url string) string {
	re := regexp.MustCompile(`anketa(\d+)\.htm`)
	matches := re.FindStringSubmatch(url)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// extractPersonalInfo extracts personal information from the page
func (s *IntimcityScraper) extractPersonalInfo(doc *goquery.Document) *listing.PersonalInfo {
	info := &listing.PersonalInfo{}

	// Get the main content text
	mainText := doc.Text()

	// Extract age - try multiple patterns, prioritizing more specific ones
	agePatterns := []string{
		`Возраст\s*(\d+)`,      // "Возраст 32"
		`(\d+)\s+(?:года|лет)`, // "32 года", "25 лет"
		`возраст[:\s]*(\d+)`,   // Generic age patterns
	}
	for _, pattern := range agePatterns {
		re := regexp.MustCompile(`(?i)` + pattern)
		if matches := re.FindStringSubmatch(mainText); len(matches) > 1 {
			if age, err := strconv.Atoi(matches[1]); err == nil && age > 16 && age < 80 {
				info.Age = int32(age)
				break
			}
		}
	}

	// Extract height - looking for "167 см" patterns
	heightPatterns := []string{
		`Рост\s*(\d+)`,    // "Рост 167"
		`(\d+)\s*см`,      // "167 см"
		`рост[:\s]*(\d+)`, // Generic height patterns
	}
	for _, pattern := range heightPatterns {
		re := regexp.MustCompile(`(?i)` + pattern)
		if matches := re.FindStringSubmatch(mainText); len(matches) > 1 {
			if height, err := strconv.Atoi(matches[1]); err == nil && height > 140 && height < 220 {
				info.Height = int32(height)
				break
			}
		}
	}

	// Extract weight - looking for "50 кг" patterns
	weightPatterns := []string{
		`Вес\s*(\d+)`,    // "Вес 50"
		`(\d+)\s*кг`,     // "50 кг"
		`вес[:\s]*(\d+)`, // Generic weight patterns
	}
	for _, pattern := range weightPatterns {
		re := regexp.MustCompile(`(?i)` + pattern)
		if matches := re.FindStringSubmatch(mainText); len(matches) > 1 {
			if weight, err := strconv.Atoi(matches[1]); err == nil && weight > 30 && weight < 150 {
				info.Weight = int32(weight)
				break
			}
		}
	}

	// Extract breast size - looking for "3 размер", "размер 3" patterns
	breastPatterns := []string{
		`Грудь\s*(\d+)`,     // "Грудь 3"
		`(\d+)\s*размер`,    // "3 размер"
		`размер[:\s]*(\d+)`, // "размер 3"
	}
	for _, pattern := range breastPatterns {
		re := regexp.MustCompile(`(?i)` + pattern)
		if matches := re.FindStringSubmatch(mainText); len(matches) > 1 {
			if size, err := strconv.Atoi(matches[1]); err == nil && size > 0 && size < 10 {
				info.BreastSize = int32(size)
				break
			}
		}
	}

	return info
}

// extractContactInfo extracts contact information
func (s *IntimcityScraper) extractContactInfo(doc *goquery.Document) *listing.ContactInfo {
	info := &listing.ContactInfo{}

	// Extract phone number - multiple approaches
	// Method 1: Look for tel: links
	doc.Find("a[href^='tel:']").Each(func(i int, sel *goquery.Selection) {
		if href, exists := sel.Attr("href"); exists {
			phone := strings.TrimPrefix(href, "tel:")
			phone = cleanString(phone)
			if phone != "" {
				info.Phone = phone
			}
		}
	})

	// Method 2: Look for phone patterns in text if not found
	if info.Phone == "" {
		mainText := doc.Text()
		phonePatterns := []string{
			`\+7\s*\(\d{3}\)\s*\d{3}-\d{2}-\d{2}`,
			`\+7\s*\d{3}\s*\d{3}\s*\d{2}\s*\d{2}`,
			`8\s*\(\d{3}\)\s*\d{3}-\d{2}-\d{2}`,
			`\+7\s*\(\d{3}\)\s*\d{3}\s*\d{2}\s*\d{2}`,
		}
		for _, pattern := range phonePatterns {
			re := regexp.MustCompile(pattern)
			if match := re.FindString(mainText); match != "" {
				info.Phone = cleanString(match)
				break
			}
		}
	}

	// Method 3: Look for phone in HTML content more broadly
	if info.Phone == "" {
		// Look for any sequence that looks like a phone number
		re := regexp.MustCompile(`(?:\+7|8)\s*[\(\s]*\d{3}[\)\s]*\s*\d{3}[\s-]*\d{2}[\s-]*\d{2}`)
		if match := re.FindString(doc.Text()); match != "" {
			info.Phone = cleanString(match)
		}
	}

	// Check for messaging services in text
	mainTextLower := strings.ToLower(doc.Text())
	if strings.Contains(mainTextLower, "telegram") || strings.Contains(mainTextLower, "тг") {
		info.Telegram = "available"
	}
	if strings.Contains(mainTextLower, "whatsapp") || strings.Contains(mainTextLower, "вотсап") {
		info.WhatsappAvailable = true
	}
	if strings.Contains(mainTextLower, "viber") || strings.Contains(mainTextLower, "вайбер") {
		info.ViberAvailable = true
	}

	return info
}

// extractPricingInfo extracts pricing information
func (s *IntimcityScraper) extractPricingInfo(doc *goquery.Document) *listing.PricingInfo {
	info := &listing.PricingInfo{
		DurationPrices: make(map[string]int32),
		ServicePrices:  make(map[string]int32),
		Currency:       "RUB",
	}

	// Look for pricing in table cells and structured data
	doc.Find("td, div, span").Each(func(i int, sel *goquery.Selection) {
		text := strings.TrimSpace(sel.Text())

		// Look for duration pricing patterns
		durationPatterns := map[string]string{
			"час":    `(?:1\s*)?час[:\s]*(\d+)`,
			"2 часа": `2\s*час[:\s]*(\d+)`,
			"ночь":   `ночь[:\s]*(\d+)`,
			"день":   `день[:\s]*(\d+)`,
		}

		for duration, pattern := range durationPatterns {
			re := regexp.MustCompile(`(?i)` + pattern)
			if matches := re.FindStringSubmatch(text); len(matches) > 1 {
				if price, err := strconv.Atoi(matches[1]); err == nil {
					// Handle thousands
					if price < 1000 {
						price *= 1000
					}
					info.DurationPrices[duration] = int32(price)
				}
			}
		}

		// Look for simple number patterns that might be prices
		if strings.Contains(text, "000") || (len(text) > 2 && len(text) < 8) {
			re := regexp.MustCompile(`(\d+)\s*(?:000)?`)
			if matches := re.FindStringSubmatch(text); len(matches) > 1 {
				if price, err := strconv.Atoi(matches[1]); err == nil {
					if price >= 5 && price <= 100 { // Likely in thousands
						info.DurationPrices["base"] = int32(price * 1000)
					} else if price >= 1000 && price <= 100000 { // Already full price
						info.DurationPrices["base"] = int32(price)
					}
				}
			}
		}
	})

	// Extract service prices from links with "+" prefix
	doc.Find("a").Each(func(i int, sel *goquery.Selection) {
		text := sel.Text()
		if strings.Contains(text, "+") {
			re := regexp.MustCompile(`([^+]+)\+(\d+)`)
			if matches := re.FindStringSubmatch(text); len(matches) > 2 {
				serviceName := strings.TrimSpace(matches[1])
				if price, err := strconv.Atoi(matches[2]); err == nil {
					info.ServicePrices[serviceName] = int32(price)
				}
			}
		}
	})

	return info
}

// extractServiceInfo extracts available services
func (s *IntimcityScraper) extractServiceInfo(doc *goquery.Document) *listing.ServiceInfo {
	info := &listing.ServiceInfo{
		AvailableServices:  []string{},
		AdditionalServices: []string{},
		Restrictions:       []string{},
	}

	// Extract services from links (more specific selectors)
	doc.Find("a[href*='style'], a[href*='type']").Each(func(i int, sel *goquery.Selection) {
		serviceText := strings.TrimSpace(sel.Text())
		if serviceText != "" && len(serviceText) > 2 && len(serviceText) < 100 {
			// Clean up service name
			serviceName := strings.Split(serviceText, "+")[0]
			serviceName = cleanString(serviceName)

			if strings.Contains(serviceText, "+") {
				info.AdditionalServices = append(info.AdditionalServices, serviceName)
			} else {
				info.AvailableServices = append(info.AvailableServices, serviceName)
			}
		}
	})

	// Remove duplicates and clean up
	info.AvailableServices = removeDuplicates(info.AvailableServices)
	info.AdditionalServices = removeDuplicates(info.AdditionalServices)

	// Extract meeting type
	mainTextLower := strings.ToLower(doc.Text())
	if strings.Contains(mainTextLower, "квартир") || strings.Contains(mainTextLower, "дома") {
		info.MeetingType = "apartment"
	}
	if strings.Contains(mainTextLower, "отель") || strings.Contains(mainTextLower, "гостин") {
		info.MeetingType = "hotel"
	}
	if strings.Contains(mainTextLower, "выезд") {
		info.MeetingType = "outcall"
	}

	return info
}

// removeDuplicates removes duplicate strings from a slice
func removeDuplicates(slice []string) []string {
	keys := make(map[string]bool)
	var result []string

	for _, item := range slice {
		if !keys[item] && item != "" {
			keys[item] = true
			result = append(result, item)
		}
	}

	return result
}

// extractLocationInfo extracts location information
func (s *IntimcityScraper) extractLocationInfo(doc *goquery.Document) *listing.LocationInfo {
	info := &listing.LocationInfo{
		MetroStations: []string{},
		City:          "Moscow", // Default for intimcity
	}

	// Extract metro stations - multiple approaches
	// Method 1: Links with metro in href
	doc.Find("a[href*='metro']").Each(func(i int, sel *goquery.Selection) {
		station := strings.TrimSpace(sel.Text())
		if station != "" && len(station) > 2 {
			info.MetroStations = append(info.MetroStations, station)
		}
	})

	// Method 2: Look for metro station patterns in text
	if len(info.MetroStations) == 0 {
		// Common Moscow metro stations that might appear
		metroStations := []string{
			"Аэропорт", "Сокол", "ЦСКА", "Динамо", "Белорусская", "Маяковская",
			"Тверская", "Пушкинская", "Чистые пруды", "Красносельская",
			"Комсомольская", "Сокольники", "Преображенская", "Черкизовская",
		}

		mainText := doc.Text()
		for _, station := range metroStations {
			if strings.Contains(mainText, station) {
				info.MetroStations = append(info.MetroStations, station)
			}
		}
	}

	// Check for availability types
	mainTextLower := strings.ToLower(doc.Text())
	if strings.Contains(mainTextLower, "выезд") || strings.Contains(mainTextLower, "приезжаю") {
		info.OutcallAvailable = true
	}
	if strings.Contains(mainTextLower, "квартир") || strings.Contains(mainTextLower, "дома") || strings.Contains(mainTextLower, "принимаю") {
		info.IncallAvailable = true
	}

	return info
}

// extractDescription extracts the main description
func (s *IntimcityScraper) extractDescription(doc *goquery.Document) string {
	var descriptions []string

	// Look for description in various locations with more specific selectors
	doc.Find("td, div.content, div.description, p").Each(func(i int, sel *goquery.Selection) {
		text := strings.TrimSpace(sel.Text())
		// Look for longer text blocks that contain emojis or descriptive content
		if len(text) > 30 && (strings.Contains(text, "🔥") || strings.Contains(text, "❤") ||
			strings.Contains(text, "девочка") || strings.Contains(text, "услуг")) {
			descriptions = append(descriptions, text)
		}
	})

	// If no specific description found, get the largest text block
	if len(descriptions) == 0 {
		doc.Find("td").Each(func(i int, sel *goquery.Selection) {
			text := strings.TrimSpace(sel.Text())
			if len(text) > 50 {
				descriptions = append(descriptions, text)
			}
		})
	}

	if len(descriptions) > 0 {
		// Return the longest description
		longest := ""
		for _, desc := range descriptions {
			if len(desc) > len(longest) {
				longest = desc
			}
		}
		return longest
	}

	return ""
}

// extractLastUpdated extracts the last updated date
func (s *IntimcityScraper) extractLastUpdated(doc *goquery.Document) string {
	var lastUpdated string

	// Look for various date patterns
	doc.Find("td, div, span").Each(func(i int, sel *goquery.Selection) {
		text := sel.Text()
		// Look for date patterns
		datePatterns := []string{
			`(\d{2}\.\d{2}\.\d{4})`,
			`(\d{1,2}\/\d{1,2}\/\d{4})`,
			`(\d{4}-\d{2}-\d{2})`,
		}

		for _, pattern := range datePatterns {
			re := regexp.MustCompile(pattern)
			if matches := re.FindStringSubmatch(text); len(matches) > 1 {
				lastUpdated = matches[1]
				return
			}
		}
	})

	return lastUpdated
}

// extractPhotos extracts photo URLs
func (s *IntimcityScraper) extractPhotos(doc *goquery.Document) []string {
	var photos []string

	doc.Find("img").Each(func(i int, sel *goquery.Selection) {
		if src, exists := sel.Attr("src"); exists {
			// Filter out system images, icons, etc.
			if strings.Contains(src, "jpg") || strings.Contains(src, "png") ||
				strings.Contains(src, "jpeg") || strings.Contains(src, "webp") {
				// Skip small system icons
				if !strings.Contains(src, "icon") && !strings.Contains(src, "logo") {
					// Convert relative URLs to absolute
					if strings.HasPrefix(src, "/") {
						src = "https://a.intimcity.gold" + src
					}
					photos = append(photos, src)
				}
			}
		}
	})

	return photos
}
