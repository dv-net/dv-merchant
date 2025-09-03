package analytics

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"time"

	"github.com/dv-net/dv-merchant/internal/cache"
	"github.com/dv-net/dv-merchant/internal/service/processing"
	"github.com/dv-net/dv-merchant/internal/service/setting"
	"github.com/dv-net/dv-merchant/internal/service/updater"
	"github.com/dv-net/dv-merchant/internal/storage"
	"github.com/dv-net/dv-merchant/pkg/admin_gateway"
	admin_requests "github.com/dv-net/dv-merchant/pkg/admin_gateway/requests"

	"github.com/shirou/gopsutil/v4/host"
	"github.com/shopspring/decimal"
)

const (
	SystemInfoCacheKey = "system_info_cache_key"
)

type EnvironmentInfo struct {
	BackendVersionHash    string `json:"backend_version_hash"`
	BackendVersionTag     string `json:"backend_version_tag"`
	ProcessingVersionHash string `json:"processing_version_hash"`
	ProcessingVersionTag  string `json:"processing_version_tag"`
	UpdaterVersionHash    string `json:"updater_version_hash"`
	UpdaterVersionTag     string `json:"updater_version_tag"`
}

type SystemInfo struct {
	OS             string    `json:"os"`
	Architecture   string    `json:"architecture"`
	Distribution   string    `json:"distribution"`
	Family         string    `json:"family"`
	IsVM           bool      `json:"is_vm"`
	Version        string    `json:"version"`
	GoVersion      string    `json:"go_version"`
	CollectionTime time.Time `json:"collection_time"`
}

type TransferStatsByCurrency struct {
	CurrencyID     string          `json:"currency_id"`
	CurrencyCode   string          `json:"currency_code"`
	TotalCount     int64           `json:"total_count"`
	TotalAmount    decimal.Decimal `json:"total_amount"`
	TotalAmountUSD decimal.Decimal `json:"total_amount_usd"`
}

type AnalyticsData struct { //nolint:revive
	EnvironmentInfo *EnvironmentInfo          `json:"environment_info"`
	SystemInfo      *SystemInfo               `json:"system_info"`
	TransferStats   []TransferStatsByCurrency `json:"transfer_stats"`
}

func (o AnalyticsData) ConvertToAdminModel() *admin_requests.AnalyticsData {
	transferStats := make([]admin_requests.TransferStatsByCurrency, len(o.TransferStats))
	for i, stat := range o.TransferStats {
		transferStats[i] = admin_requests.TransferStatsByCurrency{
			CurrencyID:   stat.CurrencyID,
			CurrencyCode: stat.CurrencyCode,
			TotalCount:   stat.TotalCount,
			TotalAmount:  stat.TotalAmount,
		}
	}

	var systemInfo *admin_requests.SystemInfo
	if o.SystemInfo != nil {
		systemInfo = &admin_requests.SystemInfo{
			OS:             o.SystemInfo.OS,
			Architecture:   o.SystemInfo.Architecture,
			Version:        o.SystemInfo.Version,
			GoVersion:      o.SystemInfo.GoVersion,
			CollectionTime: o.SystemInfo.CollectionTime,
		}
	}

	var environmentInfo *admin_requests.EnvironmentInfo
	if o.EnvironmentInfo != nil {
		environmentInfo = &admin_requests.EnvironmentInfo{
			BackendVersionHash:    o.EnvironmentInfo.BackendVersionHash,
			BackendVersionTag:     o.EnvironmentInfo.BackendVersionTag,
			ProcessingVersionHash: o.EnvironmentInfo.ProcessingVersionHash,
			ProcessingVersionTag:  o.EnvironmentInfo.ProcessingVersionTag,
			UpdaterVersionHash:    o.EnvironmentInfo.UpdaterVersionHash,
			UpdaterVersionTag:     o.EnvironmentInfo.UpdaterVersionTag,
		}
	}

	request := &admin_requests.AnalyticsData{
		EnvironmentInfo: environmentInfo,
		SystemInfo:      systemInfo,
		TransferStats:   transferStats,
	}
	return request
}

var _ IAnalytics = (*Service)(nil)

type IAnalytics interface {
	GetSystemInfo(ctx context.Context) (*SystemInfo, error)
	GetTransferStatsByCurrency(ctx context.Context) ([]TransferStatsByCurrency, error)
	GetAnalyticsData(ctx context.Context) (*AnalyticsData, error)
	IsAnalyticsEnabled(ctx context.Context) (bool, error)
}

type Service struct {
	storage       storage.IStorage
	cache         cache.ICache
	settings      setting.IRootSettings
	adminGateway  *admin_gateway.Service
	processingSvc processing.IProcessingSystem
	updaterSvc    updater.IUpdateClient
	appVersion    string
	appHash       string
}

func NewService(
	storage storage.IStorage,
	cache cache.ICache,
	settings setting.IRootSettings,
	adminGateway *admin_gateway.Service,
	processingSvc processing.IProcessingSystem,
	updaterSvc updater.IUpdateClient,
	appVersion, appHash string,
) *Service {
	return &Service{
		storage:       storage,
		cache:         cache,
		settings:      settings,
		adminGateway:  adminGateway,
		processingSvc: processingSvc,
		updaterSvc:    updaterSvc,
		appVersion:    appVersion,
		appHash:       appHash,
	}
}

func (s *Service) GetEnvironmentInfo(ctx context.Context) (*EnvironmentInfo, error) {
	environmentInfo := &EnvironmentInfo{}

	// Get processing version from processing service
	processingInfo, err := s.processingSvc.GetProcessingSystemInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("get processing version: %w", err)
	}
	environmentInfo.ProcessingVersionHash = processingInfo.Hash
	environmentInfo.ProcessingVersionTag = processingInfo.Version

	// Get backend version from injected values
	environmentInfo.BackendVersionHash = s.appHash
	environmentInfo.BackendVersionTag = s.appVersion

	if err := s.updaterSvc.Ping(ctx); err == nil {
		updaterInfo, err := s.updaterSvc.GetUpdaterVersion(ctx)
		if err != nil {
			return nil, fmt.Errorf("get updater version: %w", err)
		}

		environmentInfo.UpdaterVersionHash = updaterInfo.AppVersion
		environmentInfo.UpdaterVersionTag = updaterInfo.AppCommit
	}
	return environmentInfo, nil
}

func (s *Service) GetSystemInfo(ctx context.Context) (*SystemInfo, error) {
	// Try to get from cache first
	if cachedValue, err := s.storage.KeyValue().Get(ctx, SystemInfoCacheKey); err == nil {
		systemInfo := &SystemInfo{}
		if err := json.Unmarshal(cachedValue, systemInfo); err != nil {
			return nil, fmt.Errorf("unmarshal cached system info: %w", err)
		}
		return systemInfo, nil
	}

	// If not in cache, fetch from system
	hostInfo, err := host.Info()
	if err != nil {
		return nil, fmt.Errorf("get host info: %w", err)
	}

	systemInfo := &SystemInfo{
		OS:             hostInfo.OS,
		Architecture:   hostInfo.KernelArch,
		Distribution:   hostInfo.Platform,
		Family:         hostInfo.PlatformFamily,
		IsVM:           hostInfo.VirtualizationRole == "guest",
		Version:        hostInfo.PlatformVersion,
		GoVersion:      runtime.Version(),
		CollectionTime: time.Now(),
	}

	// Cache the result
	if err := s.storage.KeyValue().Set(ctx, SystemInfoCacheKey, systemInfo, -1); err != nil {
		return nil, fmt.Errorf("set system info cache: %w", err)
	}

	return systemInfo, nil
}

func (s *Service) GetTransferStatsByCurrency(ctx context.Context) ([]TransferStatsByCurrency, error) {
	stats, err := s.storage.Analytics().GetTransactionStatsByType(ctx)
	if err != nil {
		return nil, fmt.Errorf("get transfer stats: %w", err)
	}

	result := make([]TransferStatsByCurrency, len(stats))
	for i, stat := range stats {
		result[i] = TransferStatsByCurrency{
			CurrencyID:   stat.CurrencyID,
			CurrencyCode: stat.CurrencyCode,
			TotalCount:   stat.TotalCount,
			TotalAmount:  stat.TotalAmount,
		}
	}

	return result, nil
}

func (s *Service) IsAnalyticsEnabled(ctx context.Context) (bool, error) {
	settingValue, err := s.settings.GetRootSetting(ctx, setting.AnonymousTelemetry)
	if err != nil {
		return false, err
	}

	return settingValue.Value == setting.FlagValueEnabled, nil
}

func (s *Service) GetAnalyticsData(ctx context.Context) (*AnalyticsData, error) {
	enabled, err := s.IsAnalyticsEnabled(ctx)
	if err != nil {
		return nil, err
	}

	systemInfo, err := s.GetSystemInfo(ctx)
	if err != nil {
		return nil, err
	}

	environmentInfo, err := s.GetEnvironmentInfo(ctx)
	if err != nil {
		return nil, err
	}

	if !enabled {
		return &AnalyticsData{
			SystemInfo:      systemInfo,
			EnvironmentInfo: environmentInfo,
		}, nil
	}

	transferStats, err := s.GetTransferStatsByCurrency(ctx)
	if err != nil {
		return nil, err
	}

	return &AnalyticsData{
		SystemInfo:      systemInfo,
		EnvironmentInfo: environmentInfo,
		TransferStats:   transferStats,
	}, nil
}
