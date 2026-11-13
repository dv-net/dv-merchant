package transactions

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/dv-net/dv-merchant/internal/event"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/notify"
	"github.com/dv-net/dv-merchant/pkg/aml/providers"
	"github.com/shopspring/decimal"
)

func (s *Service) handleDepositReceiptSent(ev event.IEvent) error {
	convertedEv, ok := ev.(DepositReceiptSentEvent)
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

	paymentStatus := convertedEv.PaymentStatus
	if paymentStatus == "" {
		paymentStatus = notify.PaymentStatusCompleted
	}

	go s.notificationService.SendSystemEmail(context.Background(), models.NotificationTypeUserCryptoReceipt, userEmail, &notify.UserCryptoReceiptNotificationData{
		Email:                userEmail,
		Language:             convertedEv.GetWalletLocale(),
		PaymentStatus:        paymentStatus.String(),
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
	}, &models.NotificationArgs{
		StoreID: new(storeData.ID),
	})

	return nil
}

func (s *Service) handleDepositAMLBlocked(ev event.IEvent) error {
	convertedEv, ok := ev.(DepositAMLBlockedEvent)
	if !ok {
		return fmt.Errorf("invalid event type %s", ev.Type())
	}

	if convertedEv.WalletEmail == "" {
		s.log.Info("AML blocked email not sent: no user email provided")
		return nil
	}

	if s.notificationService == nil {
		s.log.Info("No notification service disabled")
		return nil
	}

	tx := convertedEv.Tx
	blocked := convertedEv.Blocked

	reasons := make([]notify.DepositAMLBlockReason, 0, len(blocked.BlockReasons))
	for _, r := range blocked.BlockReasons {
		reason := notify.DepositAMLBlockReason{
			RiskType: r.RiskType,
			Value:    r.Value.String(),
		}
		if label, ok := providers.SignalCategoryLabel(r.RiskType); ok {
			reason.SignalLabel = label
		}
		reasons = append(reasons, reason)
	}

	go s.notificationService.SendSystemEmail(context.Background(), models.NotificationTypeDepositAMLBlocked, convertedEv.WalletEmail, &notify.DepositAMLBlockedNotificationData{
		Email:                convertedEv.WalletEmail,
		Language:             convertedEv.WalletLocale,
		PaymentDate:          tx.CreatedAt.Time.Format(time.DateTime),
		TokenAmount:          tx.Amount.String(),
		TokenSymbol:          convertedEv.Currency.Code,
		UsdAmount:            tx.GetAmountUsd().RoundDown(2).String(),
		BlockchainName:       strings.ToUpper(convertedEv.Currency.Blockchain.String()),
		BlockchainCurrencyID: tx.CurrencyID,
		TransactionHash:      tx.TxHash,
		ToAddress:            tx.ToAddress,
		RiskLevel:            blocked.RiskLevel,
		RiskScore:            blocked.Score.String(),
		BlockReasons:         reasons,
		BlockedTransactionID: blocked.ID.String(),
	}, &models.NotificationArgs{
		StoreID: new(convertedEv.StoreID),
	})

	return nil
}
