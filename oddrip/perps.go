package oddrip

// PerpsClient groups the perpetual-futures API, reached as Client.Perps.
// Kalshi calls the product "perps"; the API itself says "margin" (paths under
// /margin, Margin-prefixed schemas), and so do the types. Perps REST is served
// from the same host as event contracts, so these services share the Client's
// base URL, auth, and retry policy. Kalshi's perps spec lists only the
// external-api hosts; for demo use
// BaseURL("https://external-api.demo.kalshi.co/trade-api/v2"), not the shared
// demo-api.kalshi.co. The perps WebSocket is on its own host; see
// ConnectPerpsWS.
type PerpsClient struct {
	Exchange     *PerpsExchangeService
	Account      *PerpsAccountService
	Markets      *PerpsMarketsService
	Orders       *PerpsOrdersService
	OrderGroups  *PerpsOrderGroupsService
	Portfolio    *PerpsPortfolioService
	Risk         *PerpsRiskService
	Fees         *PerpsFeesService
	Funding      *PerpsFundingService
	ExitTriggers *PerpsExitTriggersService
	FCM          *PerpsFCMService
}

type PerpsExchangeService struct {
	client *Client
}

type PerpsAccountService struct {
	client *Client
}

type PerpsMarketsService struct {
	client *Client
}

type PerpsOrdersService struct {
	client *Client
}

type PerpsOrderGroupsService struct {
	client *Client
}

type PerpsPortfolioService struct {
	client *Client
}

type PerpsRiskService struct {
	client *Client
}

type PerpsFeesService struct {
	client *Client
}

type PerpsFundingService struct {
	client *Client
}

type PerpsExitTriggersService struct {
	client *Client
}

type PerpsFCMService struct {
	client *Client
}

func newPerpsClient(c *Client) *PerpsClient {
	return &PerpsClient{
		Exchange:     &PerpsExchangeService{client: c},
		Account:      &PerpsAccountService{client: c},
		Markets:      &PerpsMarketsService{client: c},
		Orders:       &PerpsOrdersService{client: c},
		OrderGroups:  &PerpsOrderGroupsService{client: c},
		Portfolio:    &PerpsPortfolioService{client: c},
		Risk:         &PerpsRiskService{client: c},
		Fees:         &PerpsFeesService{client: c},
		Funding:      &PerpsFundingService{client: c},
		ExitTriggers: &PerpsExitTriggersService{client: c},
		FCM:          &PerpsFCMService{client: c},
	}
}
