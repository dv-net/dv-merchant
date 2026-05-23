package transactions

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/dv-net/dv-merchant/internal/event"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/notify"
	"github.com/dv-net/dv-merchant/internal/util"
	"github.com/shopspring/decimal"
)

func (s *Service) handleDepositReceiptSent(ev event.IEvent) error {
	convertedEv, ok := ev.(TransactionEvent)
	if !ok {
		return fmt.Errorf("invalid event type %s", ev.Type())
	}

	storeData := convertedEv.GetStore()
	tx := convertedEv.GetTx()
	currency := convertedEv.GetCurrency()
	userEmail := convertedEv.GetWalletEmail()
	usdFee := convertedEv.GetUsdFee()

	if userEmail == "" {
		s.log.Info("Email not sent: no user email provided")
		return nil
	}

	if tx.GetAmountUsd().LessThan(storeData.MinimalPayment) {
		s.log.Infof("Email not sent: txAmount=%s < minimalPayment=%s",
			tx.GetAmountUsd().String(), storeData.MinimalPayment.String(),
		)
		return nil
	}

	if s.notificationService == nil {
		s.log.Info("No notification service disabled")
		return nil
	}

	go s.notificationService.SendSystemEmail(context.Background(), models.NotificationTypeUserCryptoReceipt, userEmail, &notify.UserCryptoReceiptNotificationData{
		Email:                userEmail,
		Language:             convertedEv.GetWalletLocale(),
		PaymentStatus:        notify.PaymentStatusCompleted.String(),
		PaymentType:          notify.PaymentTypeDeposit.String(),
		ReceiptId:            tx.GetReceiptID().UUID.String(),
		PaymentDate:          tx.GetCreatedAt().Time.Format(time.DateTime),
		UsdAmount:            tx.GetAmountUsd().RoundDown(2).String(),
		TokenAmount:          tx.GetAmount().String(),
		TokenSymbol:          currency.Code,
		TransactionHash:      tx.GetTxHash(),
		BlockchainCurrencyID: tx.GetCurrencyID(),
		BlockchainName:       strings.ToUpper(currency.Blockchain.String()),
		ExchangeRate:         convertedEv.GetExchangeRate().String(),
		NetworkFeeAmount:     tx.GetFee().String(),
		NetworkFeeUSD:        usdFee.String(),
		NetworkFeeCurrency:   currency.Code,
		PlatformFeeAmount:    decimal.Zero.String(),
		PlatformFeeUSD:       decimal.Zero.String(),
		PlatformFeeCurrency:  currency.Code,
		StoreName:            storeData.Name,
		StoreUserID:          convertedEv.GetStoreExternalID(),
	}, &models.NotificationArgs{
		StoreID: util.Pointer(storeData.ID),
	})

	return nil
}
