package address_book

import (
	"context"
	"errors"
	"fmt"

	"github.com/dv-net/dv-merchant/internal/delivery/http/responses/withdrawal_response"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/currency"
	"github.com/dv-net/dv-merchant/internal/service/processing"
	"github.com/dv-net/dv-merchant/internal/service/withdrawal_wallet"
	"github.com/dv-net/dv-merchant/internal/storage"
	"github.com/dv-net/dv-merchant/internal/storage/repos"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_user_address_book"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_withdrawal_wallet_addresses"
	"github.com/dv-net/dv-merchant/pkg/logger"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

//nolint:interfacebloat
type IAddressBookService interface {
	GetUserAddresses(ctx context.Context, userID uuid.UUID) (*withdrawal_response.AddressBookListResponse, error)
	GetAddressByID(ctx context.Context, id uuid.UUID) (*models.UserAddressBook, error)
	GetUserAddressesByCurrency(ctx context.Context, userID uuid.UUID, currencyID string) ([]*models.UserAddressBook, error)
	GetUserAddressesByBlockchain(ctx context.Context, userID uuid.UUID, blockchain models.Blockchain) ([]*models.UserAddressBook, error)
	CreateAddress(ctx context.Context, params CreateAddressDTO) (*models.UserAddressBook, error)
	UpdateAddress(ctx context.Context, userID uuid.UUID, id uuid.UUID, params UpdateAddressDTO) (*models.UserAddressBook, error)
	DeleteAddress(ctx context.Context, dto DeleteAddressDTO, usr *models.User) error
	CheckAddressExists(ctx context.Context, userID uuid.UUID, address string, currencyID string) (bool, error)
	IsUniversalAddress(ctx context.Context, userID uuid.UUID, address string, blockchain models.Blockchain) (bool, int, error)
	GetUniversalAddressGroup(ctx context.Context, userID uuid.UUID, address string, blockchain models.Blockchain) ([]*models.UserAddressBook, error)
	AddWithdrawalRule(ctx context.Context, dto AddWithdrawalRuleDTO) error
	CheckWithdrawalRuleExists(ctx context.Context, entry *models.UserAddressBook, usr *models.User) (bool, error)
}

type CreateAddressDTO struct {
	UserID               uuid.UUID
	Address              string
	CurrencyID           string
	Universal            bool
	EVM                  bool
	Name                 *string
	Tag                  *string
	Blockchain           *models.Blockchain
	CreateWithdrawalRule bool
	TOTP                 string
}

type UpdateAddressDTO struct {
	Name *string
	Tag  *string
}

type Service struct {
	storage                 storage.IStorage
	logger                  logger.Logger
	currencyService         currency.ICurrency
	withdrawalWalletService withdrawal_wallet.IWithdrawalWalletService
	processingWalletService processing.IProcessingWallet
}

func New(
	storage storage.IStorage,
	logger logger.Logger,
	currencyService currency.ICurrency,
	withdrawalWalletService withdrawal_wallet.IWithdrawalWalletService,
	processingWalletService processing.IProcessingWallet,
) *Service {
	return &Service{
		storage:                 storage,
		logger:                  logger,
		currencyService:         currencyService,
		withdrawalWalletService: withdrawalWalletService,
		processingWalletService: processingWalletService,
	}
}

func (s *Service) CreateAddress(ctx context.Context, dto CreateAddressDTO) (*models.UserAddressBook, error) {
	// Validate the address format and business rules
	if err := s.ValidateCreateAddressDTO(ctx, dto); err != nil {
		return nil, fmt.Errorf("address validation failed: %w", err)
	}
	if dto.EVM {
		return s.createEVMAddress(ctx, dto)
	}
	if dto.Universal {
		return s.createUniversalAddress(ctx, dto)
	}
	return s.createSingleAddress(ctx, dto)
}

// CheckWithdrawalRuleExists checks if a withdrawal rule exists and is active for a given address book entry
func (s *Service) CheckWithdrawalRuleExists(ctx context.Context, entry *models.UserAddressBook, usr *models.User) (bool, error) {
	// Get withdrawal wallet for this user and currency
	withdrawalWallet, err := s.withdrawalWalletService.GetWithdrawalWalletsByCurrencyID(ctx, usr, entry.CurrencyID)
	if err != nil {
		// If withdrawal wallet doesn't exist, rule doesn't exist
		return false, nil //nolint:nilerr
	}

	// Check if withdrawal wallet address exists for this address
	withdrawalAddress, err := s.storage.WithdrawalWalletAddresses().GetByAddressWithTrashed(ctx, repo_withdrawal_wallet_addresses.GetByAddressWithTrashedParams{
		WithdrawalWalletID: withdrawalWallet.ID,
		Address:            entry.Address,
	})
	if err != nil {
		// If withdrawal wallet address doesn't exist, rule doesn't exist
		return false, nil //nolint:nilerr
	}

	// Rule exists and is active only if DeletedAt is null
	return !withdrawalAddress.DeletedAt.Valid, nil
}

func (s *Service) toUniversalAddressGroupResponse(ctx context.Context, entries []*models.UserAddressBook) *withdrawal_response.UniversalAddressGroupResponse {
	if len(entries) == 0 {
		return nil
	}

	first := entries[0]

	// Convert all entries to currency responses with individual withdrawal rule status
	currencies := make([]*withdrawal_response.AddressBookEntryResponseShort, len(entries))
	for i, entry := range entries {
		withdrawalRuleExists := false
		if ruleExists, err := s.CheckWithdrawalRuleExists(ctx, entry, nil); err == nil {
			withdrawalRuleExists = ruleExists
		}
		currencies[i] = &withdrawal_response.AddressBookEntryResponseShort{
			ID:                   entry.ID,
			CurrencyID:           entry.CurrencyID,
			WithdrawalRuleExists: withdrawalRuleExists,
		}
	}

	var submittedAt string
	if first.SubmittedAt.Valid {
		submittedAt = first.SubmittedAt.Time.Format("2006-01-02T15:04:05Z")
	}

	response := &withdrawal_response.UniversalAddressGroupResponse{
		Address:       first.Address,
		Blockchain:    *first.Blockchain,
		IsUniversal:   true,
		Currencies:    currencies,
		SubmittedAt:   submittedAt,
		CurrencyCount: len(entries),
	}

	if first.Name.Valid {
		response.Name = &first.Name.String
	}
	if first.Tag.Valid {
		response.Tag = &first.Tag.String
	}

	return response
}

func (s *Service) toEVMAddressGroupResponse(ctx context.Context, entries []*models.UserAddressBook) *withdrawal_response.EVMAddressGroupResponse {
	if len(entries) == 0 {
		return nil
	}

	first := entries[0]

	// Convert all entries to currency responses with individual withdrawal rule status
	currencies := make([]*withdrawal_response.AddressBookEntryResponseShort, len(entries))
	for i, entry := range entries {
		withdrawalRuleExists := false
		if ruleExists, err := s.CheckWithdrawalRuleExists(ctx, entry, nil); err == nil {
			withdrawalRuleExists = ruleExists
		}

		currencies[i] = &withdrawal_response.AddressBookEntryResponseShort{
			ID:                   entry.ID,
			CurrencyID:           entry.CurrencyID,
			WithdrawalRuleExists: withdrawalRuleExists,
		}
	}

	// Collect unique blockchains
	blockchainSet := make(map[models.Blockchain]bool)
	for _, entry := range entries {
		blockchainSet[*entry.Blockchain] = true
	}

	blockchains := make([]models.Blockchain, 0, len(blockchainSet))
	for blockchain := range blockchainSet {
		blockchains = append(blockchains, blockchain)
	}

	var submittedAt string
	if first.SubmittedAt.Valid {
		submittedAt = first.SubmittedAt.Time.Format("2006-01-02T15:04:05Z")
	}

	response := &withdrawal_response.EVMAddressGroupResponse{
		Address:       first.Address,
		IsEVM:         true,
		Blockchains:   blockchains,
		Currencies:    currencies,
		SubmittedAt:   submittedAt,
		CurrencyCount: len(entries),
	}

	if first.Name.Valid {
		response.Name = &first.Name.String
	}
	if first.Tag.Valid {
		response.Tag = &first.Tag.String
	}

	return response
}

func (s *Service) UpdateAddress(ctx context.Context, userID, addressBookEntryID uuid.UUID, params UpdateAddressDTO) (*models.UserAddressBook, error) {
	// Check if entry exists and belongs to user
	existing, err := s.storage.UserAddressBook().GetByID(ctx, addressBookEntryID)
	if err != nil {
		return nil, fmt.Errorf("address book entry not found")
	}
	if existing.UserID != userID {
		return nil, fmt.Errorf("user id mismatch for address book entry")
	}

	updateParams := repo_user_address_book.UpdateParams{
		ID: addressBookEntryID,
	}

	if params.Name != nil {
		updateParams.Name = pgtype.Text{String: *params.Name, Valid: true}
	}

	if params.Tag != nil {
		updateParams.Tag = pgtype.Text{String: *params.Tag, Valid: true}
	}

	address, err := s.storage.UserAddressBook().Update(ctx, updateParams)
	if err != nil {
		return nil, fmt.Errorf("failed to update address: %w", err)
	}

	return address, nil
}

func (s *Service) deleteSimpleAddress(ctx context.Context, usr *models.User, id uuid.UUID, deleteWithdrawalRule bool) error {
	err := repos.BeginTxFunc(ctx, s.storage.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		// Get the address entry before deleting it
		addressEntry, err := s.storage.UserAddressBook().GetByID(ctx, id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil
			}
			return fmt.Errorf("failed to get address entry: %w", err)
		}

		// Delete from address book
		err = s.storage.UserAddressBook(repos.WithTx(tx)).SoftDelete(ctx, id)
		if err != nil {
			return fmt.Errorf("failed to delete address: %w", err)
		}

		// Conditionally clean up withdrawal rule
		if deleteWithdrawalRule {
			if err := s.cleanupWithdrawalRule(ctx, usr, addressEntry, tx); err != nil {
				return fmt.Errorf("failed to cleanup withdrawal rule: %w", err)
			}
		}

		return nil
	})

	return err
}

// AddWithdrawalRule handles all types of withdrawal rule additions based on DTO
func (s *Service) AddWithdrawalRule(ctx context.Context, dto AddWithdrawalRuleDTO) error {
	// Fetch user for processing whitelist updates
	user, err := s.storage.Users().GetByID(ctx, dto.UserID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if dto.IsEVM {
		return s.addWithdrawalRulesForEVMAddress(ctx, dto.UserID, *dto.Address, user, dto.TOTP)
	}
	if dto.IsUniversal {
		return s.addWithdrawalRulesForUniversalAddress(ctx, dto.UserID, *dto.Address, *dto.Blockchain, user, dto.TOTP)
	}
	// Individual address withdrawal rule
	return s.addWithdrawalRuleForSimpleAddress(ctx, dto.UserID, *dto.ID, user, dto.TOTP)
}

// DeleteAddress handles all types of address deletions based on DTO
func (s *Service) DeleteAddress(ctx context.Context, dto DeleteAddressDTO, usr *models.User) error {
	if dto.IsEVM {
		return s.deleteEVMAddress(ctx, usr, *dto.Address, dto.DeleteWithdrawalRule)
	}
	if dto.IsUniversal {
		return s.deleteUniversalAddress(ctx, usr, *dto.Address, *dto.Blockchain, dto.DeleteWithdrawalRule)
	}
	// Default to simple address deletion by ID
	return s.deleteSimpleAddress(ctx, usr, *dto.ID, dto.DeleteWithdrawalRule)
}
