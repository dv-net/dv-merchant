package converters

import (
	"github.com/dv-net/dv-merchant/internal/delivery/http/request/withdrawal_wallets_request"
	"github.com/dv-net/dv-merchant/internal/service/withdrawal_wallet"

	"github.com/google/uuid"
)

func NewWithdrawalWalletAddressesDtoFromRequest(
	request withdrawal_wallets_request.UpdateAddressesListRequest,
	walletID uuid.UUID,
) withdrawal_wallet.UpdateAddressesListDTO {
	res := withdrawal_wallet.UpdateAddressesListDTO{
		WalletID:  walletID,
		TOTP:      request.TOTP,
		Addresses: NewAddressesListDtoFromRequestsList(request.Addresses),
	}

	return res
}

func NewAddressesListDtoFromRequestsList(addresses []withdrawal_wallets_request.WalletAddress) []withdrawal_wallet.AddressDTO {
	res := make([]withdrawal_wallet.AddressDTO, 0, len(addresses))
	for _, val := range addresses {
		res = append(res, withdrawal_wallet.AddressDTO{
			Name:    val.Name,
			Address: val.Address,
		})
	}

	return res
}
