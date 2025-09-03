package system

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/dv-net/dv-merchant/internal/config"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/analytics"
	"github.com/dv-net/dv-merchant/internal/service/permission"
	"github.com/dv-net/dv-merchant/internal/service/setting"
	"github.com/dv-net/dv-merchant/pkg/admin_gateway"
	admin_requests "github.com/dv-net/dv-merchant/pkg/admin_gateway/requests"
	"github.com/dv-net/dv-merchant/pkg/logger"
	"github.com/jackc/pgx/v5"
)

type ISystemService interface {
	GetInfo(ctx context.Context) (*models.SystemInfo, error)
	GetAppVersion(ctx context.Context) (string, string)
	GetIP(ctx context.Context) (string, error)
	GetProfile(ctx context.Context) models.AppProfile
	RunHeartbeatLoop(ctx context.Context)
}

type Service struct {
	version           string
	hash              string
	dvAdmin           admin_gateway.ISystem
	settingsService   setting.ISettingService
	permissionService permission.IPermission
	analyticsService  analytics.IAnalytics
	log               logger.Logger
	config            *config.Config
}

func (s *Service) GetIP(ctx context.Context) (string, error) {
	ip, err := s.dvAdmin.CheckIP(ctx)
	if err != nil {
		return "", err
	}
	return ip, nil
}

func New(
	settingsService setting.ISettingService,
	permissionService permission.IPermission,
	dvAdmin admin_gateway.ISystem,
	log logger.Logger,
	appVersion, commitHash string,
	cfg *config.Config,
	analyticsService analytics.IAnalytics,
) ISystemService {
	return &Service{
		version:           appVersion,
		hash:              commitHash,
		settingsService:   settingsService,
		permissionService: permissionService,
		dvAdmin:           dvAdmin,
		log:               log,
		config:            cfg,
		analyticsService:  analyticsService,
	}
}

var _ ISystemService = (*Service)(nil)

func (s *Service) GetInfo(ctx context.Context) (*models.SystemInfo, error) {
	resp := &models.SystemInfo{
		IsTurnstileEnabled: s.config.Turnstile.Enabled,
		TurnstileSiteKey:   s.config.Turnstile.SiteKey,
	}

	// Check if root user exists
	rootUsers, err := s.permissionService.RoleUsers(models.UserRoleRoot)
	if err != nil {
		return nil, err
	}
	resp.RootUserExists = len(rootUsers) > 0

	// Check if processing settings are set - processing initialized in backend
	defaultProcessingSettings := []string{setting.ProcessingURL, setting.ProcessingClientID, setting.ProcessingClientKey}
	settings, err := s.settingsService.GetRootSettingsByNames(ctx, defaultProcessingSettings)
	if err != nil {
		return nil, err
	}

	if len(settings) == len(defaultProcessingSettings) {
		resp.Initialized = true
	}

	resp.AppProfile = s.GetProfile(ctx)

	// Check if registration is enabled
	registrationSetting, err := s.settingsService.GetRootSetting(ctx, setting.RegistrationState)
	if err != nil && errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	if registrationSetting != nil {
		resp.RegistrationState = registrationSetting.Value
	}

	return resp, nil
}

func (s *Service) GetAppVersion(_ context.Context) (string, string) {
	return s.version, s.hash
}

func (s *Service) RunHeartbeatLoop(ctx context.Context) {
	err := s.runHeartBeat(ctx)
	if err == nil {
		s.log.Info("dv admin was notified")
		return
	}

	s.log.Warn("dv admin version notify", "error", err)

	ticker := time.NewTicker(max(s.config.Admin.PingVersionInterval, time.Minute*30))
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err = s.runHeartBeat(ctx); err != nil {
				s.log.Warn("admin error", "error", err)
				continue
			}

			return
		case <-ctx.Done():
			return
		}
	}
}

func (s *Service) GetProfile(_ context.Context) models.AppProfile {
	return s.config.App.Profile
}

func (s *Service) runHeartBeat(ctx context.Context) error {
	clID, err := s.settingsService.GetRootSetting(ctx, setting.ProcessingClientID)
	if err != nil || clID == nil {
		// processing is not yet initialized
		return errors.New("processing client ID is not set")
	}

	adminSecret, err := s.settingsService.GetRootSetting(ctx, setting.DvAdminSecretKey)
	if err != nil || adminSecret == nil {
		// admin secret uninitialized
		return errors.New("admin secret key is not set")
	}
	dvContext := admin_gateway.PrepareServiceContext(ctx, adminSecret.Value, clID.Value)

	var analyticsData *admin_requests.AnalyticsData
	enabled, err := s.analyticsService.IsAnalyticsEnabled(dvContext)
	if err != nil {
		return fmt.Errorf("check analytics enabled: %w", err)
	}
	if enabled {
		data, err := s.analyticsService.GetAnalyticsData(dvContext)
		if err != nil {
			return fmt.Errorf("get analytics data: %w", err)
		}
		analyticsData = data.ConvertToAdminModel()
	}

	if err := s.dvAdmin.HeartBeat(dvContext, analyticsData); err != nil {
		return err
	}

	return nil
}
