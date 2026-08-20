package requests

type PlaceOrderRequest struct {
	Symbol           string
	Side             string
	Type             string
	Quantity         string
	QuoteOrderQty    string
	NewClientOrderID string
}

type QueryOrderRequest struct {
	Symbol            string
	OrderID           string
	OrigClientOrderID string
}
