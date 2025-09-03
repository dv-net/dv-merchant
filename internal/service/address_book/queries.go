package address_book

import (
	"context"
	"fmt"

	"github.com/dv-net/dv-merchant/internal/delivery/http/responses/withdrawal_response"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_user_address_book"

	"github.com/google/uuid"
)

func (s *Service) GetUserAddresses(ctx context.Context, userID uuid.UUID) (*withdrawal_response.AddressBookListResponse, error) {
	addresses, err := s.storage.UserAddressBook().GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user addresses: %w", err)
	}

	// Group entries by type
	simpleEntries := make([]*models.UserAddressBook, 0)
	universalGroups := make(map[string][]*models.UserAddressBook)
	evmGroups := make(map[string][]*models.UserAddressBook)

	for _, entry := range addresses {
		switch entry.Type {
		case models.AddressBookTypeSimple:
			simpleEntries = append(simpleEntries, entry)
		case models.AddressBookTypeUniversal:
			universalGroups[entry.Address] = append(universalGroups[entry.Address], entry)
		case models.AddressBookTypeEVM:
			evmGroups[entry.Address] = append(evmGroups[entry.Address], entry)
		}
	}

	// Convert simple addresses with withdrawal rule status
	simpleAddressResponses := make([]*withdrawal_response.AddressBookEntryResponse, len(simpleEntries))
	for i, entry := range simpleEntries {
		resp := &withdrawal_response.AddressBookEntryResponse{
			ID:         entry.ID,
			Address:    entry.Address,
			CurrencyID: entry.CurrencyID,
			Blockchain: *entry.Blockchain,
		}

		if entry.Name.Valid {
			resp.Name = &entry.Name.String
		}
		if entry.Tag.Valid {
			resp.Tag = &entry.Tag.String
		}
		if entry.SubmittedAt.Valid {
			resp.SubmittedAt = entry.SubmittedAt.Time.Format("2006-01-02T15:04:05Z")
		}

		// Check withdrawal rule status
		if withdrawalRuleExists, err := s.CheckWithdrawalRuleExists(ctx, entry); err == nil {
			resp.WithdrawalRuleExists = withdrawalRuleExists
		}

		simpleAddressResponses[i] = resp
	}

	// Convert universal groups
	universalGroupResponses := make([]*withdrawal_response.UniversalAddressGroupResponse, 0, len(universalGroups))
	for _, group := range universalGroups {
		if len(group) > 0 {
			resp := s.toUniversalAddressGroupResponse(ctx, group)
			if resp != nil {
				universalGroupResponses = append(universalGroupResponses, resp)
			}
		}
	}

	// Convert EVM groups
	evmGroupResponses := make([]*withdrawal_response.EVMAddressGroupResponse, 0, len(evmGroups))
	for _, group := range evmGroups {
		if len(group) > 0 {
			resp := s.toEVMAddressGroupResponse(ctx, group)
			if resp != nil {
				evmGroupResponses = append(evmGroupResponses, resp)
			}
		}
	}

	return &withdrawal_response.AddressBookListResponse{
		Addresses:       simpleAddressResponses,
		UniversalGroups: universalGroupResponses,
		EVMGroups:       evmGroupResponses,
	}, nil
}

func (s *Service) GetAddressByID(ctx context.Context, id uuid.UUID) (*models.UserAddressBook, error) {
	address, err := s.storage.UserAddressBook().GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get address by ID: %w", err)
	}

	return address, nil
}

func (s *Service) GetUserAddressesByCurrency(ctx context.Context, userID uuid.UUID, currencyID string) ([]*models.UserAddressBook, error) {
	addresses, err := s.storage.UserAddressBook().GetByUserAndCurrency(ctx, repo_user_address_book.GetByUserAndCurrencyParams{
		UserID:     userID,
		CurrencyID: currencyID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get addresses by currency: %w", err)
	}

	return addresses, nil
}

func (s *Service) GetUserAddressesByBlockchain(ctx context.Context, userID uuid.UUID, blockchain models.Blockchain) ([]*models.UserAddressBook, error) {
	addresses, err := s.storage.UserAddressBook().GetByUserAndBlockchain(ctx, repo_user_address_book.GetByUserAndBlockchainParams{
		UserID:     userID,
		Blockchain: &blockchain,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get addresses by blockchain: %w", err)
	}

	return addresses, nil
}

func (s *Service) CheckAddressExists(ctx context.Context, userID uuid.UUID, address string, currencyID string) (bool, error) {
	exists, err := s.storage.UserAddressBook().CheckExists(ctx, repo_user_address_book.CheckExistsParams{
		UserID:     userID,
		Address:    address,
		CurrencyID: currencyID,
	})
	if err != nil {
		return false, fmt.Errorf("failed to check if address exists: %w", err)
	}

	return exists, nil
}

func (s *Service) IsUniversalAddress(ctx context.Context, userID uuid.UUID, address string, blockchain models.Blockchain) (bool, int, error) {
	entries, err := s.storage.UserAddressBook().GetByUserAddressAndBlockchain(ctx, repo_user_address_book.GetByUserAddressAndBlockchainParams{
		UserID:     userID,
		Address:    address,
		Blockchain: &blockchain,
	})
	if err != nil {
		return false, 0, fmt.Errorf("failed to get address entries: %w", err)
	}

	count := len(entries)
	isUniversal := count > 1

	return isUniversal, count, nil
}

func (s *Service) GetUniversalAddressGroup(ctx context.Context, userID uuid.UUID, address string, blockchain models.Blockchain) ([]*models.UserAddressBook, error) {
	entries, err := s.storage.UserAddressBook().GetByUserAddressAndBlockchain(ctx, repo_user_address_book.GetByUserAddressAndBlockchainParams{
		UserID:     userID,
		Address:    address,
		Blockchain: &blockchain,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get universal address group: %w", err)
	}

	return entries, nil
}
