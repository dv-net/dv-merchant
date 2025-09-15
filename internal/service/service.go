package service

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/dv-net/dv-merchant/internal/service/aml"
	"github.com/dv-net/dv-merchant/internal/service/analytics"
	"github.com/dv-net/dv-merchant/internal/service/notification_settings"
	"github.com/dv-net/dv-merchant/internal/tools"
	amlproviders "github.com/dv-net/dv-merchant/pkg/aml"
	"github.com/dv-net/dv-merchant/pkg/aml/providers"
	"github.com/dv-net/dv-merchant/pkg/aml/providers/aml_bot"
	"github.com/dv-net/dv-merchant/pkg/aml/providers/bitok"
	"github.com/dv-net/dv-merchant/pkg/otp"
	"github.com/dv-net/dv-merchant/pkg/turnstile"

	"github.com/dv-net/dv-merchant/internal/metrics"
	"github.com/dv-net/dv-merchant/pkg/admin_gateway"

	"github.com/dv-net/dv-merchant/internal/cache"
	"github.com/dv-net/dv-merchant/internal/config"
	"github.com/dv-net/dv-merchant/internal/event"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/address"
	"github.com/dv-net/dv-merchant/internal/service/address_book"
	"github.com/dv-net/dv-merchant/internal/service/admin"
	"github.com/dv-net/dv-merchant/internal/service/auth"
	"github.com/dv-net/dv-merchant/internal/service/callback"
	"github.com/dv-net/dv-merchant/internal/service/currconv"
	"github.com/dv-net/dv-merchant/internal/service/currency"
	"github.com/dv-net/dv-merchant/internal/service/dictionary"
	"github.com/dv-net/dv-merchant/internal/service/eproxy"
	"github.com/dv-net/dv-merchant/internal/service/exchange"
	"github.com/dv-net/dv-merchant/internal/service/exchange_manager"
	"github.com/dv-net/dv-merchant/internal/service/exchange_rules"
	"github.com/dv-net/dv-merchant/internal/service/exchange_withdrawal"
	"github.com/dv-net/dv-merchant/internal/service/exrate"
	"github.com/dv-net/dv-merchant/internal/service/log"
	"github.com/dv-net/dv-merchant/internal/service/notification_sender"
	"github.com/dv-net/dv-merchant/internal/service/notification_sender/external_sender"
	"github.com/dv-net/dv-merchant/internal/service/notification_sender/mail_sender"
	"github.com/dv-net/dv-merchant/internal/service/notification_sender/telegram_sender"
	"github.com/dv-net/dv-merchant/internal/service/notify"
	"github.com/dv-net/dv-merchant/internal/service/permission"
	"github.com/dv-net/dv-merchant/internal/service/processing"
	"github.com/dv-net/dv-merchant/internal/service/receipts"
	"github.com/dv-net/dv-merchant/internal/service/setting"
	"github.com/dv-net/dv-merchant/internal/service/store"
	"github.com/dv-net/dv-merchant/internal/service/system"
	"github.com/dv-net/dv-merchant/internal/service/templater"
	"github.com/dv-net/dv-merchant/internal/service/transactions"
	"github.com/dv-net/dv-merchant/internal/service/updater"
	"github.com/dv-net/dv-merchant/internal/service/user"
	"github.com/dv-net/dv-merchant/internal/service/wallet"
	"github.com/dv-net/dv-merchant/internal/service/webhook"
	"github.com/dv-net/dv-merchant/internal/service/withdraw"
	"github.com/dv-net/dv-merchant/internal/service/withdrawal_wallet"
	"github.com/dv-net/dv-merchant/internal/storage"
	"github.com/dv-net/dv-merchant/pkg/logger"
	"github.com/dv-net/dv-merchant/pkg/rate"

	epr "github.com/dv-net/dv-proto/go/eproxy"
)

type Services struct {
	UserService                   user.IUser
	UserCredentialsService        user.IUserCredentials
	AuthService                   auth.IAuth
	CurrencyService               currency.ICurrency
	ExRateService                 exrate.IExRateSource
	CurrConvService               currconv.ICurrencyConvertor
	PermissionService             permission.IPermission
	WebHookService                webhook.IWebHook
	StoreService                  store.IStore
	StoreAPIKeyService            store.IStoreAPIKey
	StoreWebhooksService          store.IStoreWebhooks
	StoreCurrencyService          store.IStoreCurrency
	StoreWhitelistService         store.IStoreWhitelist
	StoreSecretService            store.ISecret
	TransactionService            transactions.ITransaction
	WalletTransactionService      transactions.IWalletTransaction
	UnconfirmedTransactionService transactions.IUnconfirmedTransaction
	WalletRestorer                transactions.TxRestorer
	WalletService                 wallet.IWalletService
	WalletConverter               wallet.IWalletAddressConverter
	WalletBalanceService          wallet.IWalletBalances
	BalanceUpdater                wallet.BalanceUpdater
	AddressesService              address.IWalletAddressService
	AddressBookService            address_book.IAddressBookService
	ExplorerProxyService          eproxy.IExplorerProxy
	ReceiptService                receipts.IReceiptService
	SettingService                setting.ISettingService
	ProcessingWallet              processing.IProcessingWallet
	ProcessingOwnerService        processing.IProcessingOwner
	ProcessingClientService       processing.IProcessingClient
	ProcessingTransferService     processing.IProcessingTransfer
	ProcessingSystemService       processing.IProcessingSystem
	ProcessingService             processing.IProcessingService
	WithdrawalWalletService       withdrawal_wallet.IWithdrawalWalletService
	WithdrawService               withdraw.IWithdrawService
	DictionaryService             dictionary.IDictionaryService
	CallbackService               callback.ICallback
	NotificationService           notify.INotificationService
	AdminService                  admin.IAdmin
	ExchangeService               exchange.IExchangeService
	ExchangeManager               exchange_manager.IExchangeManager
	ExchangeWithdrawalService     exchange_withdrawal.IExchangeWithdrawalService
	SystemService                 system.ISystemService
	TemplaterService              templater.ITemplaterService
	UnconfirmedCollapser          transactions.IUnconfirmedTransactionCollapser
	ExchangeRulesService          exchange_rules.IExchangeRules
	LogService                    log.ILogService
	NotificationSettings          notification_settings.INotificationSettings
	UpdaterService                updater.IUpdater
	TurnstileVerifier             turnstile.Verifier
	AnalyticsService              analytics.IAnalytics
	AMLService                    aml.IService
	AMLKeysService                aml.KeysService
	AMLStatusChecker              aml.StatusChecker
}

func NewServices(
	ctx context.Context,
	conf *config.Config,
	storage storage.IStorage,
	cache cache.ICache,
	logger logger.Logger,
	appVersion string,
	commitHash string,
) (*Services, error) {
	logService := log.NewService(storage)
	currencyService := currency.New(conf, storage)
	eventListener := event.New()

	adminSvc := admin_gateway.New(conf.Admin.BaseURL, appVersion, logger, conf.Admin.LogStatus)
	exrateService, err := exrate.New(conf, currencyService, logger, storage, adminSvc)
	if err != nil {
		logger.Error("init currency exchange rate service failed", err)
		return nil, err
	}

	currConvService := currconv.New(exrateService)
	permissionService, err := permission.New(conf, storage.PSQLConn())
	if err != nil {
		logger.Error("init permission service", err)
		return nil, err
	}
	if err := permissionService.LoadPolicies(conf.RolesPoliciesPath); err != nil {
		logger.Error("load permission policies", err)
		return nil, err
	}

	webhookService := webhook.New(conf.WebHook, storage, logger)
	eprClient, err := epr.NewClient(conf.EProxy.GRPC.Addr)
	if err != nil {
		logger.Error("init epr client", err)
		return nil, err
	}

	eProxyService := eproxy.New(eprClient)
	receiptService := receipts.New(storage, currencyService)

	processingMetrics, err := metrics.New()
	if err != nil {
		return nil, fmt.Errorf("register metrics: %w", err)
	}

	settingService := setting.New(conf, storage, cache, eventListener, logger)
	processingService := processing.New(
		ctx,
		logger,
		currencyService,
		currConvService,
		settingService,
		eventListener,
		storage,
		processingMetrics,
		appVersion,
	)

	notificationDrivers := make(map[models.DeliveryChannel]notification_sender.IInternalSender, 2)
	templaterService := templater.New(ctx, logger, settingService)
	mailSender, err := mail_sender.New(logger, eventListener, templaterService, settingService)
	if err != nil {
		logger.Error("initialize mailer", err)
	}
	if mailSender != nil {
		notificationDrivers[models.EmailDeliveryChannel] = mailSender
	}

	notificationDrivers[models.TelegramDeliveryChannel] = telegram_sender.NewService()

	externalNotificationSender := external_sender.New(adminSvc, settingService)
	notificationSender := notification_sender.New(logger, notificationDrivers, settingService, externalNotificationSender)
	notificationService := notify.New(logger, storage, settingService, notificationSender, permissionService)

	transactionService := transactions.New(logger, storage, eProxyService, currConvService, eventListener, notificationService)
	addressesService := address.New(conf, storage, logger, processingService)
	withdrawalWalletService := withdrawal_wallet.New(storage, logger, currencyService, currConvService, processingService)
	addressBookService := address_book.New(storage, logger, currencyService, withdrawalWalletService, processingService)
	walletService := wallet.New(conf, storage, logger, currencyService, processingService, exrateService, currConvService, settingService, eProxyService, notificationService)
	storeRateLimiter := rate.NewLimiter(
		storage.KeyValue(),
		rate.WithMaxLimit(conf.ExternalStoreLimits.MaxRequestsPerInterval),
		rate.WithDuration(conf.ExternalStoreLimits.RateLimitInterval),
	)

	storeService := store.New(storage, currencyService, logger, webhookService, eventListener, exrateService, walletService, notificationService, storeRateLimiter, conf.ExternalStoreLimits.Enabled, processingService)
	otpSvc := otp.New(&otp.Config{TTL: time.Minute * 10}, tools.RandomCodeGenerator, storage.KeyValue())
	userService := user.New(conf, storage, storeService, permissionService, processingService, notificationService, logger, settingService, adminSvc, otpSvc)

	adminService := admin.New(conf, storage, logger, permissionService, userService, notificationService)

	authService := auth.New(conf, logger, storage, userService, userService, notificationService, settingService)
	withdrawService := withdraw.New(storage, logger, processingService, processingService, currConvService, currencyService, exrateService, settingService)
	updaterClient, _ := updater.NewClient(logger, conf)
	upd := updater.New(logger, conf, processingService, appVersion)
	analyticsService := analytics.NewService(storage, cache, settingService, adminSvc, processingService, updaterClient, appVersion, commitHash)
	systemService := system.New(settingService, permissionService, adminSvc, logger, appVersion, commitHash, conf, analyticsService)

	dictionaryService := dictionary.New(storage, exrateService, systemService)
	callbackService := callback.New(logger, eventListener, storage, transactionService, transactionService, storeService, currConvService, receiptService)

	exchangeManager := exchange_manager.NewManager(logger, storage, currConvService)
	exchangeRulesService := exchange_rules.NewService(logger, storage, exchangeManager)
	exchangeService := exchange.NewService(logger, storage, exchangeManager, exchangeRulesService, settingService)
	exchangeWithdrawalService := exchange_withdrawal.NewService(logger, storage, exchangeManager, currConvService, exchangeRulesService, settingService)

	notificationSettings := notification_settings.New(storage)

	turnstileVerif, err := turnstile.New(conf.Turnstile.BaseURL, conf.Turnstile.Secret, conf.Turnstile.Enabled)
	if err != nil {
		return nil, err
	}

	amlService, err := prepareAMLService(storage, logger, conf.AML)
	if err != nil {
		return nil, err
	}

	return &Services{
		UserService:                   userService,
		UserCredentialsService:        userService,
		AuthService:                   authService,
		CurrencyService:               currencyService,
		ExRateService:                 exrateService,
		CurrConvService:               currConvService,
		PermissionService:             permissionService,
		WebHookService:                webhookService,
		StoreService:                  storeService,
		StoreAPIKeyService:            storeService,
		StoreWebhooksService:          storeService,
		StoreCurrencyService:          storeService,
		StoreWhitelistService:         storeService,
		StoreSecretService:            storeService,
		TransactionService:            transactionService,
		WalletTransactionService:      transactionService,
		UnconfirmedTransactionService: transactionService,
		WalletRestorer:                transactionService,
		WalletService:                 walletService,
		WalletBalanceService:          walletService,
		BalanceUpdater:                walletService,
		WalletConverter:               walletService,
		AddressesService:              addressesService,
		AddressBookService:            addressBookService,
		ExplorerProxyService:          eProxyService,
		ReceiptService:                receiptService,
		SettingService:                settingService,
		ProcessingWallet:              processingService,
		ProcessingOwnerService:        processingService,
		ProcessingClientService:       processingService,
		ProcessingTransferService:     processingService,
		ProcessingSystemService:       processingService,
		ProcessingService:             processingService,
		WithdrawalWalletService:       withdrawalWalletService,
		WithdrawService:               withdrawService,
		DictionaryService:             dictionaryService,
		CallbackService:               callbackService,
		NotificationService:           notificationService,
		AdminService:                  adminService,
		ExchangeService:               exchangeService,
		ExchangeManager:               exchangeManager,
		SystemService:                 systemService,
		TemplaterService:              templaterService,
		UnconfirmedCollapser:          transactionService,
		LogService:                    logService,
		ExchangeWithdrawalService:     exchangeWithdrawalService,
		ExchangeRulesService:          exchangeRulesService,
		NotificationSettings:          notificationSettings,
		UpdaterService:                upd,
		TurnstileVerifier:             turnstileVerif,
		AnalyticsService:              analyticsService,
		AMLService:                    amlService,
		AMLKeysService:                amlService,
		AMLStatusChecker:              amlService,
	}, nil
}

func prepareAMLService(st storage.IStorage, l logger.Logger, conf config.AML) (*aml.Service, error) {
	bitokBaseURL, err := url.Parse(conf.BitOK.BaseURL)
	if err != nil {
		return nil, err
	}

	amlBotBaseURL, err := url.Parse(conf.AMLBot.BaseURL)
	if err != nil {
		return nil, err
	}

	amlProviderFactory := providers.NewFactory()

	if conf.BitOK.Enabled {
		amlProviderFactory.RegisterProvider(
			amlproviders.ProviderSlugBitOK,
			bitok.NewBitOK(bitokBaseURL, l),
			providers.CreateBitOKAuthorizer(),
		)
	}

	if conf.AMLBot.Enabled {
		amlProviderFactory.RegisterProvider(
			amlproviders.ProviderSlugAMLBot,
			aml_bot.NewAMBot(amlBotBaseURL, l),
			providers.CreateAMLBotAuthorizer(),
		)
	}

	return aml.NewService(st, amlProviderFactory, l, conf), nil
}
