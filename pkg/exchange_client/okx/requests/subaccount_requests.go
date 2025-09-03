//nolint:tagliatelle
package requests

type (
	ViewList struct {
		SubAcct string `json:"subAcct,omitempty"`
		Enable  bool   `json:"enable,omitempty"`
		After   int64  `json:"after,omitempty,string"`
		Before  int64  `json:"before,omitempty,string"`
		Limit   int64  `json:"limit,omitempty,string"`
	}
	CreateAPIKey struct {
		Pwd        string   `json:"pwd"`
		SubAcct    string   `json:"subAcct"`
		Label      string   `json:"label"`
		Passphrase string   `json:"Passphrase"`
		IP         []string `json:"ip,omitempty"`
		Perm       string   `json:"perm,omitempty"`
	}
	QueryAPIKey struct {
		APIKey  string `json:"apiKey"`
		SubAcct string `json:"subAcct"`
	}
	DeleteAPIKey struct {
		Pwd     string `json:"pwd"`
		APIKey  string `json:"apiKey"`
		SubAcct string `json:"subAcct"`
	}
	GetBalance struct {
		SubAcct string `json:"subAcct"`
	}
	HistoryTransfer struct {
		Ccy     string `json:"ccy,omitempty"`
		SubAcct string `json:"subAcct,omitempty"`
		After   int64  `json:"after,omitempty,string"`
		Before  int64  `json:"before,omitempty,string"`
		Limit   int64  `json:"limit,omitempty,string"`
		Type    int    `json:"type,omitempty,string"`
	}
	ManageTransfers struct {
		Ccy            string  `json:"ccy"`
		FromSubAccount string  `json:"fromSubAccount"`
		ToSubAccount   string  `json:"tiSubAccount"`
		Amt            float64 `json:"amt,string"`
		From           int     `json:"from,string"`
		To             int     `json:"to,string"`
	}
)
