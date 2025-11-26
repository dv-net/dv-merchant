package setting_request

type TransfersToggleRequest struct {
	Mode string `json:"mode,omitempty" validation:"required,oneof=enabled disabled" enums:"enabled,disabled"`
} //	@name	TransfersToggleRequest

type TronTransfersTypeRequest struct {
	Type string `json:"type,omitempty" validation:"required,oneof=burntrx resources" enums:"burntrx,resources"`
} //	@name	TronTransfersTypeRequest
