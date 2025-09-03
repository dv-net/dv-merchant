//nolint:tagliatelle
package requests

type (
	GetCurrencies struct {
		Ccy []string `json:"ccy,omitempty"`
	}
	GetFundingBalance struct {
		Ccy []string `json:"ccy,omitempty"`
	}
	FundsTransfer struct {
		Ccy      string `json:"ccy"`
		Amt      string `json:"amt"`
		SubAcct  string `json:"subAcct,omitempty"`
		InstID   string `json:"instID,omitempty"`
		ToInstID string `json:"instId,omitempty"`
		Type     int    `json:"type,omitempty,string"`
		From     int    `json:"from,string"`
		To       int    `json:"to,string"`
	}
	AssetBillsDetails struct {
		Type   string `json:"type,string,omitempty"`
		After  int64  `json:"after,string,omitempty"`
		Before int64  `json:"before,string,omitempty"`
		Limit  int64  `json:"limit,string,omitempty"`
	}
	GetDepositAddress struct {
		Ccy string `json:"ccy"`
	}
	GetDepositHistory struct {
		Ccy    string `json:"ccy,omitempty"`
		TxID   string `json:"txId,omitempty"`
		After  int64  `json:"after,omitempty,string"`
		Before int64  `json:"before,omitempty,string"`
		Limit  int64  `json:"limit,omitempty,string"`
		State  int    `json:"state,omitempty,string"`
	}
	Withdrawal struct {
		Ccy        string `json:"ccy"`
		Chain      string `json:"chain,omitempty"`
		ToAddr     string `json:"toAddr"`
		ClientID   string `json:"clientId"`
		Amt        string `json:"amt"`
		Dest       int    `json:"dest,string"`
		WalletType string `json:"walletType"`
	}
	GetWithdrawalHistory struct {
		Ccy           string `json:"ccy,omitempty"`
		ClientOrderID string `json:"clientId,omitempty"`
		WithdrawalID  string `json:"wdId,omitempty"`
		TxID          string `json:"txId,omitempty"`
		After         int64  `json:"after,omitempty,string"`
		Before        int64  `json:"before,omitempty,string"`
		Limit         int64  `json:"limit,omitempty,string"`
		State         int    `json:"state,omitempty,string"`
	}
	PiggyBankPurchaseRedemption struct {
		Ccy    string `json:"ccy,omitempty"`
		TxID   string `json:"txId,omitempty"`
		After  int64  `json:"after,omitempty,string"`
		Before int64  `json:"before,omitempty,string"`
		Limit  int64  `json:"limit,omitempty,string"`
		State  int    `json:"state,omitempty,string"`
	}
	GetPiggyBankBalance struct {
		Ccy string `json:"ccy,omitempty"`
	}
)
