package types

import (
	"encoding/json"
	"errors"
	"fmt"
)

// Server message types carried in WSMessage.Type. The list_subscriptions
// reply reuses "ok".
const (
	WSTypeSubscribed                    = "subscribed"
	WSTypeUnsubscribed                  = "unsubscribed"
	WSTypeOK                            = "ok"
	WSTypeError                         = "error"
	WSTypeOrderbookSnapshot             = "orderbook_snapshot"
	WSTypeOrderbookDelta                = "orderbook_delta"
	WSTypeTicker                        = "ticker"
	WSTypeTrade                         = "trade"
	WSTypeFill                          = "fill"
	WSTypeMarketPosition                = "market_position"
	WSTypeMarketLifecycleV2             = "market_lifecycle_v2"
	WSTypeMultivariateMarketLifecycle   = "multivariate_market_lifecycle"
	WSTypeEventLifecycle                = "event_lifecycle"
	WSTypeEventFeeUpdate                = "event_fee_update"
	WSTypeOrderGroupUpdates             = "order_group_updates"
	WSTypeUserOrder                     = "user_order"
	WSTypeRFQCreated                    = "rfq_created"
	WSTypeRFQDeleted                    = "rfq_deleted"
	WSTypeQuoteCreated                  = "quote_created"
	WSTypeQuoteAccepted                 = "quote_accepted"
	WSTypeQuoteExecuted                 = "quote_executed"
	WSTypePythValue                     = "pyth_value"
	WSTypePythValueUnderlyingList       = "pyth_value_underlying_list"
	WSTypeCFBenchmarksValue             = "cfbenchmarks_value"
	WSTypeCFBenchmarksValueIndexList    = "cfbenchmarks_value_indexlist"
	WSTypeCFBenchmarksValue5Hz          = "cfbenchmarks_value_5hz"
	WSTypeCFBenchmarksValue5HzIndexList = "cfbenchmarks_value_5hz_indexlist"
)

// Decode unmarshals the msg payload into v, which should be a pointer to the
// *Msg struct matching m.Type.
func (m *WSMessage) Decode(v any) error {
	if len(m.Msg) == 0 {
		return fmt.Errorf("ws message %q has no msg payload", m.Type)
	}
	return json.Unmarshal(m.Msg, v)
}

// OrderbookLevel is one aggregated price level, sent on the wire as a
// [price_dollars, count_fp] string pair.
type OrderbookLevel struct {
	PriceDollars string
	CountFp      string
}

func (l OrderbookLevel) MarshalJSON() ([]byte, error) {
	return json.Marshal([2]string{l.PriceDollars, l.CountFp})
}

func (l *OrderbookLevel) UnmarshalJSON(b []byte) error {
	var pair []string
	if err := json.Unmarshal(b, &pair); err != nil {
		return fmt.Errorf("orderbook level: %w", err)
	}
	if len(pair) != 2 {
		return fmt.Errorf("orderbook level: want [price, count], got %d elements", len(pair))
	}
	if pair[0] == "" || pair[1] == "" {
		return errors.New("orderbook level: empty price or count")
	}
	l.PriceDollars, l.CountFp = pair[0], pair[1]
	return nil
}

type OrderbookSnapshotMsg struct {
	MarketTicker string           `json:"market_ticker"`
	MarketID     string           `json:"market_id"`
	YesDollarsFp []OrderbookLevel `json:"yes_dollars_fp,omitempty"`
	NoDollarsFp  []OrderbookLevel `json:"no_dollars_fp,omitempty"`
}

type OrderbookDeltaMsg struct {
	MarketTicker  string `json:"market_ticker"`
	MarketID      string `json:"market_id"`
	PriceDollars  string `json:"price_dollars"`
	DeltaFp       string `json:"delta_fp"`
	Side          string `json:"side"`
	ClientOrderID string `json:"client_order_id,omitempty"`
	Subaccount    *int   `json:"subaccount,omitempty"`
	Ts            string `json:"ts,omitempty"`
	TsMs          *int64 `json:"ts_ms,omitempty"`
}

type TickerMsg struct {
	MarketTicker       string `json:"market_ticker"`
	MarketID           string `json:"market_id"`
	PriceDollars       string `json:"price_dollars"`
	YesBidDollars      string `json:"yes_bid_dollars"`
	YesAskDollars      string `json:"yes_ask_dollars"`
	VolumeFp           string `json:"volume_fp"`
	OpenInterestFp     string `json:"open_interest_fp"`
	DollarVolume       int64  `json:"dollar_volume"`
	DollarOpenInterest int64  `json:"dollar_open_interest"`
	YesBidSizeFp       string `json:"yes_bid_size_fp"`
	YesAskSizeFp       string `json:"yes_ask_size_fp"`
	LastTradeSizeFp    string `json:"last_trade_size_fp"`
	Ts                 int64  `json:"ts"`
	TsMs               int64  `json:"ts_ms"`
	Time               string `json:"time"`
}

type TradeMsg struct {
	TradeID          string `json:"trade_id"`
	MarketTicker     string `json:"market_ticker"`
	YesPriceDollars  string `json:"yes_price_dollars"`
	NoPriceDollars   string `json:"no_price_dollars"`
	CountFp          string `json:"count_fp"`
	TakerSide        string `json:"taker_side,omitempty"`
	TakerOutcomeSide string `json:"taker_outcome_side"`
	TakerBookSide    string `json:"taker_book_side"`
	IsBlockTrade     bool   `json:"is_block_trade"`
	Ts               int64  `json:"ts"`
	TsMs             int64  `json:"ts_ms"`
}

type FillMsg struct {
	TradeID         string `json:"trade_id"`
	OrderID         string `json:"order_id"`
	MarketTicker    string `json:"market_ticker"`
	ExchangeIndex   int    `json:"exchange_index"`
	IsTaker         bool   `json:"is_taker"`
	Side            string `json:"side,omitempty"`
	YesPriceDollars string `json:"yes_price_dollars"`
	CountFp         string `json:"count_fp"`
	FeeCost         string `json:"fee_cost"`
	Action          string `json:"action,omitempty"`
	Ts              int64  `json:"ts"`
	TsMs            int64  `json:"ts_ms"`
	ClientOrderID   string `json:"client_order_id,omitempty"`
	PostPositionFp  string `json:"post_position_fp"`
	PurchasedSide   string `json:"purchased_side,omitempty"`
	OutcomeSide     string `json:"outcome_side"`
	BookSide        string `json:"book_side"`
	Subaccount      *int   `json:"subaccount,omitempty"`
}

type MarketPositionMsg struct {
	UserID                 string `json:"user_id"`
	MarketTicker           string `json:"market_ticker"`
	PositionFp             string `json:"position_fp"`
	PositionCostDollars    string `json:"position_cost_dollars"`
	RealizedPnlDollars     string `json:"realized_pnl_dollars"`
	FeesPaidDollars        string `json:"fees_paid_dollars"`
	PositionFeeCostDollars string `json:"position_fee_cost_dollars"`
	VolumeFp               string `json:"volume_fp"`
	Subaccount             *int   `json:"subaccount,omitempty"`
}

type UserOrderMsg struct {
	OrderID                 string `json:"order_id"`
	UserID                  string `json:"user_id"`
	Ticker                  string `json:"ticker"`
	ExchangeIndex           int    `json:"exchange_index"`
	Status                  string `json:"status"`
	Side                    string `json:"side,omitempty"`
	IsYes                   bool   `json:"is_yes"`
	OutcomeSide             string `json:"outcome_side"`
	BookSide                string `json:"book_side"`
	YesPriceDollars         string `json:"yes_price_dollars"`
	FillCountFp             string `json:"fill_count_fp"`
	RemainingCountFp        string `json:"remaining_count_fp"`
	InitialCountFp          string `json:"initial_count_fp"`
	TakerFillCostDollars    string `json:"taker_fill_cost_dollars"`
	MakerFillCostDollars    string `json:"maker_fill_cost_dollars"`
	TakerFeesDollars        string `json:"taker_fees_dollars"`
	MakerFeesDollars        string `json:"maker_fees_dollars"`
	ClientOrderID           string `json:"client_order_id"`
	OrderGroupID            string `json:"order_group_id,omitempty"`
	SelfTradePreventionType string `json:"self_trade_prevention_type,omitempty"`
	CreatedTime             string `json:"created_time"`
	CreatedTsMs             int64  `json:"created_ts_ms"`
	LastUpdateTime          string `json:"last_update_time,omitempty"`
	LastUpdatedTsMs         *int64 `json:"last_updated_ts_ms,omitempty"`
	ExpirationTime          string `json:"expiration_time,omitempty"`
	ExpirationTsMs          *int64 `json:"expiration_ts_ms,omitempty"`
	SubaccountNumber        *int   `json:"subaccount_number,omitempty"`
}

type OrderGroupUpdatesMsg struct {
	EventType        string `json:"event_type"`
	OrderGroupID     string `json:"order_group_id"`
	ContractsLimitFp string `json:"contracts_limit_fp,omitempty"`
	TsMs             int64  `json:"ts_ms"`
}

type MultivariateMarketLifecycleMsg struct {
	EventType           string                             `json:"event_type"`
	MarketTicker        string                             `json:"market_ticker"`
	ExchangeIndex       *int                               `json:"exchange_index,omitempty"`
	OpenTs              *int64                             `json:"open_ts,omitempty"`
	CloseTs             *int64                             `json:"close_ts,omitempty"`
	Result              string                             `json:"result,omitempty"`
	DeterminationTs     *int64                             `json:"determination_ts,omitempty"`
	SettlementValue     string                             `json:"settlement_value,omitempty"`
	SettledTs           *int64                             `json:"settled_ts,omitempty"`
	IsDeactivated       *bool                              `json:"is_deactivated,omitempty"`
	PriceLevelStructure string                             `json:"price_level_structure,omitempty"`
	AdditionalMetadata  *MarketLifecycleAdditionalMetadata `json:"additional_metadata,omitempty"`
}

type EventLifecycleMsg struct {
	EventTicker          string `json:"event_ticker"`
	ExchangeIndex        int    `json:"exchange_index"`
	Title                string `json:"title"`
	Subtitle             string `json:"subtitle"`
	CollateralReturnType string `json:"collateral_return_type"`
	SeriesTicker         string `json:"series_ticker"`
	StrikeDate           *int64 `json:"strike_date,omitempty"`
	StrikePeriod         string `json:"strike_period,omitempty"`
}

// EventFeeUpdateMsg carries an event-level fee override; both override fields
// are null when the override has been cleared.
type EventFeeUpdateMsg struct {
	EventTicker           string   `json:"event_ticker"`
	FeeTypeOverride       *string  `json:"fee_type_override"`
	FeeMultiplierOverride *float64 `json:"fee_multiplier_override"`
}

type RFQCreatedMsg struct {
	ID                  string           `json:"id"`
	CreatorID           string           `json:"creator_id"`
	MarketTicker        string           `json:"market_ticker"`
	EventTicker         string           `json:"event_ticker,omitempty"`
	ContractsFp         string           `json:"contracts_fp,omitempty"`
	TargetCostDollars   string           `json:"target_cost_dollars,omitempty"`
	CreatedTs           string           `json:"created_ts"`
	MveCollectionTicker string           `json:"mve_collection_ticker,omitempty"`
	MveSelectedLegs     []MveSelectedLeg `json:"mve_selected_legs,omitempty"`
}

type RFQDeletedMsg struct {
	ID                string `json:"id"`
	CreatorID         string `json:"creator_id"`
	MarketTicker      string `json:"market_ticker"`
	EventTicker       string `json:"event_ticker,omitempty"`
	ContractsFp       string `json:"contracts_fp,omitempty"`
	TargetCostDollars string `json:"target_cost_dollars,omitempty"`
	DeletedTs         string `json:"deleted_ts"`
}

type QuoteCreatedMsg struct {
	QuoteID               string `json:"quote_id"`
	RFQID                 string `json:"rfq_id"`
	QuoteCreatorID        string `json:"quote_creator_id"`
	RFQCreatorID          string `json:"rfq_creator_id,omitempty"`
	MarketTicker          string `json:"market_ticker"`
	EventTicker           string `json:"event_ticker,omitempty"`
	YesBidDollars         string `json:"yes_bid_dollars"`
	NoBidDollars          string `json:"no_bid_dollars"`
	YesContractsOfferedFp string `json:"yes_contracts_offered_fp,omitempty"`
	NoContractsOfferedFp  string `json:"no_contracts_offered_fp,omitempty"`
	RFQTargetCostDollars  string `json:"rfq_target_cost_dollars,omitempty"`
	CreatedTs             string `json:"created_ts"`
	Subaccount            *int   `json:"subaccount,omitempty"`
}

type QuoteAcceptedMsg struct {
	QuoteID               string `json:"quote_id"`
	RFQID                 string `json:"rfq_id"`
	QuoteCreatorID        string `json:"quote_creator_id"`
	RFQCreatorID          string `json:"rfq_creator_id,omitempty"`
	MarketTicker          string `json:"market_ticker"`
	EventTicker           string `json:"event_ticker,omitempty"`
	YesBidDollars         string `json:"yes_bid_dollars"`
	NoBidDollars          string `json:"no_bid_dollars"`
	AcceptedSide          string `json:"accepted_side,omitempty"`
	ContractsAcceptedFp   string `json:"contracts_accepted_fp,omitempty"`
	YesContractsOfferedFp string `json:"yes_contracts_offered_fp,omitempty"`
	NoContractsOfferedFp  string `json:"no_contracts_offered_fp,omitempty"`
	RFQTargetCostDollars  string `json:"rfq_target_cost_dollars,omitempty"`
	Subaccount            *int   `json:"subaccount,omitempty"`
}

type QuoteExecutedMsg struct {
	QuoteID        string `json:"quote_id"`
	RFQID          string `json:"rfq_id"`
	QuoteCreatorID string `json:"quote_creator_id"`
	RFQCreatorID   string `json:"rfq_creator_id"`
	OrderID        string `json:"order_id"`
	ClientOrderID  string `json:"client_order_id"`
	MarketTicker   string `json:"market_ticker"`
	ExecutedTs     string `json:"executed_ts"`
	Subaccount     *int   `json:"subaccount,omitempty"`
}
