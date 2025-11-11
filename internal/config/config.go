package config

import (
	"time"

	"github.com/dv-net/dv-merchant/internal/models"

	"github.com/dv-net/mx/logger"
	"github.com/dv-net/mx/ops"
)

type (
	Config struct {
		RolesModelPath      string              `yaml:"roles_model_path" required:"true" default:"configs/rbac_model.conf"`
		RolesPoliciesPath   string              `yaml:"roles_policies_path" required:"true" default:"configs/rbac_policies.csv"`
		App                 AppConfig           `yaml:"app"`
		HTTP                HTTPConfig          `yaml:"http"`
		Seed                SeedConfig          `yaml:"seed"`
		Postgres            PostgresDB          `yaml:"postgres"`
		Redis               RedisDB             `yaml:"redis"`
		Exrate              ExrateConfig        `yaml:"exrate"`
		Admin               AdminConfig         `yaml:"admin"`
		Notify              NotifyConfig        `yaml:"notify"`
		WebHook             WebHook             `yaml:"web_hook"`
		EProxy              EProxy              `yaml:"e_proxy"`
		Transfers           Transfers           `yaml:"transfers"`
		Ops                 ops.Config          `yaml:"ops"`
		KeyValue            KeyValue            `yaml:"key_value"`
		Transactions        Transactions        `yaml:"transactions"`
		Wallets             Wallets             `yaml:"wallets"`
		ExternalStoreLimits ExternalStoreLimits `yaml:"external_store_limits"`
		Log                 logger.Config       `yaml:"log"`
		Blockchain          Blockchain          `yaml:"blockchain"`
		Updater             Updater             `yaml:"updater"`
		Turnstile           Turnstile           `yaml:"turnstile"`
		AML                 AML                 `yaml:"aml"`
	}

	AppConfig struct {
		Profile models.AppProfile `yaml:"profile" default:"prod" validate:"oneof=dev prod demo"`
	}

	HTTPCorsConfig struct {
		Enabled        bool     `yaml:"enabled" default:"true" usage:"allows to disable cors" example:"true / false"`
		AllowedOrigins []string `yaml:"allowed_origins"`
	}

	HTTPConfig struct {
		Host               string         `yaml:"host" default:"localhost"`
		Port               string         `yaml:"port" default:"80"`
		FetchInterval      time.Duration  `yaml:"fetch_interval" env:"FETCH_INTERVAL" default:"30s"`
		ConnectTimeout     time.Duration  `yaml:"connect_timeout" env:"CONNECT_TIMEOUT" default:"5s"`
		ReadTimeout        time.Duration  `yaml:"read_timeout" env:"READ_TIMEOUT" default:"10s"`
		WriteTimeout       time.Duration  `yaml:"write_timeout" env:"WRITE_TIMEOUT" default:"10s"`
		MaxHeaderMegabytes int            `yaml:"max_header_megabytes" env:"MAX_HEADER_MEGABYTES" default:"1"`
		Cors               HTTPCorsConfig `yaml:"cors"`
	}

	SeedConfig struct {
		Base string `yaml:"base" default:"seeds"`
	}

	ExrateConfig struct {
		FetchInterval time.Duration `yaml:"fetch_interval" env:"FETCH_INTERVAL" default:"1m"`
	}

	AdminConfig struct {
		BaseURL             string        `yaml:"base_url" default:"https://api.dv.net/"`
		PingVersionInterval time.Duration `yaml:"ping_version_interval" default:"1h"`
		LogStatus           bool          `yaml:"log_status" default:"false"`
	}

	NotifyConfig struct {
		Telegram NotifyTelegram `yaml:"telegram"`
	}
	NotifyTelegram struct {
		Enabled bool   `yaml:"enabled" default:"false"`
		Token   string `yaml:"token" secret:"true"`
	}

	WebHook struct {
		MaxTries int `yaml:"max_tries" default:"30"`
	}
	Transfers struct {
		GroupSize int `yaml:"group_size" default:"5"`
	}

	KeyValue struct {
		Engine KeyValueEngine `yaml:"engine" required:"true" validate:"oneof=redis in_memory" example:"redis / in_memory" default:"redis"`
	}

	Transactions struct {
		UnconfirmedCollapseInterval time.Duration `yaml:"unconfirmed_collapse_interval" default:"30s"`
	}

	Wallets struct {
		UpdateBalancesInterval      time.Duration `yaml:"update_balances_interval" default:"2s"`
		UpdateTronResourcesInterval time.Duration `yaml:"update_tron_resources_interval" default:"1h"`
	}

	ExternalStoreLimits struct {
		Enabled                bool          `yaml:"enabled" default:"false"`
		RateLimitInterval      time.Duration `yaml:"rate_limit_interval" default:"24h"`
		MaxRequestsPerInterval int64         `yaml:"max_requests_per_interval" default:"3"`
	}

	Turnstile struct {
		Enabled bool   `yaml:"enabled" default:"false"`
		Secret  string `yaml:"secret"`
		SiteKey string `yaml:"site_key"`
		BaseURL string `yaml:"base_url" default:"https://challenges.cloudflare.com"`
	}

	AML struct {
		CheckInterval time.Duration `yaml:"check_interval" default:"2m"`
		CheckTimeout  time.Duration `yaml:"check_timeout" default:"30s"`
		MaxAttempts   int32         `yaml:"max_attempts" default:"5"`

		BitOK  BitOK  `yaml:"bit_ok" required:"true"`
		AMLBot AMLBot `yaml:"aml_bot" required:"true"`
	}

	BitOK struct {
		Enabled bool   `yaml:"enabled" default:"true"`
		BaseURL string `yaml:"base_url" default:"https://kyt-api.bitok.org/"`
	}

	AMLBot struct {
		Enabled bool   `yaml:"enabled" default:"true"`
		BaseURL string `yaml:"base_url" default:"https://extrnlapiendpoint.silencatech.com/"`
	}
)

type KeyValueEngine string

const (
	KeyValueEngineInMemory KeyValueEngine = "in_memory"
	KeyValueEngineRedis    KeyValueEngine = "redis"
)

type GrpcConfig struct {
	Name string `default:"connectrpc-client" validate:"required" example:"backend-connectrpc-client"`
	Addr string `default:"https://explorer-proxy.dv.net" validate:"required" usage:"connectrpc server address" example:"localhost:9000"`
}

type EProxy struct {
	GRPC GrpcConfig `yaml:"grpc"`
}
