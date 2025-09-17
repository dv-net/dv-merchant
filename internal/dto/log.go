package dto

import "time"

type LogDTO struct {
	Level   string    `json:"level"`
	Message string    `json:"message"`
	Time    time.Time `json:"time"`
}
