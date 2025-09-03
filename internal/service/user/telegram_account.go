package user

import (
	"context"

	"github.com/dv-net/dv-merchant/internal/models"
)

type TelegramAccount interface {
	GenerateTgLink(context.Context, *models.User) (string, error)
	InitTgInUnlink(context.Context, *models.User) error
	ConfirmTgUnlink(context.Context, *models.User, string) error
}

func (s *Service) InitTgInUnlink(ctx context.Context, user *models.User) error {
	ctx, err := s.prepareDvNetContext(ctx)
	if err != nil {
		return err
	}

	if err = s.ensureDVAuthDataValid(user); err != nil {
		return err
	}

	if err = s.dvAuth.UnlinkOwnerTg(ctx, user.ProcessingOwnerID.UUID, user.DvnetToken.String); err != nil {
		return err
	}

	return nil
}

func (s *Service) ConfirmTgUnlink(ctx context.Context, user *models.User, otp string) error {
	ctx, err := s.prepareDvNetContext(ctx)
	if err != nil {
		return err
	}

	if err = s.ensureDVAuthDataValid(user); err != nil {
		return err
	}

	if err = s.dvAuth.ConfirmUnlinkTg(ctx, otp, user.DvnetToken.String); err != nil {
		return err
	}

	return nil
}

func (s *Service) GenerateTgLink(ctx context.Context, user *models.User) (string, error) {
	ctx, err := s.prepareDvNetContext(ctx)
	if err != nil {
		return "", err
	}

	if err = s.ensureDVAuthDataValid(user); err != nil {
		return "", err
	}

	res, err := s.dvAuth.InitOwnerTg(ctx, user.DvnetToken.String)
	if err != nil {
		return "", err
	}

	return res.Link, nil
}
