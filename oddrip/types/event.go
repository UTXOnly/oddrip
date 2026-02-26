package types

type EventData struct {
	EventTicker          string     `json:"event_ticker"`
	SeriesTicker         string     `json:"series_ticker"`
	SubTitle             string     `json:"sub_title"`
	Title                string     `json:"title"`
	CollateralReturnType string     `json:"collateral_return_type"`
	MutuallyExclusive    bool       `json:"mutually_exclusive"`
	Category             string     `json:"category"`
	StrikeDate           *string    `json:"strike_date,omitempty"`
	StrikePeriod         *string    `json:"strike_period,omitempty"`
	Markets              []Market   `json:"markets,omitempty"`
	AvailableOnBrokers   bool                    `json:"available_on_brokers"`
	ProductMetadata      map[string]interface{}  `json:"product_metadata,omitempty"`
	LastUpdatedTs        string                 `json:"last_updated_ts,omitempty"`
}

type GetEventsOpts struct {
	Limit             *int64
	Cursor            string
	WithNestedMarkets  *bool
	WithMilestones     *bool
	Status             string
	SeriesTicker       string
	MinCloseTs         *int64
	MinUpdatedTs       *int64
}

type GetEventsResponse struct {
	Events    []EventData `json:"events"`
	Milestones []Milestone `json:"milestones,omitempty"`
	Cursor    string      `json:"cursor"`
}

type Milestone struct {
	ID                    string   `json:"id"`
	Category              string   `json:"category"`
	Type                  string   `json:"type"`
	StartDate             string   `json:"start_date"`
	EndDate               *string  `json:"end_date,omitempty"`
	RelatedEventTickers   []string `json:"related_event_tickers"`
	Title                 string   `json:"title"`
	NotificationMessage   string   `json:"notification_message"`
	SourceID              *string  `json:"source_id,omitempty"`
	Details               map[string]interface{} `json:"details"`
	PrimaryEventTickers  []string `json:"primary_event_tickers"`
	LastUpdatedTs        string   `json:"last_updated_ts"`
}

type GetMultivariateEventsOpts struct {
	Limit            *int64
	Cursor           string
	SeriesTicker     string
	CollectionTicker string
	WithNestedMarkets *bool
}

type GetMultivariateEventsResponse struct {
	Events []EventData `json:"events"`
	Cursor string      `json:"cursor"`
}

type GetEventResponse struct {
	Event   EventData `json:"event"`
	Markets []Market  `json:"markets"`
}

type GetEventOpts struct {
	WithNestedMarkets *bool
}

type MarketMetadata struct {
	MarketTicker string `json:"market_ticker"`
	ImageURL     string `json:"image_url"`
	ColorCode    string `json:"color_code"`
}

type SettlementSource struct {
	Name string `json:"name,omitempty"`
	URL  string `json:"url,omitempty"`
}

type GetEventMetadataResponse struct {
	ImageURL         string             `json:"image_url"`
	FeaturedImageURL string             `json:"featured_image_url,omitempty"`
	MarketDetails    []MarketMetadata   `json:"market_details"`
	SettlementSources []SettlementSource `json:"settlement_sources"`
	Competition      *string            `json:"competition,omitempty"`
	CompetitionScope *string            `json:"competition_scope,omitempty"`
}
