package exchange

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_user_exchanges"
	"github.com/dv-net/dv-merchant/pkg/exchange_client/utils"
	"github.com/dv-net/dv-merchant/pkg/netiface"
)

type NetworkSettingsDTO struct {
	ProxyURL           *string
	NetworkInterface   *string
	InterfaceAvailable bool
}

func (s *Service) GetNetworkSettings(ctx context.Context, userID uuid.UUID, slug models.ExchangeSlug) (*NetworkSettingsDTO, error) {
	ue, err := s.getUserExchange(ctx, userID, slug)
	if err != nil {
		return nil, err
	}
	return newNetworkSettingsDTO(ue), nil
}

func (s *Service) SetNetworkSettings(ctx context.Context, userID uuid.UUID, slug models.ExchangeSlug, proxyURL, networkInterface *string) (*NetworkSettingsDTO, error) {
	ue, err := s.getUserExchange(ctx, userID, slug)
	if err != nil {
		return nil, err
	}

	netCfg := utils.NetConfig{}

	if raw := strings.TrimSpace(ptrValue(proxyURL)); raw != "" {
		u, err := utils.ParseProxyURL(raw)
		if err != nil {
			return nil, err
		}
		restoreMaskedPassword(u, ue.ProxyUrl)
		netCfg.ProxyURL = u.String()
	}

	if name := strings.TrimSpace(ptrValue(networkInterface)); name != "" {
		if !netiface.Exists(name) {
			return nil, ErrNetworkInterfaceNotFound
		}
		netCfg.Interface = name
	}

	if netCfg != (utils.NetConfig{}) {
		if err := s.checkNetworkSettings(ctx, userID, slug, netCfg); err != nil {
			return nil, err
		}
	}

	params := repo_user_exchanges.UpdateNetworkSettingsParams{
		UserID:           userID,
		ExchangeID:       ue.ExchangeID,
		ProxyUrl:         pgtype.Text{String: netCfg.ProxyURL, Valid: netCfg.ProxyURL != ""},
		NetworkInterface: pgtype.Text{String: netCfg.Interface, Valid: netCfg.Interface != ""},
	}
	if err := s.st.UserExchanges().UpdateNetworkSettings(ctx, params); err != nil {
		return nil, fmt.Errorf("update exchange network settings: %w", err)
	}

	ue.ProxyUrl = params.ProxyUrl
	ue.NetworkInterface = params.NetworkInterface
	return newNetworkSettingsDTO(ue), nil
}

func (s *Service) checkNetworkSettings(ctx context.Context, userID uuid.UUID, slug models.ExchangeSlug, netCfg utils.NetConfig) error {
	drv, err := s.exManager.GetDriverWithNetConfig(ctx, slug, userID, netCfg)
	if err != nil {
		return fmt.Errorf("create exchange client: %w", err)
	}
	if err := drv.TestConnection(ctx); err != nil {
		return fmt.Errorf("%w: %w", ErrNetworkSettingsCheck, err)
	}
	return nil
}

func (s *Service) getUserExchange(ctx context.Context, userID uuid.UUID, slug models.ExchangeSlug) (*models.UserExchange, error) {
	ex, err := s.st.Exchanges().GetExchangeBySlug(ctx, slug)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrExchangeNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get exchange: %w", err)
	}

	ue, err := s.st.UserExchanges().GetByUserAndExchangeID(ctx, repo_user_exchanges.GetByUserAndExchangeIDParams{
		UserID:     userID,
		ExchangeID: ex.ID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrExchangeKeysNotSet
	}
	if err != nil {
		return nil, fmt.Errorf("get user exchange: %w", err)
	}
	return ue, nil
}

func newNetworkSettingsDTO(ue *models.UserExchange) *NetworkSettingsDTO {
	dto := &NetworkSettingsDTO{InterfaceAvailable: true}
	if ue.ProxyUrl.Valid {
		masked := redactProxyURL(ue.ProxyUrl.String)
		dto.ProxyURL = &masked
	}
	if ue.NetworkInterface.Valid {
		name := ue.NetworkInterface.String
		dto.NetworkInterface = &name
		dto.InterfaceAvailable = netiface.Exists(name)
	}
	return dto
}

const maskedPassword = "xxxxx"

func restoreMaskedPassword(u *url.URL, saved pgtype.Text) {
	if u.User == nil || !saved.Valid {
		return
	}
	if p, ok := u.User.Password(); !ok || p != maskedPassword {
		return
	}
	savedURL, err := utils.ParseProxyURL(saved.String)
	if err != nil || savedURL.User == nil || savedURL.User.Username() != u.User.Username() {
		return
	}
	if savedPassword, ok := savedURL.User.Password(); ok {
		u.User = url.UserPassword(u.User.Username(), savedPassword)
	}
}

func redactProxyURL(raw string) string {
	u, err := utils.ParseProxyURL(raw)
	if err != nil {
		return ""
	}
	return u.Redacted()
}

func ptrValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
