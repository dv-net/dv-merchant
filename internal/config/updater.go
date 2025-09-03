package config

type Updater struct {
	BaseURL string `yaml:"base_url" json:"base_url" usage:"allows to set updater service endpoint" default:"http://localhost:8081" example:"http://localhost:8081"`
}
