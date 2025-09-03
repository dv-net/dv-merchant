package util

import (
	"testing"
	"time"
)

func TestParseDate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantErr  bool
		wantTime *time.Time
	}{
		{
			name:    "Valid RFC3339 with timezone",
			input:   "2025-07-22T10:15:00+02:00",
			wantErr: false,
			wantTime: func() *time.Time {
				t, _ := time.Parse(time.RFC3339, "2025-07-22T10:15:00+02:00")
				return &t
			}(),
		},
		{
			name:    "Valid DateTime without timezone",
			input:   "2025-07-22 10:15:00",
			wantErr: false,
			wantTime: func() *time.Time {
				t, _ := time.Parse(time.DateTime, "2025-07-22 10:15:00")
				return &t
			}(),
		},
		{
			name:     "Invalid format",
			input:    "2025-07-22-invalid",
			wantErr:  true,
			wantTime: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == nil {
				t.Errorf("ParseDate() returned nil, expected a time value")
				return
			}
			if !tt.wantErr && tt.wantTime != nil && !got.Equal(*tt.wantTime) {
				t.Errorf("ParseDate() = %v, want %v", got, tt.wantTime)
			}
		})
	}
}
