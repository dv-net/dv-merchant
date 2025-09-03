package store_whitelist_request

type UpdateRequest struct {
	Ips []string `db:"ips" json:"ips" validate:"required,dive,ip" format:"ipv4"`
} // @name UpdateWhitelistRequest
