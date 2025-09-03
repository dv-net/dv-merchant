package templater

import (
	"errors"
	"fmt"
)

var (
	ErrPartialNotFound  = errors.New("partial not found")
	ErrTemplateNotFound = errors.New("template not found")
	ErrPayloadNil       = errors.New("payload is nil")
	ErrSettingsLoad     = errors.New("failed to load mailer settings")
)

func wrapError(err error, msg string, args ...any) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf(fmt.Sprintf(msg, args...)+": %w", err)
}
