package dto

type LogDTO struct {
	Level   string `json:"level"`
	Message string `json:"message"`
	Time    string `json:"time"`
	Fields  string `json:"fields"`
}
