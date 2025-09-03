package setting

type UpdateDto struct {
	Name  string
	Value string
}

type Dto struct {
	Name                          string   `json:"name"`
	Value                         *string  `json:"value"`
	IsEditable                    bool     `json:"is_editable"`
	TwoFactorVerificationRequired bool     `json:"two_factor_verification_required"`
	AvailableValues               []string `json:"available_values"`
}

type ByName []Dto

func (a ByName) Len() int           { return len(a) }
func (a ByName) Less(i, j int) bool { return a[i].Name < a[j].Name }
func (a ByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
