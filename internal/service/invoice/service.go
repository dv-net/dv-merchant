package invoice

import (
	"context"
	"errors"
	"time"

	"github.com/dv-net/dv-merchant/internal/constant"
	"github.com/dv-net/dv-merchant/internal/dto"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/currconv"
	"github.com/dv-net/dv-merchant/internal/service/exrate"
	"github.com/dv-net/dv-merchant/internal/service/user"
	"github.com/dv-net/dv-merchant/internal/service/wallet"
	"github.com/dv-net/dv-merchant/internal/storage"
	"github.com/dv-net/dv-merchant/internal/storage/repos"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_invoice_addresses"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_invoices"
	"github.com/dv-net/dv-merchant/pkg/dbutils/pgerror"
	"github.com/dv-net/dv-merchant/pkg/logger"
	"github.com/dv-net/dv-merchant/pkg/pgtypeutils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
)

const (
	ExpirationTime = 20 * time.Minute
)

type IInvoiceService interface {
	CreateInvoice(ctx context.Context, d dto.CreateInvoiceDTO) (*models.Invoice, error)
	GetInvoice(ctx context.Context, id uuid.UUID) (*dto.InvoiceInfoDTO, error)
	AttachAddress(ctx context.Context, invoiceID uuid.UUID, cur *models.Currency, store *models.Store) (*dto.InvoiceAddressDTO, error)
	UpdateStatus(ctx context.Context, invoiceID uuid.UUID, status constant.InvoiceStatus, opts ...repos.Option) error
	Run(ctx context.Context)
}

type service struct {
	logger          logger.Logger
	storage         storage.IStorage
	walletService   wallet.IWalletService
	currConvService currconv.ICurrencyConvertor
	userService     user.IUser
	exRateService   exrate.IExRateSource
}

func New(
	logger logger.Logger,
	st storage.IStorage,
	wService wallet.IWalletService,
	curService currconv.ICurrencyConvertor,
	uService user.IUser,
	exRateService exrate.IExRateSource,
) IInvoiceService {
	return &service{
		logger:          logger,
		storage:         st,
		walletService:   wService,
		currConvService: curService,
		userService:     uService,
		exRateService:   exRateService,
	}
}

func (s *service) CreateInvoice(ctx context.Context, d dto.CreateInvoiceDTO) (*models.Invoice, error) {
	invoice, err := s.storage.Invoices().Create(ctx, repo_invoices.CreateParams{
		UserID:            d.User.ID,
		StoreID:           d.Store.ID,
		OrderID:           d.OrderID,
		ExpectedAmountUsd: d.AmountUSD,
		Status:            constant.InvoiceStatusPending,
		ExpiresAt:         pgtype.Timestamptz(pgtypeutils.EncodeTime(time.Now().Add(ExpirationTime))),
	})
	if err != nil {
		parsedErr := pgerror.ParseError(err)
		return nil, parsedErr
	}
	return invoice, nil
}

func (s *service) GetInvoice(ctx context.Context, id uuid.UUID) (*dto.InvoiceInfoDTO, error) {
	invoice, err := s.storage.Invoices().GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	invoiceResp, err := s.storage.InvoiceAddresses().GetByInvoiceID(ctx, id)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	addresses := make([]*dto.InvoiceAddressDTO, 0, len(invoiceResp))
	for _, addr := range invoiceResp {
		addresses = append(addresses, &dto.InvoiceAddressDTO{
			Address:        addr.Address.String,
			CurrencyID:     addr.CurrencyID.String,
			Blockchain:     addr.Blockchain,
			RateAtCreation: addr.RateAtCreation,
		})
	}

	return &dto.InvoiceInfoDTO{
		Invoice:        invoice,
		InvoiceAddress: addresses,
	}, nil
}

func (s *service) AttachAddress(
	ctx context.Context,
	invoiceID uuid.UUID,
	cur *models.Currency,
	store *models.Store,
) (*dto.InvoiceAddressDTO, error) {
	var createdAddress *dto.InvoiceAddressDTO

	err := repos.BeginTxFunc(ctx, s.storage.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		invoice, err := s.storage.Invoices(repos.WithTx(tx)).GetByID(ctx, invoiceID)
		if err != nil {
			return NewErrInvoiceNotFound(invoiceID)
		}

		if invoice.ExpiresAt.Valid && invoice.ExpiresAt.Time.Before(time.Now()) {
			return ErrInvoiceExpired
		}

		// Check if address with this currency already exists
		existingAddress, err := s.storage.InvoiceAddresses(repos.WithTx(tx)).GetInvoiceAddressByInvoiceAndCurrency(ctx,
			repo_invoice_addresses.GetInvoiceAddressByInvoiceAndCurrencyParams{
				InvoiceID:  invoiceID,
				CurrencyID: cur.ID,
			})
		if err == nil {
			// Address already exists, return it
			createdAddress = &dto.InvoiceAddressDTO{
				Address:        existingAddress.Address.String,
				CurrencyID:     existingAddress.CurrencyID.String,
				Blockchain:     existingAddress.Blockchain,
				RateAtCreation: existingAddress.RateAtCreation,
			}
			return nil
		} else if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}

		usr, err := s.userService.GetUserByID(ctx, invoice.UserID)
		if err != nil {
			return err
		}

		var walletAddress *models.WalletAddress

		invoiceAddresses, err := s.storage.InvoiceAddresses(repos.WithTx(tx)).GetByInvoiceID(ctx, invoiceID)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return err
		}

		if len(invoiceAddresses) > 0 {
			for _, invAddr := range invoiceAddresses {
				if invAddr.Blockchain == *cur.Blockchain {
					existingWA, err := s.walletService.GetByAccountAndCurrency(ctx, invAddr.AccountID.UUID, cur.ID)
					if err != nil && !errors.Is(err, pgx.ErrNoRows) {
						return err
					}

					if err == nil && existingWA.Status == constant.WalletStatusAvailable && !existingWA.Dirty {
						walletAddress = existingWA
						s.logger.Infow("reusing existing address from same blockchain",
							"address", existingWA.Address,
							"blockchain", existingWA.Blockchain,
							"old_currency", invAddr.CurrencyID.String,
							"new_currency", cur.ID)
						break
					}
				}
			}
		}

		if walletAddress == nil {
			walletAddress, err = s.walletService.GetAvailable(ctx, usr, store, cur)
			if err != nil {
				s.logger.Errorw("failed to get available wallet address", "error", err)
				return err
			}
		}

		err = s.walletService.Reserve(ctx, walletAddress, repos.WithTx(tx))
		if err != nil {
			s.logger.Errorw("failed to reserve wallet address", "error", err)
			return err
		}

		rate, err := s.exRateService.GetCurrencyRate(ctx, usr.RateSource.String(), models.CurrencyCodeUSDT, cur.Code, usr.RateScale)
		if err != nil {
			return err
		}
		rateDecimal, err := decimal.NewFromString(rate)
		if err != nil {
			return err
		}
		newAddress, err := s.storage.InvoiceAddresses(repos.WithTx(tx)).CreateWithWalletAddress(ctx, repo_invoice_addresses.CreateWithWalletAddressParams{
			InvoiceID:       invoice.ID,
			WalletAddressID: walletAddress.ID,
			RateAtCreation:  decimal.NullDecimal{Valid: true, Decimal: rateDecimal},
		})
		if err != nil {
			return err
		}

		createdAddress = &dto.InvoiceAddressDTO{
			Address:        newAddress.Address.String,
			CurrencyID:     newAddress.CurrencyID.String,
			Blockchain:     newAddress.Blockchain,
			RateAtCreation: newAddress.RateAtCreation,
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return createdAddress, nil
}

func (s *service) UpdateStatus(ctx context.Context, invoiceID uuid.UUID, status constant.InvoiceStatus, opts ...repos.Option) error {
	_, err := s.storage.Invoices(opts...).UpdateStatus(ctx, repo_invoices.UpdateStatusParams{
		ID:     invoiceID,
		Status: status,
	})
	if err != nil {
		return err
	}
	return nil
}

// Run starts a ticker that expires invoices every minute
func (s *service) Run(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	s.logger.Info("Invoice expiration ticker started")

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Invoice expiration ticker stopped")
			return
		case <-ticker.C:
			s.expireInvoices(ctx)
		}
	}
}

func (s *service) expireInvoices(ctx context.Context) {
	// Update expired invoices status
	err := s.storage.Invoices().ExpireInvoices(ctx)
	if err != nil {
		s.logger.Errorw("failed to expire invoices", "error", err)
		return
	}

	// Get expired invoices to release their wallet addresses
	expiredInvoices, err := s.storage.Invoices().GetExpiredInvoices(ctx)
	if err != nil {
		s.logger.Errorw("failed to get expired invoices", "error", err)
		return
	}

	if len(expiredInvoices) == 0 {
		return
	}

	s.logger.Infow("expiring invoices", "count", len(expiredInvoices))

	// Release wallet addresses for expired invoices
	for _, invoice := range expiredInvoices {
		// Get invoice addresses
		invoiceAddresses, err := s.storage.InvoiceAddresses().GetByInvoiceID(ctx, invoice.ID)
		if err != nil {
			s.logger.Errorw("failed to get invoice addresses",
				"invoice_id", invoice.ID,
				"error", err)
			continue
		}

		// Release each wallet address
		for _, invAddr := range invoiceAddresses {
			if !invAddr.AccountID.Valid {
				continue
			}

			err = s.walletService.ReleaseByAccountID(ctx, invAddr.AccountID.UUID)
			if err != nil {
				s.logger.Errorw("failed to release wallet address",
					"invoice_id", invoice.ID,
					"account_id", invAddr.AccountID.UUID,
					"error", err)
				continue
			}

			s.logger.Debugw("wallet address released for expired invoice",
				"invoice_id", invoice.ID,
				"account_id", invAddr.AccountID.UUID)
		}
	}
}
