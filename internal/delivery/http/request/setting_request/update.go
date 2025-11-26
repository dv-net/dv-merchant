package setting_request

type Setting struct {
	Name  string `json:"name" validate:"required,min=4,max=255"`
	Value string `json:"value" validate:"required,min=2,max=255"`
} //	@name	Setting
