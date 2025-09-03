package converters

import (
	"sort"

	"github.com/dv-net/dv-merchant/internal/delivery/http/request/withdrawal_requests"
	"github.com/dv-net/dv-merchant/internal/delivery/http/responses/withdrawal_response"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/address_book"
	"github.com/google/uuid"
)

func FromCreateAddressBookRequest(req withdrawal_requests.CreateAddressBookRequest, userID uuid.UUID) address_book.CreateAddressDTO {
	createWithdrawalRule := false
	if req.CreateWithdrawalRule != nil {
		createWithdrawalRule = *req.CreateWithdrawalRule
	}

	dto := address_book.CreateAddressDTO{
		UserID:               userID,
		Address:              req.Address,
		CurrencyID:           req.CurrencyID,
		Universal:            req.IsUniversal,
		EVM:                  req.IsEVM,
		Name:                 req.Name,
		Tag:                  req.Tag,
		Blockchain:           req.Blockchain,
		CreateWithdrawalRule: createWithdrawalRule,
		TOTP:                 req.TOTP,
	}

	return dto
}

func FromUpdateAddressBookRequest(req withdrawal_requests.UpdateAddressBookRequest) address_book.UpdateAddressDTO {
	return address_book.UpdateAddressDTO{
		Name: req.Name,
		Tag:  req.Tag,
	}
}

func FromDeleteAddressBookRequest(req withdrawal_requests.DeleteAddressBookRequest, userID uuid.UUID) address_book.DeleteAddressDTO {
	deleteWithdrawalRule := true
	if req.DeleteWithdrawalRule != nil {
		deleteWithdrawalRule = *req.DeleteWithdrawalRule
	}

	dto := address_book.DeleteAddressDTO{
		UserID:               userID,
		DeleteWithdrawalRule: deleteWithdrawalRule,
		IsEVM:                req.IsEVM,
		IsUniversal:          req.IsUniversal,
		Address:              req.Address,
		Blockchain:           req.Blockchain,
	}

	if req.ID != nil {
		if id, err := uuid.Parse(*req.ID); err == nil {
			dto.ID = &id
		}
	}

	return dto
}

func FromAddWithdrawalRuleRequest(req withdrawal_requests.AddWithdrawalRuleRequest, userID uuid.UUID) address_book.AddWithdrawalRuleDTO {
	dto := address_book.AddWithdrawalRuleDTO{
		UserID:      userID,
		IsEVM:       req.IsEVM,
		IsUniversal: req.IsUniversal,
		Address:     req.Address,
		Blockchain:  req.Blockchain,
		TOTP:        req.TOTP,
	}

	if req.ID != nil {
		if id, err := uuid.Parse(*req.ID); err == nil {
			dto.ID = &id
		}
	}

	return dto
}

func ToAddressBookEntryResponse(entry *models.UserAddressBook) *withdrawal_response.AddressBookEntryResponse {
	resp := &withdrawal_response.AddressBookEntryResponse{
		ID:         entry.ID,
		Address:    entry.Address,
		CurrencyID: entry.CurrencyID,
		Blockchain: *entry.Blockchain, // Safe since blockchain is now always populated
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

	return resp
}

func ToAddressBookShortEntryResponse(entry *models.UserAddressBook) *withdrawal_response.AddressBookEntryResponseShort {
	resp := &withdrawal_response.AddressBookEntryResponseShort{
		ID:         entry.ID,
		CurrencyID: entry.CurrencyID,
	}

	return resp
}

func ToAddressBookListResponse(entries []*models.UserAddressBook) *withdrawal_response.AddressBookListResponse {
	// Group entries by type
	simpleEntries := make([]*models.UserAddressBook, 0)
	universalGroups := make(map[string][]*models.UserAddressBook) // key: address
	evmGroups := make(map[string][]*models.UserAddressBook)       // key: address

	for _, entry := range entries {
		switch entry.Type {
		case models.AddressBookTypeSimple:
			simpleEntries = append(simpleEntries, entry)
		case models.AddressBookTypeUniversal:
			universalGroups[entry.Address] = append(universalGroups[entry.Address], entry)
		case models.AddressBookTypeEVM:
			evmGroups[entry.Address] = append(evmGroups[entry.Address], entry)
		}
	}

	// Convert simple addresses
	addresses := make([]*withdrawal_response.AddressBookEntryResponse, len(simpleEntries))
	for i, entry := range simpleEntries {
		addresses[i] = ToAddressBookEntryResponse(entry)
	}

	// Convert universal groups
	finalUniversalGroups := make([]*withdrawal_response.UniversalAddressGroupResponse, 0, len(universalGroups))
	for _, group := range universalGroups {
		universalGroup := ToUniversalAddressGroupResponse(group)
		if universalGroup != nil {
			finalUniversalGroups = append(finalUniversalGroups, universalGroup)
		}
	}

	// Convert EVM groups
	finalEVMGroups := make([]*withdrawal_response.EVMAddressGroupResponse, 0, len(evmGroups))
	for _, group := range evmGroups {
		evmGroup := ToEVMAddressGroupResponse(group)
		if evmGroup != nil {
			finalEVMGroups = append(finalEVMGroups, evmGroup)
		}
	}

	// Sort all collections by address for consistent ordering
	sort.Slice(addresses, func(i, j int) bool {
		return addresses[i].Address < addresses[j].Address
	})

	sort.Slice(finalUniversalGroups, func(i, j int) bool {
		return finalUniversalGroups[i].Address < finalUniversalGroups[j].Address
	})

	sort.Slice(finalEVMGroups, func(i, j int) bool {
		return finalEVMGroups[i].Address < finalEVMGroups[j].Address
	})

	return &withdrawal_response.AddressBookListResponse{
		Addresses:       addresses,
		UniversalGroups: finalUniversalGroups,
		EVMGroups:       finalEVMGroups,
	}
}

func ToUniversalAddressGroupResponse(entries []*models.UserAddressBook) *withdrawal_response.UniversalAddressGroupResponse {
	if len(entries) == 0 {
		return nil
	}

	// Use first entry as template for common fields
	first := entries[0]

	// Convert all entries to currency responses
	currencies := make([]*withdrawal_response.AddressBookEntryResponseShort, len(entries))
	for i, entry := range entries {
		currencies[i] = ToAddressBookShortEntryResponse(entry)
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

func ToEVMAddressGroupResponse(entries []*models.UserAddressBook) *withdrawal_response.EVMAddressGroupResponse {
	if len(entries) == 0 {
		return nil
	}

	// Use first entry as template for common fields
	first := entries[0]

	// Convert all entries to currency responses
	currencies := make([]*withdrawal_response.AddressBookEntryResponseShort, len(entries))
	for i, entry := range entries {
		currencies[i] = ToAddressBookShortEntryResponse(entry)
	}

	// Collect unique blockchains
	blockchainSet := make(map[models.Blockchain]bool)
	for _, entry := range entries {
		blockchainSet[*entry.Blockchain] = true
	}

	// Convert set to slice
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
