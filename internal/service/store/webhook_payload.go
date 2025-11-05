package store

import (
	"context"
	"fmt"
	"time"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/mx/util"
	"github.com/goccy/go-json"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
)

func (s *Service) prepareDepositHookPayload(
	tx models.ITransaction,
	curr models.Currency,
	whType models.WebhookEvent,
	storeExternalID string,
) ([]byte, error) {
	var prefix string
	if !tx.IsConfirmed() {
		prefix = "unconfirmed_"
	}
	payload := map[string]any{
		prefix + "type":       whType,
		prefix + "status":     models.TransactionStatusCompleted,
		prefix + "created_at": tx.GetCreatedAt(),
		prefix + "paid_at":    tx.GetNetworkCreatedAt(),
		prefix + "amount":     tx.GetAmountUsd().String(),
		prefix + "transactions": map[string]any{
			prefix + "tx_id":       tx.GetID().String(),
			prefix + "tx_hash":     tx.GetTxHash(),
			prefix + "bc_uniq_key": tx.GetBcUniqKey(),
			prefix + "created_at":  tx.GetCreatedAt(),
			prefix + "currency":    curr.Code,
			prefix + "currency_id": curr.ID,
			prefix + "blockchain":  curr.Blockchain.String(),
			prefix + "amount":      tx.GetAmount().String(),
			prefix + "amount_usd":  tx.GetAmountUsd().String(),
		},
		prefix + "wallet": map[string]any{
			prefix + "id":                tx.GetAccountID(),
			prefix + "store_external_id": storeExternalID,
		},
	}

	preparedPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("prepare wh body: %w", err)
	}

	return preparedPayload, nil
}

func (s *Service) prepareMockTransactionDataForWhTest(whType models.WebhookEvent, wh models.StoreWebhook, userID uuid.UUID) (models.ITransaction, error) {
	walletID, _ := uuid.NewRandom()
	preparedAccountID := uuid.NullUUID{
		UUID:  walletID,
		Valid: true,
	}
	preparedStoreID := uuid.NullUUID{
		UUID:  wh.StoreID,
		Valid: true,
	}
	txID, _ := uuid.NewRandom()
	pgTimeStamp := pgtype.Timestamp{Time: time.Now(), Valid: true}
	curr, _ := models.BlockchainBitcoin.NativeCurrency()

	var txData models.ITransaction
	switch whType {
	case models.WebhookEventPaymentReceived, models.WebhookEventWithdrawalFromProcessingReceived:
		receiptID, _ := uuid.NewRandom()
		txData = models.Transaction{
			ID:                 txID,
			UserID:             userID,
			StoreID:            preparedStoreID,
			ReceiptID:          uuid.NullUUID{UUID: receiptID, Valid: true},
			AccountID:          preparedAccountID,
			CurrencyID:         curr,
			Blockchain:         models.BlockchainBitcoin.String(),
			TxHash:             "tx_hash_example",
			BcUniqKey:          util.Pointer("bc_uniq_key_example"),
			Type:               models.TransactionsTypeDeposit,
			FromAddress:        "15muvlleOFc9nh10zTJSoM08Fil96tXBfn",
			ToAddress:          "1pmlFcSaUPBhJYeuG7ahQvTWQGWJff0IW1",
			Amount:             decimal.New(100, 0),
			AmountUsd:          decimal.NullDecimal{Decimal: decimal.New(100, 0), Valid: true},
			Fee:                decimal.Decimal{},
			WithdrawalIsManual: false,
			NetworkCreatedAt:   pgTimeStamp,
			CreatedAt:          pgTimeStamp,
			UpdatedAt:          pgTimeStamp,
			CreatedAtIndex:     pgtype.Int8{Int64: 1, Valid: true},
		}
	case models.WebhookEventPaymentNotConfirmed:
		txData = models.UnconfirmedTransaction{
			ID:               txID,
			UserID:           userID,
			StoreID:          preparedStoreID,
			AccountID:        preparedAccountID,
			CurrencyID:       curr,
			TxHash:           "tx_hash_example",
			BcUniqKey:        util.Pointer("bc_uniq_key_example"),
			Type:             models.TransactionsTypeDeposit,
			FromAddress:      "15muvlleOFc9nh10zTJSoM08Fil96tXBfn",
			ToAddress:        "1pmlFcSaUPBhJYeuG7ahQvTWQGWJff0IW1",
			Amount:           decimal.New(100, 10),
			AmountUsd:        decimal.NullDecimal{Decimal: decimal.New(100, 10), Valid: true},
			NetworkCreatedAt: pgTimeStamp,
			CreatedAt:        pgTimeStamp,
			UpdatedAt:        pgTimeStamp,
			Blockchain:       models.BlockchainBitcoin.String(),
		}
	default:
		return nil, fmt.Errorf("undefined wh_type: %s", whType.String())
	}

	return txData, nil
}

func (s *Service) prepareInvoiceChangeStatusPayload(
	invoice *models.Invoice,
	transactions []*models.Transaction,
	webhookEvent models.WebhookEvent,
) ([]byte, error) {
	transactionsData := make([]map[string]any, 0, len(transactions))
	for _, tx := range transactions {
		currency, err := s.storage.Currencies().GetByID(context.Background(), tx.CurrencyID)
		if err != nil {
			s.log.Errorw("failed to get currency for transaction",
				"error", err,
				"tx_id", tx.ID,
				"currency_id", tx.CurrencyID,
			)
			continue
		}

		txData := map[string]any{
			"tx_id":       tx.ID.String(),
			"tx_hash":     tx.TxHash,
			"bc_uniq_key": tx.BcUniqKey,
			"created_at":  tx.CreatedAt,
			"currency":    currency.Code,
			"currency_id": tx.CurrencyID,
			"blockchain":  tx.Blockchain,
			"amount":      tx.GetAmount().String(),
			"amount_usd":  tx.GetAmountUsd().String(),
		}

		if tx.NetworkCreatedAt.Valid {
			txData["network_created_at"] = tx.NetworkCreatedAt.Time
		}

		transactionsData = append(transactionsData, txData)
	}

	payload := map[string]any{
		"type":                webhookEvent,
		"invoice_id":          invoice.ID.String(),
		"invoice_status":      invoice.Status,
		"order_id":            invoice.OrderID,
		"expected_amount_usd": invoice.ExpectedAmountUsd.String(),
		"received_amount_usd": invoice.ReceivedAmountUsd.String(),
		"created_at":          invoice.CreatedAt,
		"transactions":        transactionsData,
	}

	if invoice.ExpiresAt.Valid {
		payload["expires_at"] = invoice.ExpiresAt.Time
	}

	preparedPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal invoice webhook payload: %w", err)
	}

	return preparedPayload, nil
}

func (s *Service) prepareWithdrawalFromProcessingPayload(
	tx *models.Transaction,
	currency *models.Currency,
	withdrawalID string,
	webhookEvent models.WebhookEvent,
) ([]byte, error) {
	payload := map[string]any{
		"type":          webhookEvent,
		"created_at":    tx.CreatedAt,
		"paid_at":       tx.NetworkCreatedAt,
		"amount":        tx.GetAmountUsd().String(),
		"withdrawal_id": withdrawalID,
		"transactions": map[string]any{
			"tx_id":       tx.ID.String(),
			"tx_hash":     tx.TxHash,
			"bc_uniq_key": tx.BcUniqKey,
			"created_at":  tx.CreatedAt,
			"currency":    currency.Code,
			"currency_id": currency.ID,
			"blockchain":  currency.Blockchain.String(),
			"amount":      tx.GetAmount().String(),
			"amount_usd":  tx.GetAmountUsd().String(),
		},
	}

	preparedPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal withdrawal webhook payload: %w", err)
	}

	return preparedPayload, nil
}
