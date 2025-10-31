package wallet

import (
	"context"
	"errors"
	"fmt"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_wallet_addresses_activity_logs"
	"github.com/dv-net/dv-merchant/pkg/pgtypeutils"

	"github.com/goccy/go-json"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *Service) logProcessingAddressReceived(ctx context.Context, walletAddress *models.WalletAddress, ip *string) error {
	textVariables := struct {
		Address string  `json:"address"`
		IP      *string `json:"ip"`
	}{
		Address: walletAddress.Address,
		IP:      ip,
	}
	textVariablesJSON, err := json.Marshal(textVariables)
	if err != nil {
		return fmt.Errorf("failed to marshal text variables: %w", err)
	}

	_, err = s.storage.WalletAddressesActivityLog().Create(ctx, repo_wallet_addresses_activity_logs.CreateParams{
		WalletAddressesID: walletAddress.ID,
		Text:              "Address received from processing {address} from user ip {ip}",
		TextVariables:     textVariablesJSON,
	})
	if err != nil {
		return fmt.Errorf("failed to create log: %w", err)
	}

	return nil
}

func (s *Service) logProcessingLoadAddressPrivateKey(ctx context.Context, address string, walletID uuid.UUID, ip string) error {
	textVariables := struct {
		Address   string `json:"address"`
		UserEmail string `json:"user_email"`
		IP        string `json:"ip"`
	}{
		Address: address,
		IP:      ip,
	}

	textVariablesJSON, err := json.Marshal(textVariables)
	if err != nil {
		return fmt.Errorf("failed to marshal text variables: %w", err)
	}

	_, err = s.storage.WalletAddressesActivityLog().Create(ctx, repo_wallet_addresses_activity_logs.CreateParams{
		WalletAddressesID: walletID,
		Text:              "Load Private key {address} user {user_email} from ip {ip}",
		TextVariables:     textVariablesJSON,
	})
	if err != nil {
		return fmt.Errorf("failed to create log: %w", err)
	}

	return nil
}

func (s *Service) logWalletStatusChanged(ctx context.Context, walletAddress *models.WalletAddress, oldStatus, newStatus string) error {
	textVariables := struct {
		OldStatus string `json:"old_status"`
		NewStatus string `json:"new_status"`
	}{
		OldStatus: oldStatus,
		NewStatus: newStatus,
	}

	textVariablesJSON, err := json.Marshal(textVariables)
	if err != nil {
		return fmt.Errorf("failed to marshal text variables: %w", err)
	}

	_, err = s.storage.WalletAddressesActivityLog().Create(ctx, repo_wallet_addresses_activity_logs.CreateParams{
		WalletAddressesID: walletAddress.ID,
		Text:              "Wallet status changed from {old_status} to {new_status}",
		TextVariables:     textVariablesJSON,
	})
	if err != nil {
		return fmt.Errorf("failed to create log: %w", err)
	}

	return nil
}

func (s *Service) GetWalletLogs(ctx context.Context, walletAddressID uuid.UUID) ([]*AddressLog, error) {
	walletLogs, err := s.storage.WalletAddressesActivityLog().GetLogByWalletAddressID(ctx, walletAddressID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []*AddressLog{}, nil
		}
		return nil, err
	}

	if len(walletLogs) == 0 {
		return []*AddressLog{}, nil
	}

	logs := make([]*AddressLog, len(walletLogs))
	for i, walletLog := range walletLogs {
		var textVariables map[string]interface{}
		if err := json.Unmarshal(walletLog.TextVariables, &textVariables); err != nil {
			return nil, fmt.Errorf("failed to unmarshal text variables: %w", err)
		}

		logs[i] = &AddressLog{
			Text:          walletLog.Text,
			TextVariables: textVariables,
			CreatedAt:     pgtypeutils.DecodeTime(walletLog.CreatedAt),
			UpdatedAt:     pgtypeutils.DecodeTime(walletLog.UpdatedAt),
		}
	}

	return logs, nil
}
