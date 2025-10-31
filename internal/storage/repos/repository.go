package repos

import (
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_aml_check_history"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_aml_check_queue"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_aml_checks"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_aml_service_keys"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_aml_services"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_aml_supported_assets"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_aml_user_keys"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_analytics"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_currencies"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_currency_exrate"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_exchange_addresses"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_exchange_chains"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_exchange_orders"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_exchange_user_keys"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_exchange_withdrawal_history"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_exchange_withdrawal_settings"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_exchanges"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_invoice_addresses"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_invoices"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_log_types"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_logs"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_multi_withdrawal_rules"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_notification_send_history"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_notification_send_queue"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_notifications"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_personal_access_tokens"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_receipts"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_settings"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_store_api_keys"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_store_currencies"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_store_secrets"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_store_webhooks"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_store_whitelist"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_stores"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_transactions"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_transfer_transactions"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_transfers"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_tron_wallet_balance_statistics"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_unconfirmed_transactions"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_update_balance_queue"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_user_address_book"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_user_exchange_pairs"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_user_exchanges"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_user_notifications"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_user_stores"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_user_verification"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_users"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_wallet_addresses"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_wallet_addresses_activity_logs"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_wallets"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_webhook_send_histories"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_webhook_send_queue"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_withdrawal_from_processing_wallets"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_withdrawal_wallet_addresses"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_withdrawal_wallets"
	"github.com/dv-net/dv-merchant/pkg/database"
	"github.com/dv-net/dv-merchant/pkg/key_value"
)

type IRepository interface {
	Currencies(opts ...Option) repo_currencies.Querier
	Users(opts ...Option) repo_users.ICustomQuerier
	PersonalAccessToken(opts ...Option) repo_personal_access_tokens.Querier
	Stores(opts ...Option) repo_stores.Querier
	StoreAPIKeys(opts ...Option) repo_store_api_keys.Querier
	StoreWebhooks(opts ...Option) repo_store_webhooks.Querier
	StoreCurrencies(opts ...Option) repo_store_currencies.Querier
	StoreWhitelist(opts ...Option) repo_store_whitelist.Querier
	WebHookSendHistories(opts ...Option) repo_webhook_send_histories.ICustomQuerier
	StoreSecrets(opts ...Option) repo_store_secrets.Querier
	Transactions(opts ...Option) repo_transactions.ICustomQuerier
	UnconfirmedTransactions(opts ...Option) repo_unconfirmed_transactions.Querier
	Wallets(opts ...Option) repo_wallets.ICustomQuerier
	WalletAddresses(opts ...Option) repo_wallet_addresses.ICustomQuerier
	WalletAddressesActivityLog(opts ...Option) repo_wallet_addresses_activity_logs.Querier
	Receipts(opts ...Option) repo_receipts.Querier
	Settings(opts ...Option) repo_settings.Querier
	WithdrawalWallets(opts ...Option) repo_withdrawal_wallets.Querier
	WithdrawalWalletAddresses(opts ...Option) repo_withdrawal_wallet_addresses.Querier
	Transfers(opts ...Option) repo_transfers.ICustomQuerier
	WebHookSendQueue(opts ...Option) repo_webhook_send_queue.Querier
	CurrencyExchange() repo_currency_exrate.ICurrencyRateRepo
	KeyValue() key_value.IKeyValue
	Exchanges(opts ...Option) repo_exchanges.Querier
	ExchangeUserKeys(opts ...Option) repo_exchange_user_keys.Querier
	UserStores(opts ...Option) repo_user_stores.Querier
	NotificationSendHistory(opts ...Option) repo_notification_send_history.ICustomQuerier
	NotificationSendQueue(opts ...Option) repo_notification_send_queue.Querier
	WithdrawalsFromProcessing(opts ...Option) repo_withdrawal_from_processing_wallets.Querier
	UserExchangePairs(opts ...Option) repo_user_exchange_pairs.Querier
	UserExchanges(opts ...Option) repo_user_exchanges.ICustomQuerier
	ExchangeOrders(opts ...Option) repo_exchange_orders.ICustomQuerier
	ExchangeAddresses(opts ...Option) repo_exchange_addresses.Querier
	ExchangeWithdrawalHistory(opts ...Option) repo_exchange_withdrawal_history.ICustomQuerier
	ExchangeWithdrawalSettings(opts ...Option) repo_exchange_withdrawal_settings.Querier
	ExchangeChains(opts ...Option) repo_exchange_chains.Querier
	LogTypes(opts ...Option) repo_log_types.Querier
	Logs(opts ...Option) repo_logs.Querier
	UserNotifications(opts ...Option) repo_user_notifications.Querier
	Notifications(opts ...Option) repo_notifications.Querier
	UpdateBalanceQueue(opts ...Option) repo_update_balance_queue.Querier
	MultiWithdrawalRules(opts ...Option) repo_multi_withdrawal_rules.Querier
	TransferTransactions(opts ...Option) repo_transfer_transactions.Querier
	TronWalletBalanceStatistics(opts ...Option) repo_tron_wallet_balance_statistics.Querier
	Analytics(opts ...Option) repo_analytics.Querier
	AmlServices(opts ...Option) repo_aml_services.Querier
	AmlServiceKeys(opts ...Option) repo_aml_service_keys.Querier
	AmlUserKeys(opts ...Option) repo_aml_user_keys.ICustomQuerier
	AmlChecks(opts ...Option) repo_aml_checks.ICustomQuerier
	AmlCheckQueue(opts ...Option) repo_aml_check_queue.Querier
	AmlCheckHistory(opts ...Option) repo_aml_check_history.Querier
	AmlSupportedAssets(opts ...Option) repo_aml_supported_assets.Querier
	UserAddressBook(opts ...Option) repo_user_address_book.Querier
	Invoices(opts ...Option) repo_invoices.Querier
	InvoiceAddresses(opts ...Option) repo_invoice_addresses.Querier
}

type repository struct {
	currencies                  *repo_currencies.Queries
	users                       *repo_users.CustomQuerier
	personalAccessToken         *repo_personal_access_tokens.Queries
	stores                      *repo_stores.Queries
	storeAPIKeys                *repo_store_api_keys.Queries
	storeWebhooks               *repo_store_webhooks.Queries
	storeCurrencies             *repo_store_currencies.Queries
	storeWhitelist              *repo_store_whitelist.Queries
	webhookSendHistories        *repo_webhook_send_histories.CustomQuerier
	transactions                *repo_transactions.CustomQuerier
	unconfirmedTransactions     *repo_unconfirmed_transactions.Queries
	wallets                     *repo_wallets.CustomQuerier
	walletAddresses             *repo_wallet_addresses.CustomQuerier
	walletAddressesActivityLog  *repo_wallet_addresses_activity_logs.Queries
	receipts                    *repo_receipts.Queries
	settings                    *repo_settings.Queries
	transfers                   *repo_transfers.CustomQuerier
	withdrawalWallets           *repo_withdrawal_wallets.Queries
	withdrawalWalletAddresses   *repo_withdrawal_wallet_addresses.Queries
	webhookSendQueue            *repo_webhook_send_queue.Queries
	storeSecrets                *repo_store_secrets.Queries
	currencyExchange            *repo_currency_exrate.Repo
	verificationCodes           *repo_user_verification.Repo
	keyValue                    key_value.IKeyValue
	exchanges                   *repo_exchanges.Queries
	exchangeUserKeys            *repo_exchange_user_keys.Queries
	userStores                  *repo_user_stores.Queries
	notificationSendHistory     *repo_notification_send_history.CustomQuerier
	withdrawalsFromProcessing   *repo_withdrawal_from_processing_wallets.Queries
	userExchangePairs           *repo_user_exchange_pairs.Queries
	userExchanges               *repo_user_exchanges.CustomQuerier
	exchangeOrders              *repo_exchange_orders.CustomQuerier
	exchangeAddresses           *repo_exchange_addresses.Queries
	notificationSendQueue       *repo_notification_send_queue.Queries
	exchangeWithdrawalHistory   *repo_exchange_withdrawal_history.CustomQuerier
	exchangeWithdrawalSettings  *repo_exchange_withdrawal_settings.Queries
	exchangeChains              *repo_exchange_chains.Queries
	logTypes                    *repo_log_types.Queries
	logs                        *repo_logs.Queries
	userNotifications           *repo_user_notifications.Queries
	notifications               *repo_notifications.Queries
	updateBalanceQueue          *repo_update_balance_queue.Queries
	multiWithdrawalRules        *repo_multi_withdrawal_rules.Queries
	transferTransactions        *repo_transfer_transactions.Queries
	tronWalletBalanceStatistics *repo_tron_wallet_balance_statistics.Queries
	analytics                   *repo_analytics.Queries
	amlServices                 *repo_aml_services.Queries
	amlServiceKeys              *repo_aml_service_keys.Queries
	amlUserKeys                 *repo_aml_user_keys.CustomQueries
	amlChecks                   *repo_aml_checks.CustomQuerier
	amlCheckQueue               *repo_aml_check_queue.Queries
	amlCheckHistory             *repo_aml_check_history.Queries
	amlSupportedAssets          *repo_aml_supported_assets.Queries
	userAddressBook             *repo_user_address_book.Queries
	invoices                    *repo_invoices.Queries
	invoiceAddresses            *repo_invoice_addresses.Queries
}

func InitRepository(psql *database.PostgresClient, keyValue key_value.IKeyValue) IRepository {
	return &repository{
		currencies:                  repo_currencies.New(psql.DB),
		users:                       repo_users.NewCustom(psql.DB),
		personalAccessToken:         repo_personal_access_tokens.New(psql.DB),
		stores:                      repo_stores.New(psql.DB),
		storeAPIKeys:                repo_store_api_keys.New(psql.DB),
		storeWebhooks:               repo_store_webhooks.New(psql.DB),
		storeCurrencies:             repo_store_currencies.New(psql.DB),
		storeWhitelist:              repo_store_whitelist.New(psql.DB),
		webhookSendHistories:        repo_webhook_send_histories.NewCustom(psql.DB),
		transactions:                repo_transactions.NewCustom(psql.DB),
		unconfirmedTransactions:     repo_unconfirmed_transactions.New(psql.DB),
		wallets:                     repo_wallets.NewCustom(psql.DB),
		walletAddresses:             repo_wallet_addresses.NewCustom(psql.DB),
		walletAddressesActivityLog:  repo_wallet_addresses_activity_logs.New(psql.DB),
		receipts:                    repo_receipts.New(psql.DB),
		settings:                    repo_settings.New(psql.DB),
		transfers:                   repo_transfers.NewCustom(psql.DB),
		withdrawalWallets:           repo_withdrawal_wallets.New(psql.DB),
		withdrawalWalletAddresses:   repo_withdrawal_wallet_addresses.New(psql.DB),
		webhookSendQueue:            repo_webhook_send_queue.New(psql.DB),
		currencyExchange:            repo_currency_exrate.New(keyValue),
		verificationCodes:           repo_user_verification.New(keyValue),
		exchanges:                   repo_exchanges.New(psql.DB),
		exchangeUserKeys:            repo_exchange_user_keys.New(psql.DB),
		userStores:                  repo_user_stores.New(psql.DB),
		withdrawalsFromProcessing:   repo_withdrawal_from_processing_wallets.New(psql.DB),
		storeSecrets:                repo_store_secrets.New(psql.DB),
		keyValue:                    keyValue,
		notificationSendHistory:     repo_notification_send_history.NewCustom(psql.DB),
		notificationSendQueue:       repo_notification_send_queue.New(psql.DB),
		userExchangePairs:           repo_user_exchange_pairs.New(psql.DB),
		userExchanges:               repo_user_exchanges.NewCustom(psql.DB),
		exchangeOrders:              repo_exchange_orders.NewCustom(psql.DB),
		exchangeAddresses:           repo_exchange_addresses.New(psql.DB),
		logTypes:                    repo_log_types.New(psql.DB),
		logs:                        repo_logs.New(psql.DB),
		exchangeWithdrawalHistory:   repo_exchange_withdrawal_history.NewCustom(psql.DB),
		exchangeWithdrawalSettings:  repo_exchange_withdrawal_settings.New(psql.DB),
		exchangeChains:              repo_exchange_chains.New(psql.DB),
		userNotifications:           repo_user_notifications.New(psql.DB),
		notifications:               repo_notifications.New(psql.DB),
		updateBalanceQueue:          repo_update_balance_queue.New(psql.DB),
		multiWithdrawalRules:        repo_multi_withdrawal_rules.New(psql.DB),
		transferTransactions:        repo_transfer_transactions.New(psql.DB),
		tronWalletBalanceStatistics: repo_tron_wallet_balance_statistics.New(psql.DB),
		analytics:                   repo_analytics.New(psql.DB),
		amlServices:                 repo_aml_services.New(psql.DB),
		amlServiceKeys:              repo_aml_service_keys.New(psql.DB),
		amlUserKeys:                 repo_aml_user_keys.NewCustom(psql.DB),
		amlChecks:                   repo_aml_checks.NewCustom(psql.DB),
		amlCheckQueue:               repo_aml_check_queue.New(psql.DB),
		amlCheckHistory:             repo_aml_check_history.New(psql.DB),
		amlSupportedAssets:          repo_aml_supported_assets.New(psql.DB),
		userAddressBook:             repo_user_address_book.New(psql.DB),
		invoices:                    repo_invoices.New(psql.DB),
		invoiceAddresses:            repo_invoice_addresses.New(psql.DB),
	}
}

func (r *repository) Currencies(opts ...Option) repo_currencies.Querier {
	options := parseOptions(opts...)

	if options.Tx != nil {
		return r.currencies.WithTx(options.Tx)
	}
	return r.currencies
}

func (r *repository) Users(opts ...Option) repo_users.ICustomQuerier {
	options := parseOptions(opts...)
	if options.Tx != nil {
		return r.users.WithTx(options.Tx)
	}
	return r.users
}

func (r *repository) PersonalAccessToken(opts ...Option) repo_personal_access_tokens.Querier {
	options := parseOptions(opts...)
	if options.Tx != nil {
		return r.personalAccessToken.WithTx(options.Tx)
	}
	return r.personalAccessToken
}

func (r *repository) Stores(opts ...Option) repo_stores.Querier {
	options := parseOptions(opts...)
	if options.Tx != nil {
		return r.stores.WithTx(options.Tx)
	}
	return r.stores
}

func (r *repository) StoreAPIKeys(opts ...Option) repo_store_api_keys.Querier {
	options := parseOptions(opts...)
	if options.Tx != nil {
		return r.storeAPIKeys.WithTx(options.Tx)
	}
	return r.storeAPIKeys
}

func (r *repository) StoreWebhooks(opts ...Option) repo_store_webhooks.Querier {
	options := parseOptions(opts...)
	if options.Tx != nil {
		return r.storeWebhooks.WithTx(options.Tx)
	}
	return r.storeWebhooks
}

func (r *repository) StoreCurrencies(opts ...Option) repo_store_currencies.Querier {
	options := parseOptions(opts...)
	if options.Tx != nil {
		return r.storeCurrencies.WithTx(options.Tx)
	}
	return r.storeCurrencies
}

func (r *repository) StoreWhitelist(opts ...Option) repo_store_whitelist.Querier {
	options := parseOptions(opts...)
	if options.Tx != nil {
		return r.storeWhitelist.WithTx(options.Tx)
	}
	return r.storeWhitelist
}

func (r *repository) WebHookSendHistories(opts ...Option) repo_webhook_send_histories.ICustomQuerier {
	options := parseOptions(opts...)
	if options.Tx != nil {
		return r.webhookSendHistories.WithTx(options.Tx)
	}
	return r.webhookSendHistories
}

func (r *repository) Transactions(opts ...Option) repo_transactions.ICustomQuerier {
	options := parseOptions(opts...)
	if options.Tx != nil {
		return r.transactions.WithTx(options.Tx)
	}
	return r.transactions
}

func (r *repository) UnconfirmedTransactions(opts ...Option) repo_unconfirmed_transactions.Querier {
	options := parseOptions(opts...)
	if options.Tx != nil {
		return r.unconfirmedTransactions.WithTx(options.Tx)
	}
	return r.unconfirmedTransactions
}

func (r *repository) Wallets(opts ...Option) repo_wallets.ICustomQuerier {
	options := parseOptions(opts...)
	if options.Tx != nil {
		return r.wallets.WithTx(options.Tx)
	}
	return r.wallets
}

func (r *repository) WalletAddresses(opts ...Option) repo_wallet_addresses.ICustomQuerier {
	options := parseOptions(opts...)
	if options.Tx != nil {
		return r.walletAddresses.WithTx(options.Tx)
	}
	return r.walletAddresses
}

func (r *repository) WalletAddressesActivityLog(opts ...Option) repo_wallet_addresses_activity_logs.Querier {
	options := parseOptions(opts...)
	if options.Tx != nil {
		return r.walletAddressesActivityLog.WithTx(options.Tx)
	}
	return r.walletAddressesActivityLog
}

func (r *repository) Receipts(opts ...Option) repo_receipts.Querier {
	options := parseOptions(opts...)
	if options.Tx != nil {
		return r.receipts.WithTx(options.Tx)
	}
	return r.receipts
}

func (r *repository) Settings(opts ...Option) repo_settings.Querier {
	options := parseOptions(opts...)
	if options.Tx != nil {
		return r.settings.WithTx(options.Tx)
	}
	return r.settings
}

func (r *repository) WithdrawalWallets(opts ...Option) repo_withdrawal_wallets.Querier {
	options := parseOptions(opts...)
	if options.Tx != nil {
		return r.withdrawalWallets.WithTx(options.Tx)
	}
	return r.withdrawalWallets
}

func (r *repository) WithdrawalWalletAddresses(opts ...Option) repo_withdrawal_wallet_addresses.Querier {
	options := parseOptions(opts...)
	if options.Tx != nil {
		return r.withdrawalWalletAddresses.WithTx(options.Tx)
	}
	return r.withdrawalWalletAddresses
}

func (r *repository) Transfers(opts ...Option) repo_transfers.ICustomQuerier {
	options := parseOptions(opts...)
	if options.Tx != nil {
		return r.transfers.WithTx(options.Tx)
	}
	return r.transfers
}

func (r *repository) WebHookSendQueue(opts ...Option) repo_webhook_send_queue.Querier {
	options := parseOptions(opts...)
	if options.Tx != nil {
		return r.webhookSendQueue.WithTx(options.Tx)
	}
	return r.webhookSendQueue
}

func (r *repository) KeyValue() key_value.IKeyValue {
	return r.keyValue
}

func (r *repository) CurrencyExchange() repo_currency_exrate.ICurrencyRateRepo {
	return r.currencyExchange
}

func (r *repository) Exchanges(opts ...Option) repo_exchanges.Querier {
	options := parseOptions(opts...)
	if options.Tx != nil {
		return r.exchanges.WithTx(options.Tx)
	}

	return r.exchanges
}

func (r *repository) ExchangeUserKeys(opts ...Option) repo_exchange_user_keys.Querier {
	options := parseOptions(opts...)
	if options.Tx != nil {
		return r.exchangeUserKeys.WithTx(options.Tx)
	}

	return r.exchangeUserKeys
}

func (r *repository) UserStores(opts ...Option) repo_user_stores.Querier {
	options := parseOptions(opts...)
	if options.Tx != nil {
		return r.userStores.WithTx(options.Tx)
	}

	return r.userStores
}

func (r *repository) NotificationSendHistory(opts ...Option) repo_notification_send_history.ICustomQuerier {
	options := parseOptions(opts...)
	if options.Tx != nil {
		return r.notificationSendHistory.WithTx(options.Tx)
	}

	return r.notificationSendHistory
}

func (r *repository) NotificationSendQueue(opts ...Option) repo_notification_send_queue.Querier {
	options := parseOptions(opts...)
	if options.Tx != nil {
		return r.notificationSendQueue.WithTx(options.Tx)
	}

	return r.notificationSendQueue
}

func (r *repository) WithdrawalsFromProcessing(opts ...Option) repo_withdrawal_from_processing_wallets.Querier {
	options := parseOptions(opts...)
	if options.Tx != nil {
		return r.withdrawalsFromProcessing.WithTx(options.Tx)
	}

	return r.withdrawalsFromProcessing
}

func (r *repository) UserExchangePairs(opts ...Option) repo_user_exchange_pairs.Querier {
	options := parseOptions(opts...)
	if options.Tx != nil {
		return r.userExchangePairs.WithTx(options.Tx)
	}

	return r.userExchangePairs
}

func (r *repository) UserExchanges(opts ...Option) repo_user_exchanges.ICustomQuerier {
	options := parseOptions(opts...)
	if options.Tx != nil {
		return r.userExchanges.WithTx(options.Tx)
	}

	return r.userExchanges
}

func (r *repository) ExchangeOrders(opts ...Option) repo_exchange_orders.ICustomQuerier {
	options := parseOptions(opts...)
	if options.Tx != nil {
		return r.exchangeOrders.WithTx(options.Tx)
	}

	return r.exchangeOrders
}

func (r *repository) ExchangeAddresses(opts ...Option) repo_exchange_addresses.Querier {
	options := parseOptions(opts...)
	if options.Tx != nil {
		return r.exchangeAddresses.WithTx(options.Tx)
	}

	return r.exchangeAddresses
}

func (r *repository) ExchangeWithdrawalHistory(opts ...Option) repo_exchange_withdrawal_history.ICustomQuerier {
	options := parseOptions(opts...)
	if options.Tx != nil {
		return r.exchangeWithdrawalHistory.WithTx(options.Tx)
	}

	return r.exchangeWithdrawalHistory
}

func (r *repository) ExchangeWithdrawalSettings(opts ...Option) repo_exchange_withdrawal_settings.Querier {
	options := parseOptions(opts...)
	if options.Tx != nil {
		return r.exchangeWithdrawalSettings.WithTx(options.Tx)
	}

	return r.exchangeWithdrawalSettings
}

func (r *repository) ExchangeChains(opts ...Option) repo_exchange_chains.Querier {
	options := parseOptions(opts...)
	if options.Tx != nil {
		return r.exchangeChains.WithTx(options.Tx)
	}

	return r.exchangeChains
}

func (r *repository) LogTypes(opts ...Option) repo_log_types.Querier {
	options := parseOptions(opts...)
	if options.Tx != nil {
		return r.logTypes.WithTx(options.Tx)
	}

	return r.logTypes
}

func (r *repository) Logs(opts ...Option) repo_logs.Querier {
	options := parseOptions(opts...)
	if options.Tx != nil {
		return r.logs.WithTx(options.Tx)
	}

	return r.logs
}

func (r *repository) StoreSecrets(opts ...Option) repo_store_secrets.Querier {
	options := parseOptions(opts...)
	if options.Tx != nil {
		return r.storeSecrets.WithTx(options.Tx)
	}

	return r.storeSecrets
}

func (r *repository) UserNotifications(opts ...Option) repo_user_notifications.Querier {
	options := parseOptions(opts...)
	if options.Tx != nil {
		return r.userNotifications.WithTx(options.Tx)
	}

	return r.userNotifications
}

func (r *repository) Notifications(opts ...Option) repo_notifications.Querier {
	options := parseOptions(opts...)
	if options.Tx != nil {
		return r.notifications.WithTx(options.Tx)
	}

	return r.notifications
}

func (r *repository) UpdateBalanceQueue(opts ...Option) repo_update_balance_queue.Querier {
	options := parseOptions(opts...)
	if options.Tx != nil {
		return r.updateBalanceQueue.WithTx(options.Tx)
	}

	return r.updateBalanceQueue
}

func (r *repository) MultiWithdrawalRules(opts ...Option) repo_multi_withdrawal_rules.Querier {
	options := parseOptions(opts...)
	if options.Tx != nil {
		return r.multiWithdrawalRules.WithTx(options.Tx)
	}

	return r.multiWithdrawalRules
}

func (r *repository) TransferTransactions(opts ...Option) repo_transfer_transactions.Querier {
	options := parseOptions(opts...)
	if options.Tx != nil {
		return r.transferTransactions.WithTx(options.Tx)
	}

	return r.transferTransactions
}

func (r *repository) TronWalletBalanceStatistics(opts ...Option) repo_tron_wallet_balance_statistics.Querier {
	options := parseOptions(opts...)
	if options.Tx != nil {
		return r.tronWalletBalanceStatistics.WithTx(options.Tx)
	}

	return r.tronWalletBalanceStatistics
}

func (r *repository) Analytics(opts ...Option) repo_analytics.Querier {
	options := parseOptions(opts...)
	if options.Tx != nil {
		return r.analytics.WithTx(options.Tx)
	}

	return r.analytics
}

func (r *repository) AmlServices(opts ...Option) repo_aml_services.Querier {
	options := parseOptions(opts...)
	if options.Tx != nil {
		return r.amlServices.WithTx(options.Tx)
	}

	return r.amlServices
}

func (r *repository) AmlServiceKeys(opts ...Option) repo_aml_service_keys.Querier {
	options := parseOptions(opts...)
	if options.Tx != nil {
		return r.amlServiceKeys.WithTx(options.Tx)
	}

	return r.amlServiceKeys
}

func (r *repository) AmlUserKeys(opts ...Option) repo_aml_user_keys.ICustomQuerier {
	options := parseOptions(opts...)
	if options.Tx != nil {
		return r.amlUserKeys.WithTx(options.Tx)
	}

	return r.amlUserKeys
}

func (r *repository) AmlChecks(opts ...Option) repo_aml_checks.ICustomQuerier {
	options := parseOptions(opts...)
	if options.Tx != nil {
		return r.amlChecks.WithTx(options.Tx)
	}

	return r.amlChecks
}

func (r *repository) AmlCheckQueue(opts ...Option) repo_aml_check_queue.Querier {
	options := parseOptions(opts...)
	if options.Tx != nil {
		return r.amlCheckQueue.WithTx(options.Tx)
	}

	return r.amlCheckQueue
}

func (r *repository) AmlCheckHistory(opts ...Option) repo_aml_check_history.Querier {
	options := parseOptions(opts...)
	if options.Tx != nil {
		return r.amlCheckHistory.WithTx(options.Tx)
	}

	return r.amlCheckHistory
}

func (r *repository) AmlSupportedAssets(opts ...Option) repo_aml_supported_assets.Querier {
	options := parseOptions(opts...)
	if options.Tx != nil {
		return r.amlSupportedAssets.WithTx(options.Tx)
	}

	return r.amlSupportedAssets
}

func (r *repository) UserAddressBook(opts ...Option) repo_user_address_book.Querier {
	options := parseOptions(opts...)
	if options.Tx != nil {
		return r.userAddressBook.WithTx(options.Tx)
	}

	return r.userAddressBook
}

func (r *repository) Invoices(opts ...Option) repo_invoices.Querier {
	options := parseOptions(opts...)
	if options.Tx != nil {
		return r.invoices.WithTx(options.Tx)
	}

	return r.invoices
}

func (r *repository) InvoiceAddresses(opts ...Option) repo_invoice_addresses.Querier {
	options := parseOptions(opts...)
	if options.Tx != nil {
		return r.invoiceAddresses.WithTx(options.Tx)
	}

	return r.invoiceAddresses
}
