//nolint:tagliatelle
package requests

type (
	AccountAssetsRequest struct {
		Coin      string `json:"coin" url:"coin,omitempty"`
		AssetType string `json:"assetType" url:"assetType,omitempty"`
	}

	DepositAddressRequest struct {
		Coin  string `json:"coin" url:"coin"`
		Chain string `json:"chain" url:"chain,omitempty"`
		Size  string `json:"size" url:"size,omitempty"`
	}

	DepositRecordsRequest struct {
		Coin       string `json:"coin" url:"coin,omitempty"`
		OrderID    string `json:"orderId" url:"orderId,omitempty"`
		StartTime  string `json:"startTime" url:"startTime"`
		EndTime    string `json:"endTime" url:"endTime"`
		IDLestThan string `json:"idLestThan" url:"idLestThan,omitempty"`
		Limit      string `json:"limit" url:"limit,omitempty"`
	}

	WithdrawalRecordsRequest struct {
		Coin       string `json:"coin,omitempty" url:"coin,omitempty"`
		ClientOid  string `json:"clientOid,omitempty" url:"clientOid,omitempty"`
		StartTime  string `json:"startTime" url:"startTime"`
		EndTime    string `json:"endTime" url:"endTime"`
		IDLestThan string `json:"idLestThan,omitempty" url:"idLestThan,omitempty"`
		OrderID    string `json:"orderId,omitempty" url:"orderId,omitempty"`
		Limit      string `json:"limit,omitempty" url:"limit,omitempty"`
	}

	WalletWithdrawalRequest struct {
		Coin         string `json:"coin" url:"coin"`
		TransferType string `json:"transferType" url:"transferType"`
		Address      string `json:"address" url:"address"`
		Chain        string `json:"chain" url:"chain"`
		Size         string `json:"size" url:"size"`
		ClientOID    string `json:"clientOid" url:"clientOid,omitempty"`
	}
)
