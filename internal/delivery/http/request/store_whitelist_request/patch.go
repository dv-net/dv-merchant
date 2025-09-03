package store_whitelist_request

type PatchRequest struct {
	IP string `db:"ip" json:"ip" validate:"required,ip" format:"ipv4"`
} // @name PatchWhitelistRequest
