package clickhouse

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	mainConfig "github.com/gregor-tokarev/hoe_parser/internal/config"
	listing "github.com/gregor-tokarev/hoe_parser/proto"
)

// Config holds ClickHouse connection configuration
// This is compatible with the main config.ClickHouseConfig but adds Debug option
type Config struct {
	Host           string
	Port           int
	Database       string
	User           string
	Password       string
	MaxConnections int
	Debug          bool
}

// FromMainConfig creates a ClickHouse adapter Config from the main application config
func FromMainConfig(mainCfg *mainConfig.Config, debug bool) Config {
	return Config{
		Host:           mainCfg.ClickHouse.Host,
		Port:           mainCfg.ClickHouse.Port,
		Database:       mainCfg.ClickHouse.Database,
		User:           mainCfg.ClickHouse.User,
		Password:       mainCfg.ClickHouse.Password,
		MaxConnections: mainCfg.ClickHouse.MaxConnections,
		Debug:          debug,
	}
}

// Adapter handles ClickHouse operations for listings
type Adapter struct {
	conn   clickhouse.Conn
	config Config
}

// FlattenedListing represents a flattened listing structure for ClickHouse
type FlattenedListing struct {
	// Primary identification
	ID          string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	LastScraped time.Time
	SourceURL   string

	// Personal information
	PersonalName       string
	PersonalAge        uint8
	PersonalHeight     uint16
	PersonalWeight     uint16
	PersonalBreastSize uint8
	PersonalHairColor  string
	PersonalEyeColor   string
	PersonalBodyType   string

	// Contact information
	ContactPhone    string
	ContactTelegram string
	ContactEmail    string

	// Pricing information
	PricingCurrency string

	// Structured pricing - Apartments/Incall rates
	PriceApartmentsDayHour    uint32
	PriceApartmentsDay2Hour   uint32
	PriceApartmentsNightHour  uint32
	PriceApartmentsNight2Hour uint32

	// Structured pricing - Outcall rates
	PriceOutcallDayHour    uint32
	PriceOutcallDay2Hour   uint32
	PriceOutcallNightHour  uint32
	PriceOutcallNight2Hour uint32

	// Legacy/computed pricing fields for compatibility
	PriceHour   uint32
	Price2Hours uint32
	PriceNight  uint32
	PriceDay    uint32
	PriceBase   uint32

	// Additional pricing data (for any other price types)
	PricingDurationPrices map[string]uint32
	PricingServicePrices  map[string]uint32

	// Service information
	ServiceAvailable    []string
	ServiceAdditional   []string
	ServiceRestrictions []string
	ServiceMeetingType  string

	// Location information
	LocationMetroStations    []string
	LocationDistrict         string
	LocationCity             string
	LocationOutcallAvailable bool
	LocationIncallAvailable  bool

	// General information
	Description string
	LastUpdated string
	Photos      []string
	PhotosCount uint16
}

// NewAdapter creates a new ClickHouse adapter
func NewAdapter(config Config) (*Adapter, error) {
	// Set default MaxConnections if not specified
	maxConns := config.MaxConnections
	if maxConns <= 0 {
		maxConns = 10
	}

	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{fmt.Sprintf("%s:%d", config.Host, config.Port)},
		Auth: clickhouse.Auth{
			Database: config.Database,
			Username: config.User,
			Password: config.Password,
		},
		Debug: config.Debug,
		Debugf: func(format string, v ...interface{}) {
			if config.Debug {
				fmt.Printf("[ClickHouse Debug] "+format+"\n", v...)
			}
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		DialTimeout:      30 * time.Second,
		MaxOpenConns:     maxConns,
		MaxIdleConns:     maxConns / 2,
		ConnMaxLifetime:  time.Hour,
		ConnOpenStrategy: clickhouse.ConnOpenInOrder,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ClickHouse: %w", err)
	}

	// Test the connection
	if err := conn.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping ClickHouse: %w", err)
	}

	return &Adapter{
		conn:   conn,
		config: config,
	}, nil
}

// Close closes the ClickHouse connection
func (a *Adapter) Close() error {
	return a.conn.Close()
}

// FlattenListing converts a protobuf Listing to FlattenedListing
func (a *Adapter) FlattenListing(listing *listing.Listing, sourceURL string) *FlattenedListing {
	now := time.Now()

	flattened := &FlattenedListing{
		ID:          listing.Id,
		CreatedAt:   now,
		UpdatedAt:   now,
		LastScraped: now,
		SourceURL:   sourceURL,
		Description: listing.Description,
		LastUpdated: listing.LastUpdated,
		Photos:      listing.Photos,
		PhotosCount: uint16(len(listing.Photos)),
	}

	// Flatten personal info
	if listing.PersonalInfo != nil {
		flattened.PersonalName = listing.PersonalInfo.Name
		flattened.PersonalAge = uint8(listing.PersonalInfo.Age)
		flattened.PersonalHeight = uint16(listing.PersonalInfo.Height)
		flattened.PersonalWeight = uint16(listing.PersonalInfo.Weight)
		flattened.PersonalBreastSize = uint8(listing.PersonalInfo.BreastSize)
		flattened.PersonalHairColor = listing.PersonalInfo.HairColor
		flattened.PersonalEyeColor = listing.PersonalInfo.EyeColor
		flattened.PersonalBodyType = listing.PersonalInfo.BodyType
	}

	// Flatten contact info
	if listing.ContactInfo != nil {
		flattened.ContactPhone = listing.ContactInfo.Phone
		flattened.ContactTelegram = listing.ContactInfo.Telegram
		flattened.ContactEmail = listing.ContactInfo.Email
	}

	// Flatten pricing info
	if listing.PricingInfo != nil {
		flattened.PricingCurrency = listing.PricingInfo.Currency
		if flattened.PricingCurrency == "" {
			flattened.PricingCurrency = "RUB"
		}

		// Convert duration prices map
		flattened.PricingDurationPrices = make(map[string]uint32)
		for k, v := range listing.PricingInfo.DurationPrices {
			flattened.PricingDurationPrices[k] = uint32(v)
		}

		// Extract structured pricing fields directly
		if price, exists := listing.PricingInfo.DurationPrices["apartments_day_hour"]; exists {
			flattened.PriceApartmentsDayHour = uint32(price)
		}
		if price, exists := listing.PricingInfo.DurationPrices["apartments_day_2hour"]; exists {
			flattened.PriceApartmentsDay2Hour = uint32(price)
		}
		if price, exists := listing.PricingInfo.DurationPrices["apartments_night_hour"]; exists {
			flattened.PriceApartmentsNightHour = uint32(price)
		}
		if price, exists := listing.PricingInfo.DurationPrices["apartments_night_2hour"]; exists {
			flattened.PriceApartmentsNight2Hour = uint32(price)
		}

		if price, exists := listing.PricingInfo.DurationPrices["outcall_day_hour"]; exists {
			flattened.PriceOutcallDayHour = uint32(price)
		}
		if price, exists := listing.PricingInfo.DurationPrices["outcall_day_2hour"]; exists {
			flattened.PriceOutcallDay2Hour = uint32(price)
		}
		if price, exists := listing.PricingInfo.DurationPrices["outcall_night_hour"]; exists {
			flattened.PriceOutcallNightHour = uint32(price)
		}
		if price, exists := listing.PricingInfo.DurationPrices["outcall_night_2hour"]; exists {
			flattened.PriceOutcallNight2Hour = uint32(price)
		}

		// Extract legacy pricing fields with priority: apartments -> outcall -> legacy keys
		if flattened.PriceApartmentsDayHour > 0 {
			flattened.PriceHour = flattened.PriceApartmentsDayHour
		} else if flattened.PriceOutcallDayHour > 0 {
			flattened.PriceHour = flattened.PriceOutcallDayHour
		} else if price, exists := listing.PricingInfo.DurationPrices["час"]; exists {
			flattened.PriceHour = uint32(price)
		} else if price, exists := listing.PricingInfo.DurationPrices["hour"]; exists {
			flattened.PriceHour = uint32(price)
		}

		if flattened.PriceApartmentsDay2Hour > 0 {
			flattened.Price2Hours = flattened.PriceApartmentsDay2Hour
		} else if flattened.PriceOutcallDay2Hour > 0 {
			flattened.Price2Hours = flattened.PriceOutcallDay2Hour
		} else if price, exists := listing.PricingInfo.DurationPrices["2 часа"]; exists {
			flattened.Price2Hours = uint32(price)
		} else if price, exists := listing.PricingInfo.DurationPrices["2 hours"]; exists {
			flattened.Price2Hours = uint32(price)
		}

		if flattened.PriceApartmentsNightHour > 0 {
			flattened.PriceNight = flattened.PriceApartmentsNightHour
		} else if flattened.PriceOutcallNightHour > 0 {
			flattened.PriceNight = flattened.PriceOutcallNightHour
		} else if price, exists := listing.PricingInfo.DurationPrices["ночь"]; exists {
			flattened.PriceNight = uint32(price)
		} else if price, exists := listing.PricingInfo.DurationPrices["night"]; exists {
			flattened.PriceNight = uint32(price)
		}

		// Day price: prefer 2-hour rates
		if flattened.PriceApartmentsDay2Hour > 0 {
			flattened.PriceDay = flattened.PriceApartmentsDay2Hour
		} else if flattened.PriceOutcallDay2Hour > 0 {
			flattened.PriceDay = flattened.PriceOutcallDay2Hour
		} else if price, exists := listing.PricingInfo.DurationPrices["день"]; exists {
			flattened.PriceDay = uint32(price)
		} else if price, exists := listing.PricingInfo.DurationPrices["day"]; exists {
			flattened.PriceDay = uint32(price)
		}

		// Base price
		if price, exists := listing.PricingInfo.DurationPrices["base"]; exists {
			flattened.PriceBase = uint32(price)
		} else {
			flattened.PriceBase = flattened.PriceHour
		}

		// Convert service prices map
		flattened.PricingServicePrices = make(map[string]uint32)
		for k, v := range listing.PricingInfo.ServicePrices {
			flattened.PricingServicePrices[k] = uint32(v)
		}
	}

	// Flatten service info
	if listing.ServiceInfo != nil {
		flattened.ServiceAvailable = listing.ServiceInfo.AvailableServices
		flattened.ServiceAdditional = listing.ServiceInfo.AdditionalServices
		flattened.ServiceRestrictions = listing.ServiceInfo.Restrictions
		flattened.ServiceMeetingType = listing.ServiceInfo.MeetingType
	}

	// Flatten location info
	if listing.LocationInfo != nil {
		flattened.LocationMetroStations = listing.LocationInfo.MetroStations
		flattened.LocationDistrict = listing.LocationInfo.District
		flattened.LocationCity = listing.LocationInfo.City
		flattened.LocationOutcallAvailable = listing.LocationInfo.OutcallAvailable
		flattened.LocationIncallAvailable = listing.LocationInfo.IncallAvailable
	}

	// Set default city if empty
	if flattened.LocationCity == "" {
		flattened.LocationCity = "Unknown"
	}

	return flattened
}

// InsertListing inserts a single listing into ClickHouse
func (a *Adapter) InsertListing(ctx context.Context, listing *listing.Listing, sourceURL string) error {
	flattened := a.FlattenListing(listing, sourceURL)
	return a.InsertFlattenedListing(ctx, flattened)
}

// InsertFlattenedListing inserts a flattened listing into ClickHouse
func (a *Adapter) InsertFlattenedListing(ctx context.Context, flattened *FlattenedListing) error {
	query := `
		INSERT INTO listings (
			id, created_at, updated_at, last_scraped, source_url,
			personal_name, personal_age, personal_height, personal_weight, personal_breast_size,
			personal_hair_color, personal_eye_color, personal_body_type,
			contact_phone, contact_telegram, contact_email,
			pricing_currency,
			price_apartments_day_hour, price_apartments_day_2hour, price_apartments_night_hour, price_apartments_night_2hour,
			price_outcall_day_hour, price_outcall_day_2hour, price_outcall_night_hour, price_outcall_night_2hour,
			price_hour, price_2_hours, price_night, price_day, price_base,
			pricing_duration_prices, pricing_service_prices,
			service_available, service_additional, service_restrictions, service_meeting_type,
			location_metro_stations, location_district, location_city, 
			location_outcall_available, location_incall_available,
			description, last_updated, photos, photos_count
		) VALUES (
			?, ?, ?, ?, ?,
			?, ?, ?, ?, ?,
			?, ?, ?,
			?, ?, ?,
			?,
			?, ?, ?, ?,
			?, ?, ?, ?,
			?, ?, ?, ?, ?,
			?, ?,
			?, ?, ?, ?,
			?, ?, ?,
			?, ?,
			?, ?, ?, ?
		)`

	err := a.conn.Exec(ctx, query,
		flattened.ID, flattened.CreatedAt, flattened.UpdatedAt, flattened.LastScraped, flattened.SourceURL,
		flattened.PersonalName, flattened.PersonalAge, flattened.PersonalHeight, flattened.PersonalWeight, flattened.PersonalBreastSize,
		flattened.PersonalHairColor, flattened.PersonalEyeColor, flattened.PersonalBodyType,
		flattened.ContactPhone, flattened.ContactTelegram, flattened.ContactEmail,
		flattened.PricingCurrency,
		flattened.PriceApartmentsDayHour, flattened.PriceApartmentsDay2Hour, flattened.PriceApartmentsNightHour, flattened.PriceApartmentsNight2Hour,
		flattened.PriceOutcallDayHour, flattened.PriceOutcallDay2Hour, flattened.PriceOutcallNightHour, flattened.PriceOutcallNight2Hour,
		flattened.PriceHour, flattened.Price2Hours, flattened.PriceNight, flattened.PriceDay, flattened.PriceBase,
		flattened.PricingDurationPrices, flattened.PricingServicePrices,
		flattened.ServiceAvailable, flattened.ServiceAdditional, flattened.ServiceRestrictions, flattened.ServiceMeetingType,
		flattened.LocationMetroStations, flattened.LocationDistrict, flattened.LocationCity,
		flattened.LocationOutcallAvailable, flattened.LocationIncallAvailable,
		flattened.Description, flattened.LastUpdated, flattened.Photos, flattened.PhotosCount,
	)

	if err != nil {
		return fmt.Errorf("failed to insert listing %s: %w", flattened.ID, err)
	}

	return nil
}

// BatchInsertListings inserts multiple listings in a batch
func (a *Adapter) BatchInsertListings(ctx context.Context, listings []*listing.Listing, sourceURLs []string) error {
	if len(listings) == 0 {
		return nil
	}

	if len(sourceURLs) != len(listings) {
		return fmt.Errorf("sourceURLs length (%d) must match listings length (%d)", len(sourceURLs), len(listings))
	}

	batch, err := a.conn.PrepareBatch(ctx, `
		INSERT INTO listings (
			id, created_at, updated_at, last_scraped, source_url,
			personal_name, personal_age, personal_height, personal_weight, personal_breast_size,
			personal_hair_color, personal_eye_color, personal_body_type,
			contact_phone, contact_telegram, contact_email,
			pricing_currency,
			price_apartments_day_hour, price_apartments_day_2hour, price_apartments_night_hour, price_apartments_night_2hour,
			price_outcall_day_hour, price_outcall_day_2hour, price_outcall_night_hour, price_outcall_night_2hour,
			price_hour, price_2_hours, price_night, price_day, price_base,
			pricing_duration_prices, pricing_service_prices,
			service_available, service_additional, service_restrictions, service_meeting_type,
			location_metro_stations, location_district, location_city, 
			location_outcall_available, location_incall_available,
			description, last_updated, photos, photos_count
		)
	`)

	if err != nil {
		return fmt.Errorf("failed to prepare batch: %w", err)
	}

	for i, listing := range listings {
		flattened := a.FlattenListing(listing, sourceURLs[i])

		err := batch.Append(
			flattened.ID, flattened.CreatedAt, flattened.UpdatedAt, flattened.LastScraped, flattened.SourceURL,
			flattened.PersonalName, flattened.PersonalAge, flattened.PersonalHeight, flattened.PersonalWeight, flattened.PersonalBreastSize,
			flattened.PersonalHairColor, flattened.PersonalEyeColor, flattened.PersonalBodyType,
			flattened.ContactPhone, flattened.ContactTelegram, flattened.ContactEmail,
			flattened.PricingCurrency,
			flattened.PriceApartmentsDayHour, flattened.PriceApartmentsDay2Hour, flattened.PriceApartmentsNightHour, flattened.PriceApartmentsNight2Hour,
			flattened.PriceOutcallDayHour, flattened.PriceOutcallDay2Hour, flattened.PriceOutcallNightHour, flattened.PriceOutcallNight2Hour,
			flattened.PriceHour, flattened.Price2Hours, flattened.PriceNight, flattened.PriceDay, flattened.PriceBase,
			flattened.PricingDurationPrices, flattened.PricingServicePrices,
			flattened.ServiceAvailable, flattened.ServiceAdditional, flattened.ServiceRestrictions, flattened.ServiceMeetingType,
			flattened.LocationMetroStations, flattened.LocationDistrict, flattened.LocationCity,
			flattened.LocationOutcallAvailable, flattened.LocationIncallAvailable,
			flattened.Description, flattened.LastUpdated, flattened.Photos, flattened.PhotosCount,
		)

		if err != nil {
			return fmt.Errorf("failed to append listing %s to batch: %w", flattened.ID, err)
		}
	}

	err = batch.Send()
	if err != nil {
		return fmt.Errorf("failed to send batch: %w", err)
	}

	return nil
}

// UpdateListing updates an existing listing or inserts if not exists
func (a *Adapter) UpdateListing(ctx context.Context, listing *listing.Listing, sourceURL string) error {
	flattened := a.FlattenListing(listing, sourceURL)
	flattened.UpdatedAt = time.Now()

	// ClickHouse ReplacingMergeTree will automatically handle updates based on the sorting key
	return a.InsertFlattenedListing(ctx, flattened)
}

// GetListingByID retrieves a listing by ID
func (a *Adapter) GetListingByID(ctx context.Context, id string) (*FlattenedListing, error) {
	query := `
		SELECT 
			id, created_at, updated_at, last_scraped, source_url,
			personal_name, personal_age, personal_height, personal_weight, personal_breast_size,
			personal_hair_color, personal_eye_color, personal_body_type,
			contact_phone, contact_telegram, contact_email,
			pricing_currency,
			price_apartments_day_hour, price_apartments_day_2hour, price_apartments_night_hour, price_apartments_night_2hour,
			price_outcall_day_hour, price_outcall_day_2hour, price_outcall_night_hour, price_outcall_night_2hour,
			price_hour, price_2_hours, price_night, price_day, price_base,
			pricing_duration_prices, pricing_service_prices,
			service_available, service_additional, service_restrictions, service_meeting_type,
			location_metro_stations, location_district, location_city,
			location_outcall_available, location_incall_available,
			description, last_updated, photos, photos_count
		FROM listings 
		WHERE id = ? 
		ORDER BY updated_at DESC 
		LIMIT 1
	`

	row := a.conn.QueryRow(ctx, query, id)

	var flattened FlattenedListing
	err := row.Scan(
		&flattened.ID, &flattened.CreatedAt, &flattened.UpdatedAt, &flattened.LastScraped, &flattened.SourceURL,
		&flattened.PersonalName, &flattened.PersonalAge, &flattened.PersonalHeight, &flattened.PersonalWeight, &flattened.PersonalBreastSize,
		&flattened.PersonalHairColor, &flattened.PersonalEyeColor, &flattened.PersonalBodyType,
		&flattened.ContactPhone, &flattened.ContactTelegram, &flattened.ContactEmail,
		&flattened.PricingCurrency,
		&flattened.PriceApartmentsDayHour, &flattened.PriceApartmentsDay2Hour, &flattened.PriceApartmentsNightHour, &flattened.PriceApartmentsNight2Hour,
		&flattened.PriceOutcallDayHour, &flattened.PriceOutcallDay2Hour, &flattened.PriceOutcallNightHour, &flattened.PriceOutcallNight2Hour,
		&flattened.PriceHour, &flattened.Price2Hours, &flattened.PriceNight, &flattened.PriceDay, &flattened.PriceBase,
		&flattened.PricingDurationPrices, &flattened.PricingServicePrices,
		&flattened.ServiceAvailable, &flattened.ServiceAdditional, &flattened.ServiceRestrictions, &flattened.ServiceMeetingType,
		&flattened.LocationMetroStations, &flattened.LocationDistrict, &flattened.LocationCity,
		&flattened.LocationOutcallAvailable, &flattened.LocationIncallAvailable,
		&flattened.Description, &flattened.LastUpdated, &flattened.Photos, &flattened.PhotosCount,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("listing with ID %s not found", id)
		}
		return nil, fmt.Errorf("failed to get listing %s: %w", id, err)
	}

	return &flattened, nil
}

// GetStats returns basic statistics about listings in the database
func (a *Adapter) GetStats(ctx context.Context) (map[string]interface{}, error) {
	query := `
		SELECT 
			count() as total_listings,
			countIf(personal_age > 0) as listings_with_age,
			countIf(price_hour > 0) as listings_with_price,
			countIf(length(contact_phone) > 0) as listings_with_phone,
			countIf(length(photos) > 0) as listings_with_photos,
			avg(personal_age) as avg_age,
			avg(price_hour) as avg_price_hour,
			uniqExact(location_city) as unique_cities
		FROM listings
		FINAL
	`

	row := a.conn.QueryRow(ctx, query)

	var stats struct {
		TotalListings      uint64
		ListingsWithAge    uint64
		ListingsWithPrice  uint64
		ListingsWithPhone  uint64
		ListingsWithPhotos uint64
		AvgAge             float64
		AvgPriceHour       float64
		UniqueCities       uint64
	}

	err := row.Scan(
		&stats.TotalListings,
		&stats.ListingsWithAge,
		&stats.ListingsWithPrice,
		&stats.ListingsWithPhone,
		&stats.ListingsWithPhotos,
		&stats.AvgAge,
		&stats.AvgPriceHour,
		&stats.UniqueCities,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}

	result := map[string]interface{}{
		"total_listings":       stats.TotalListings,
		"listings_with_age":    stats.ListingsWithAge,
		"listings_with_price":  stats.ListingsWithPrice,
		"listings_with_phone":  stats.ListingsWithPhone,
		"listings_with_photos": stats.ListingsWithPhotos,
		"avg_age":              stats.AvgAge,
		"avg_price_hour":       stats.AvgPriceHour,
		"unique_cities":        stats.UniqueCities,
	}

	return result, nil
}

// LogChange logs a change to the listing_changes table
func (a *Adapter) LogChange(ctx context.Context, listingID, changeType, oldValue, newValue, fieldName, source string) error {
	query := `
		INSERT INTO listing_changes (listing_id, change_type, old_value, new_value, field_name, source)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	err := a.conn.Exec(ctx, query, listingID, changeType, oldValue, newValue, fieldName, source)
	if err != nil {
		return fmt.Errorf("failed to log change for listing %s: %w", listingID, err)
	}

	return nil
}
