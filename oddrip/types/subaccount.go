package types

type CreateSubaccountRequest struct {
	ExchangeIndex *int `json:"exchange_index,omitempty"`
}

type CreateSubaccountResponse struct {
	SubaccountNumber int `json:"subaccount_number"`
}

// ApplySubaccountTransferRequest moves AmountCents between subaccounts; 0 is
// the primary account, 1-63 are numbered subaccounts.
type ApplySubaccountTransferRequest struct {
	ClientTransferID string `json:"client_transfer_id"`
	FromSubaccount   int    `json:"from_subaccount"`
	ToSubaccount     int    `json:"to_subaccount"`
	AmountCents      int64  `json:"amount_cents"`
	ExchangeIndex    *int   `json:"exchange_index,omitempty"`
}

type SubaccountBalance struct {
	SubaccountNumber int    `json:"subaccount_number"`
	ExchangeIndex    int    `json:"exchange_index"`
	Balance          string `json:"balance"`
	UpdatedTs        int64  `json:"updated_ts"`
}

type GetSubaccountBalancesResponse struct {
	SubaccountBalances []SubaccountBalance `json:"subaccount_balances"`
}

type SubaccountTransfer struct {
	TransferID     string `json:"transfer_id"`
	FromSubaccount int    `json:"from_subaccount"`
	ToSubaccount   int    `json:"to_subaccount"`
	AmountCents    int64  `json:"amount_cents"`
	CreatedTs      int64  `json:"created_ts"`
	ExchangeIndex  int    `json:"exchange_index"`
}

type GetSubaccountTransfersResponse struct {
	Transfers []SubaccountTransfer `json:"transfers"`
	Cursor    string               `json:"cursor,omitempty"`
}

type GetSubaccountTransfersOpts struct {
	Limit  *int64
	Cursor string
}

type UpdateSubaccountNettingRequest struct {
	SubaccountNumber int  `json:"subaccount_number"`
	Enabled          bool `json:"enabled"`
}

type SubaccountNettingConfig struct {
	SubaccountNumber int  `json:"subaccount_number"`
	Enabled          bool `json:"enabled"`
	ExchangeIndex    int  `json:"exchange_index"`
}

type GetSubaccountNettingResponse struct {
	NettingConfigs []SubaccountNettingConfig `json:"netting_configs"`
}
