package types

type EventData struct {
	EventTicker            string                 `json:"event_ticker"`
	SeriesTicker           string                 `json:"series_ticker"`
	SubTitle               string                 `json:"sub_title"`
	Title                  string                 `json:"title"`
	CollateralReturnType   string                 `json:"collateral_return_type"`
	MutuallyExclusive      bool                   `json:"mutually_exclusive"`
	Category               string                 `json:"category"`
	StrikeDate             *string                `json:"strike_date,omitempty"`
	StrikePeriod           *string                `json:"strike_period,omitempty"`
	Markets                []Market               `json:"markets,omitempty"`
	ProductMetadata        map[string]interface{} `json:"product_metadata,omitempty"`
	SettlementSources      []SettlementSource     `json:"settlement_sources,omitempty"`
	LastUpdatedTs          string                 `json:"last_updated_ts,omitempty"`
	FeeTypeOverride        string                 `json:"fee_type_override,omitempty"`
	FeeMultiplierOverride  *float64               `json:"fee_multiplier_override,omitempty"`
	ExchangeIndex          int                    `json:"exchange_index,omitempty"`
}

type GetEventsOpts struct {
	Limit             *int64
	Cursor            string
	WithNestedMarkets *bool
	WithMilestones    *bool
	Status            string
	SeriesTicker      string
	Tickers           string
	MinCloseTs        *int64
	MinUpdatedTs      *int64
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

const (
	CollectionStatusUnopened = "unopened"
	CollectionStatusOpen     = "open"
	CollectionStatusClosed   = "closed"
)

type AssociatedEvent struct {
	Ticker        string   `json:"ticker"`
	IsYesOnly     bool     `json:"is_yes_only"`
	SizeMax       *int     `json:"size_max,omitempty"`
	SizeMin       *int     `json:"size_min,omitempty"`
	ActiveQuoters []string `json:"active_quoters"`
}

type MultivariateEventCollection struct {
	CollectionTicker       string            `json:"collection_ticker"`
	SeriesTicker           string            `json:"series_ticker"`
	ExchangeIndex          int               `json:"exchange_index,omitempty"`
	Title                  string            `json:"title"`
	Description            string            `json:"description"`
	OpenDate               string            `json:"open_date"`
	CloseDate              string            `json:"close_date"`
	AssociatedEvents       []AssociatedEvent `json:"associated_events"`
	AssociatedEventTickers []string          `json:"associated_event_tickers"`
	IsOrdered              bool              `json:"is_ordered"`
	IsSingleMarketPerEvent bool              `json:"is_single_market_per_event"`
	IsAllYes               bool              `json:"is_all_yes"`
	SizeMin                int               `json:"size_min"`
	SizeMax                int               `json:"size_max"`
	FunctionalDescription  string            `json:"functional_description"`
}

type GetMultivariateEventCollectionResponse struct {
	MultivariateContract MultivariateEventCollection `json:"multivariate_contract"`
}

type GetMultivariateEventCollectionsResponse struct {
	MultivariateContracts []MultivariateEventCollection `json:"multivariate_contracts"`
	Cursor                string                        `json:"cursor,omitempty"`
}

type GetMultivariateEventCollectionsOpts struct {
	Status                string
	AssociatedEventTicker string
	SeriesTicker          string
	Limit                 *int64
	Cursor                string
}

type TickerPair struct {
	MarketTicker string `json:"market_ticker"`
	EventTicker  string `json:"event_ticker"`
	Side         string `json:"side"`
}

type CreateMarketInMultivariateEventCollectionRequest struct {
	SelectedMarkets   []TickerPair `json:"selected_markets"`
	WithMarketPayload *bool        `json:"with_market_payload,omitempty"`
}

type CreateMarketInMultivariateEventCollectionResponse struct {
	EventTicker  string  `json:"event_ticker"`
	MarketTicker string  `json:"market_ticker"`
	Market       *Market `json:"market,omitempty"`
}
