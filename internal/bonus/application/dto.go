package application

// BalanceDTO is bonus balance response.
type BalanceDTO struct {
	Balance                   int64 `json:"balance"`
	ActiveDiscountPercent     int   `json:"activeDiscountPercent"`
	RemainingDiscountHeadroom int   `json:"remainingDiscountHeadroom"`
}

// TransactionDTO is ledger line for API.
type TransactionDTO struct {
	ID        string  `json:"id"`
	Amount    int64   `json:"amount"`
	Type      string  `json:"type"`
	Reference *string `json:"reference,omitempty"`
	CreatedAt string  `json:"createdAt"`
}

// ConvertResultDTO is convert response.
type ConvertResultDTO struct {
	TransactionID           string `json:"transactionId"`
	PointsConverted         int64  `json:"pointsConverted"`
	DiscountPercentAdded    int    `json:"discountPercentAdded"`
	ActiveDiscountPercent   int    `json:"activeDiscountPercent"`
	BalanceAfter            int64  `json:"balanceAfter"`
}
