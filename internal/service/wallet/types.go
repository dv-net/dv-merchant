package wallet

import (
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_wallets"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
)

type GetAllByStoreIDResponse struct {
	repo_wallets.GetFullDataByIDRow
	Addresses           []*models.WalletAddress
	AvailableCurrencies []*models.Currency
	Rates               map[string]string
} // @name GetAllByStoreIDResponse

type ProcessingWalletWithAssets struct {
	Address        string                    `json:"address"`
	Blockchain     models.Blockchain         `json:"blockchain"`
	Currency       *models.CurrencyShort     `json:"currency"`
	Balance        *Balance                  `json:"balance,omitempty"`
	Assets         []*Asset                  `json:"assets"`
	AdditionalData *BlockchainAdditionalData `json:"additional_data,omitempty"`
} // @name ProcessingWalletWithAssets

type Balance struct {
	NativeToken    string `json:"native_token"`
	NativeTokenUSD string `json:"native_token_usd"`
} // @name Balance

type Asset struct {
	CurrencyID string `json:"currency_id"`
	Identity   string `json:"identity"`
	Amount     string `json:"amount"`
	AmountUSD  string `json:"amount_usd"`
} // @name Asset

func (a *Asset) IsNativeTokenByBlockChain(blockchain models.Blockchain) bool {
	nativeCurr, _ := blockchain.NativeCurrency()
	return strings.Contains(nativeCurr, a.Identity)
}

type EVMData struct {
	SuggestedGasPrice  string `json:"suggested_gas_price"`
	MaxTransfersERC20  string `json:"max_transfers_erc20"`
	MaxTransfersNative string `json:"max_transfers_native"`
	CostPerERC20       string `json:"cost_per_erc20"`       // Cost in Wei for single ERC20 transfer
	CostPerNative      string `json:"cost_per_native"`      // Cost in Wei for single native transfer
	IsL2               bool   `json:"is_l2"`                // Whether this is an L2 chain
	L1DataFeeEstimate  string `json:"l1_data_fee_estimate"` // Estimated L1 data fee for L2 chains
}

type BlockchainAdditionalData struct {
	TronData *TronData `json:"tron_data,omitempty"`
	EVMData  *EVMData  `json:"evm_data,omitempty"`
} // @name BlockchainAdditionalData

type TronData struct {
	TronTransferData         `json:"tron_transfer_data"`
	AvailableEnergyForUse    string `json:"available_energy_for_use"`
	TotalEnergy              string `json:"total_energy"`
	AvailableBandwidthForUse string `json:"available_bandwidth_for_use"`
	TotalBandwidth           string `json:"total_bandwidth"`
	StackedTrx               string `json:"stacked_trx"`
	StackedBandwidth         string `json:"stacked_bandwidth"`
	StackedEnergy            string `json:"stacked_energy"`
	StackedBandwidthTrx      string `json:"stacked_bandwidth_trx"`
	StackedEnergyTrx         string `json:"stacked_energy_trx"`
	TotalUsedEnergy          string `json:"total_used_energy"`
	TotalUsedBandwidth       string `json:"total_used_bandwidth"`
} // @name TronData

type TronTransferData struct {
	MaxTransfersTRC20  string `json:"max_transfers_trc20"`
	MaxTransfersNative string `json:"max_transfers_native"`
} // @name TronTransferData

type WithAddressDto struct {
	ID              uuid.UUID               `json:"id,omitempty"`
	StoreID         uuid.UUID               `json:"store_id"`
	StoreExternalID string                  `json:"store_external_id"`
	Rates           map[string]string       `json:"rates"`
	Address         []*models.WalletAddress `json:"address"`
	PayURL          *url.URL                `json:"pay_url,omitempty"`
	AmountUSD       string                  `json:"amount_usd"`
	CreatedAt       pgtype.Timestamp        `json:"created_at,omitempty"`
	UpdatedAt       pgtype.Timestamp        `json:"updated_at,omitempty"`
} // @name WithAddressDto

func (o *WithAddressDto) Encode(m *models.Wallet, frontendBaseURL string) (err error) {
	o.ID = m.ID
	o.StoreID = m.StoreID
	o.StoreExternalID = m.StoreExternalID
	o.CreatedAt = m.CreatedAt
	o.UpdatedAt = m.UpdatedAt

	o.PayURL, err = url.Parse(frontendBaseURL)
	if err != nil {
		return err
	}
	o.PayURL.Path = path.Join(o.PayURL.Path, "/pay/wallet/", o.ID.String())

	return err
}

type CurrencyDTO struct {
	ID              string            `json:"id"`
	Code            string            `json:"code"`
	Name            string            `json:"name"`
	Blockchain      models.Blockchain `json:"blockchain"`
	SortOrder       int64             `json:"sort_order"`
	IsNative        bool              `json:"is_native"`
	ContractAddress string            `json:"contract_address"`
} // @name CurrencyDTO

type SummaryDTO struct {
	Currency         CurrencyDTO `json:"currency"`
	Balance          string      `json:"balance"`
	BalanceUSD       string      `json:"balance_usd"`
	Count            int64       `json:"count"`
	CountWithBalance int64       `json:"count_with_balance"`
} // @name SummaryDTO

type BlockchainGroup struct {
	Blockchain models.Blockchain `json:"blockchain"`
	Assets     []AssetWallet     `json:"assets"`
} // @name BlockchainGroup

type AssetWallet struct {
	Currency     string          `json:"currency"`
	Amount       decimal.Decimal `json:"amount"`
	AmountUSD    decimal.Decimal `json:"amount_usd"`
	TxCount      decimal.Decimal `json:"tx_count"`
	TotalDeposit decimal.Decimal `json:"total_deposit"`
} // @name AssetWallet

type AddressLog struct {
	Text          string                 `json:"text"`
	TextVariables map[string]interface{} `json:"text_variables"`
	CreatedAt     *time.Time             `json:"created_at"`
	UpdatedAt     *time.Time             `json:"updated_at"`
} // @name WalletAddressLog

type WithBlockchains struct {
	WalletCreatedAt time.Time         `json:"wallet_created_at"`
	WalletID        uuid.UUID         `json:"wallet_id"`
	StoreExternalID string            `json:"store_external_id"`
	StoreID         uuid.UUID         `json:"store_id"`
	StoreName       string            `json:"store_name"`
	Email           *string           `json:"email"`
	Address         string            `json:"address"`
	TotalTx         decimal.Decimal   `json:"total_tx"`
	Blockchains     []BlockchainGroup `json:"blockchains"`
	Logs            []*AddressLog     `json:"logs"`
} // @name WalletWithBlockchains

type HotKeyCsv struct {
	Blockchain string `json:"blockchain" csv:"blockchain"`
	PublicKey  string `json:"public_key" csv:"public_key"`
	PrivateKey string `json:"private_key" csv:"private_key"`
	Address    string `json:"address" csv:"address"`
}

type GetOwnerHotWalletKeysParams struct {
	WalletAddressIDs           []uuid.UUID
	ExcludedWalletAddressesIDs []uuid.UUID
}
