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
	Url string
}

// NewIntimcityScraper creates a new intimcity scraper
func NewIntimcityScraper(url string) *IntimcityScraper {
	return &IntimcityScraper{Url: url}
}

// ScrapeListing scrapes a single listing from intimcity and returns protobuf model
func (s *IntimcityScraper) ScrapeListing() (*listing.Listing, error) {
	doc, err := FetchAndParsePage(s.Url)

	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Extract listing ID from URL
	listingID := s.extractListingID()

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
func (s *IntimcityScraper) extractListingID() string {
	re := regexp.MustCompile(`anketa(\d+)\.htm`)
	matches := re.FindStringSubmatch(s.Url)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// extractPersonalInfo extracts personal information from the page
func (s *IntimcityScraper) extractPersonalInfo(doc *goquery.Document) *listing.PersonalInfo {
	info := &listing.PersonalInfo{}

	// Extract name from page title
	if title := doc.Find("title").Text(); title != "" {
		info.Name = strings.TrimSpace(title)
	}

	// Extract using specific element IDs where available
	if age := doc.Find("#tdankage").Text(); age != "" {
		if ageVal, err := strconv.Atoi(strings.TrimSpace(age)); err == nil {
			info.Age = int32(ageVal)
		}
	}

	if height := doc.Find("#tdankhei").Text(); height != "" {
		if heightVal, err := strconv.Atoi(strings.TrimSpace(height)); err == nil {
			info.Height = int32(heightVal)
		}
	}

	if weight := doc.Find("#tdankwei").Text(); weight != "" {
		if weightVal, err := strconv.Atoi(strings.TrimSpace(weight)); err == nil {
			info.Weight = int32(weightVal)
		}
	}

	if breast := doc.Find("#tdankbre").Text(); breast != "" {
		if breastVal, err := strconv.Atoi(strings.TrimSpace(breast)); err == nil {
			info.BreastSize = int32(breastVal)
		}
	}

	if clothSize := doc.Find("#tdankcloth").Text(); clothSize != "" {
		info.BodyType = strings.TrimSpace(clothSize)
	}

	if haircut := doc.Find("#tdankinhc").Text(); haircut != "" {
		info.HairColor = strings.TrimSpace(haircut)
	}

	return info
}

// extractContactInfo extracts contact information
func (s *IntimcityScraper) extractContactInfo(doc *goquery.Document) *listing.ContactInfo {
	info := &listing.ContactInfo{}

	// Extract phone using specific ID first
	if phone := doc.Find("#tdmobphone a").First(); phone.Length() > 0 {
		if href, exists := phone.Attr("href"); exists && strings.HasPrefix(href, "tel:") {
			info.Phone = cleanString(strings.TrimPrefix(href, "tel:"))
		} else {
			info.Phone = cleanString(phone.Text())
		}
	}

	// Fallback to tel: links anywhere
	if info.Phone == "" {
		doc.Find("a[href^='tel:']").First().Each(func(i int, sel *goquery.Selection) {
			if href, exists := sel.Attr("href"); exists {
				info.Phone = cleanString(strings.TrimPrefix(href, "tel:"))
			}
		})
	}

	// Extract telegram from text patterns
	pageText := doc.Text()
	telegramPatterns := []string{
		`@([a-zA-Z0-9_]+)`,
		`[Tt]елеграм[мь]?\s*[@:]?\s*([a-zA-Z0-9_]+)`,
		`[Tt]г\s*[@:]?\s*([a-zA-Z0-9_]+)`,
	}

	for _, pattern := range telegramPatterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(pageText); len(matches) > 1 {
			username := strings.TrimSpace(matches[1])
			if len(username) > 2 && len(username) < 50 {
				info.Telegram = "@" + username
				break
			}
		}
	}

	// Check for messaging app availability
	if doc.Find("a[href*='whatsapp'], .sWhatsApp").Length() > 0 {
		info.WhatsappAvailable = true
	}

	if doc.Find("a[href*='telegram'], .sTelegram").Length() > 0 && info.Telegram == "" {
		info.Telegram = "available"
	}

	if strings.Contains(strings.ToLower(pageText), "viber") {
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

	// Use table.table-price class for pricing table
	pricingTable := doc.Find("table.table-price table.table-price-inner tbody")

	trs := pricingTable.Find("tr")

	apartments := trs.Eq(2)

	apartmentsDayHour := apartments.Find("td").Eq(1).Text()
	apartmentsDay2Hour := apartments.Find("td").Eq(2).Text()
	apartmentsNightHour := apartments.Find("td").Eq(3).Text()
	apartmentsNight2Hour := apartments.Find("td").Eq(4).Text()

	outcall := trs.Eq(3)
	outcallDayHour := outcall.Find("td").Eq(1).Text()
	outcallDay2Hour := outcall.Find("td").Eq(2).Text()
	outcallNightHour := outcall.Find("td").Eq(3).Text()
	outcallNight2Hour := outcall.Find("td").Eq(4).Text()

	info.DurationPrices["apartments_day_hour"] = extractPrice(apartmentsDayHour)
	info.DurationPrices["apartments_day_2hour"] = extractPrice(apartmentsDay2Hour)
	info.DurationPrices["apartments_night_hour"] = extractPrice(apartmentsNightHour)
	info.DurationPrices["apartments_night_2hour"] = extractPrice(apartmentsNight2Hour)

	info.DurationPrices["outcall_day_hour"] = extractPrice(outcallDayHour)
	info.DurationPrices["outcall_day_2hour"] = extractPrice(outcallDay2Hour)
	info.DurationPrices["outcall_night_hour"] = extractPrice(outcallNightHour)
	info.DurationPrices["outcall_night_2hour"] = extractPrice(outcallNight2Hour)

	// pricingTable.Find("tr").Each(func(i int, row *goquery.Selection) {
	// 	tds := row.Find("td")
	// 	if tds.Length() >= 2 {
	// 		label := strings.TrimSpace(tds.Eq(0).Text())
	// 		value := strings.TrimSpace(tds.Eq(1).Text())
	// 	}
	// })

	return info
}

// extractPrice helper function to parse price from text
func extractPrice(text string) int32 {
	numStr := strings.ReplaceAll(text, " ", "")
	numStr = strings.ReplaceAll(text, " ", "")
	numStr = strings.ReplaceAll(numStr, "\u2009", "")
	numStr = strings.ReplaceAll(numStr, "₽", "")

	if price, err := strconv.Atoi(numStr); err == nil {
		fmt.Println(price)
		return int32(price)
	}

	return 0
}

// extractServiceInfo extracts available services
func (s *IntimcityScraper) extractServiceInfo(doc *goquery.Document) *listing.ServiceInfo {
	info := &listing.ServiceInfo{
		AvailableServices:  []string{},
		AdditionalServices: []string{},
		Restrictions:       []string{},
	}

	// Use table.uslugi_block class for services table
	servicesTable := doc.Find("table.uslugi_block")
	if servicesTable.Length() == 0 {
		// Fallback to any table containing services info
		doc.Find("table").Each(func(i int, table *goquery.Selection) {
			if strings.Contains(table.Text(), "Секс") || strings.Contains(table.Text(), "Массаж") {
				servicesTable = table
				return
			}
		})
	}

	if servicesTable.Length() > 0 {
		// Extract services from checkboxes
		servicesTable.Find("input[type='checkbox']").Each(func(j int, checkbox *goquery.Selection) {
			// Find the service link next to checkbox
			serviceLink := checkbox.NextFiltered("a")
			if serviceLink.Length() == 0 {
				serviceLink = checkbox.NextAllFiltered("a").First()
			}

			var serviceName string
			if serviceLink.Length() > 0 {
				serviceName = strings.TrimSpace(serviceLink.Text())
			} else {
				// Extract from parent text
				parent := checkbox.Parent()
				serviceName = strings.TrimSpace(parent.Text())
				serviceName = strings.TrimPrefix(serviceName, "✓")
				serviceName = strings.TrimPrefix(serviceName, "☑")
				serviceName = strings.TrimSpace(serviceName)
			}

			if serviceName != "" && len(serviceName) > 2 && len(serviceName) < 100 {
				serviceName = cleanString(serviceName)

				// Check if checkbox is checked
				if _, checked := checkbox.Attr("checked"); checked {
					info.AvailableServices = append(info.AvailableServices, serviceName)
				} else {
					info.Restrictions = append(info.Restrictions, serviceName)
				}
			}
		})

		// Extract from service links with href patterns
		servicesTable.Find("a[href*='style'], a[href*='type']").Each(func(j int, link *goquery.Selection) {
			serviceName := strings.TrimSpace(link.Text())
			if serviceName != "" && len(serviceName) > 2 && len(serviceName) < 100 {
				serviceName = cleanString(serviceName)

				if strings.Contains(serviceName, "+") {
					parts := strings.Split(serviceName, "+")
					if len(parts) > 0 {
						serviceName = strings.TrimSpace(parts[0])
						info.AdditionalServices = append(info.AdditionalServices, serviceName)
					}
				} else {
					info.AvailableServices = append(info.AvailableServices, serviceName)
				}
			}
		})
	}

	// Determine meeting type
	pageText := strings.ToLower(doc.Text())
	if strings.Contains(pageText, "апартаменты") {
		info.MeetingType = "apartment"
	}
	if strings.Contains(pageText, "выезд") {
		if info.MeetingType != "" {
			info.MeetingType = "both"
		} else {
			info.MeetingType = "outcall"
		}
	}

	return info
}

// extractLocationInfo extracts location information
func (s *IntimcityScraper) extractLocationInfo(doc *goquery.Document) *listing.LocationInfo {
	info := &listing.LocationInfo{
		MetroStations: []string{},
		City:          "Moscow", // Default for intimcity
	}

	// Extract city using specific ID
	if city := doc.Find("#tdankcity").Text(); city != "" {
		info.City = strings.TrimSpace(city)
	}

	// Extract metro stations from links with metro in href
	doc.Find("a[href*='metro']").Each(func(i int, link *goquery.Selection) {
		station := strings.TrimSpace(link.Text())
		if station != "" && len(station) > 2 {
			info.MetroStations = append(info.MetroStations, station)
		}
	})

	// Extract district from links with district in href
	doc.Find("a[href*='district']").Each(func(i int, link *goquery.Selection) {
		district := strings.TrimSpace(link.Text())
		if district != "" {
			info.District = district
			return
		}
	})

	// Fallback to table parsing
	if len(info.MetroStations) == 0 || info.District == "" {
		doc.Find("table tr").Each(func(i int, row *goquery.Selection) {
			cells := row.Find("td")
			if cells.Length() >= 2 {
				label := strings.TrimSpace(cells.Eq(0).Text())

				if strings.Contains(label, "Метро") && len(info.MetroStations) == 0 {
					cells.Eq(1).Find("a").Each(func(j int, link *goquery.Selection) {
						station := strings.TrimSpace(link.Text())
						if station != "" && len(station) > 2 {
							info.MetroStations = append(info.MetroStations, station)
						}
					})
				}

				if strings.Contains(label, "Район") && info.District == "" {
					info.District = strings.TrimSpace(cells.Eq(1).Text())
				}
			}
		})
	}

	// Check availability from pricing table
	pageText := strings.ToLower(doc.Text())
	if strings.Contains(pageText, "выезд") {
		info.OutcallAvailable = true
	}
	if strings.Contains(pageText, "апартаменты") || strings.Contains(pageText, "принимаю") {
		info.IncallAvailable = true
	}

	return info
}

// extractDescription extracts the main description
func (s *IntimcityScraper) extractDescription(doc *goquery.Document) string {
	// Use p.pnletter class for description
	if desc := doc.Find("p.pnletter").First(); desc.Length() > 0 {
		return cleanString(desc.Text())
	}

	// Fallback to table cells with colspan="2"
	var descriptions []string
	doc.Find("table tr td[colspan='2']").Each(func(i int, cell *goquery.Selection) {
		text := strings.TrimSpace(cell.Text())
		if len(text) > 50 {
			descriptions = append(descriptions, text)
		}
	})

	if len(descriptions) > 0 {
		// Return the longest description
		longest := ""
		for _, desc := range descriptions {
			if len(desc) > len(longest) {
				longest = desc
			}
		}
		return cleanString(longest)
	}

	return ""
}

// extractLastUpdated extracts the last updated date
func (s *IntimcityScraper) extractLastUpdated(doc *goquery.Document) string {
	// Look for update date in table with noprint class
	updateText := doc.Find("tr.noprint td").Last().Text()
	if updateText != "" {
		re := regexp.MustCompile(`(\d{2}\.\d{2}\.\d{4})`)
		if matches := re.FindStringSubmatch(updateText); len(matches) > 1 {
			return matches[1]
		}
	}

	// Fallback to any date pattern in table
	var lastUpdated string
	doc.Find("table tr td").Each(func(i int, cell *goquery.Selection) {
		text := cell.Text()
		re := regexp.MustCompile(`(\d{2}\.\d{2}\.\d{4})`)
		if matches := re.FindStringSubmatch(text); len(matches) > 1 {
			lastUpdated = matches[1]
		}
	})

	return lastUpdated
}

// extractPhotos extracts photo URLs
func (s *IntimcityScraper) extractPhotos(doc *goquery.Document) []string {
	var photos []string

	imageData, err := FetchJsonImgs(s.Url)
	if err != nil {
		return photos
	}

	for _, img := range imageData {
		href := "https://a.intimcity.gold" + img.BIMG
		photos = append(photos, href)
	}

	return photos
}

// Helper function to check if slice contains string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
