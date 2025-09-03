package models

type AppProfile string

const (
	AppProfileDev  AppProfile = "dev"
	AppProfileProd AppProfile = "prod"
	AppProfileDemo AppProfile = "demo"
)

type SystemInfo struct {
	AppProfile         AppProfile `json:"app_profile"`
	Initialized        bool       `json:"initialized"`
	RootUserExists     bool       `json:"root_user_exists"`
	RegistrationState  string     `json:"registration_state"`
	IsTurnstileEnabled bool       `json:"is_turnstile_enabled"`
	TurnstileSiteKey   string     `json:"turnstile_site_key"`
} // @name SystemInfo
