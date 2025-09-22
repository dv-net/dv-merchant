package user

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/dv-net/dv-merchant/internal/dto"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/notify"
	"github.com/dv-net/dv-merchant/internal/storage/repos"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_personal_access_tokens"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_users"
	"github.com/dv-net/dv-merchant/internal/tools"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type IUserCredentials interface {
	InitPasswordReset(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, dto ResetPasswordDto) error
	ChangeUserPassword(ctx context.Context, userID uuid.UUID, newPassword string) error
	ChangePassword(ctx context.Context, user *models.User, d dto.ChangePasswordDTO, currentToken string, opts ...repos.Option) error
	InitEmailConfirmation(ctx context.Context, usr *models.User) error
	ConfirmEmail(ctx context.Context, user *models.User, confirmationCode int) error
	InitEmailChange(ctx context.Context, usr *models.User) error
	ConfirmEmailChange(ctx context.Context, user *models.User, dto ChangeEmailConfirmationDto) error
}

var _ IUserCredentials = (*Service)(nil)

func (s *Service) InitEmailChange(ctx context.Context, usr *models.User) error {
	if !usr.EmailVerifiedAt.Valid {
		// If emil is not verified - no 2fa code is required
		return nil
	}

	code, err := s.otpSvc.InitCode(ctx, "", usr.Email)
	if err != nil {
		s.logger.Error("init otp code", err)
		return errors.New("failed to init otp code")
	}

	payload := &notify.UserInitEmailChangeData{
		Language: usr.Language,
		Code:     code,
	}
	go s.notificationService.SendUser(ctx, models.NotificationTypeUserEmailChange, usr, payload, &models.NotificationArgs{
		UserID: &usr.ID,
	})

	return nil
}

func (s *Service) ConfirmEmailChange(ctx context.Context, user *models.User, dto ChangeEmailConfirmationDto) error {
	if dto.NewEmail != dto.NewEmailConfirmation {
		return errors.New("email confirmation mismatch")
	}

	// No verification code required if old email was not confirmed
	if !user.EmailVerifiedAt.Valid {
		return s.changeUserEmail(ctx, user, dto.NewEmail)
	}

	code, err := strconv.Atoi(dto.Code)
	if err != nil {
		return fmt.Errorf("invalid verification code format")
	}

	if err := s.otpSvc.VerifyCode(ctx, code, "", user.Email); err != nil {
		s.logger.Warn("failed to verify email code", err)
		return fmt.Errorf("incorrect verification code or email")
	}

	if err := s.changeUserEmail(ctx, user, dto.NewEmail); err != nil {
		s.logger.Error("change email with logout error", err)
		return errors.New("change email with logout error")
	}

	payload := &notify.EmailChangeConfirmData{
		Language: user.Language,
		Email:    user.Email,
		NewEmail: dto.NewEmail,
	}

	go s.notificationService.SendUser(ctx, models.NotificationTypeUserEmailReset, user, payload, &models.NotificationArgs{
		UserID: &user.ID,
	})

	return nil
}

func (s *Service) changeUserEmail(ctx context.Context, usr *models.User, newEmail string) error {
	if err := s.storage.Users().ChangeEmail(ctx, repo_users.ChangeEmailParams{
		NewEmail: newEmail,
		OldEmail: usr.Email,
	}); err != nil {
		s.logger.Error("failed to change email", err)
		return errors.New("failed to change email")
	}

	return nil
}

func (s *Service) ChangePassword(ctx context.Context, user *models.User, d dto.ChangePasswordDTO, currentToken string, opts ...repos.Option) error {
	if !user.EmailVerifiedAt.Valid {
		return errors.New("email verificatino is required for password change")
	}

	if !tools.CheckPasswordHash(d.OldPassword, user.Password) {
		return errors.New("invalid credentials")
	}

	hashPassword, err := tools.HashPassword(d.NewPassword)
	if err != nil {
		return errors.New("can't hash password")
	}

	params := repo_users.ChangePasswordParams{
		Password: hashPassword,
		ID:       user.ID,
	}
	_, err = s.storage.Users(opts...).ChangePassword(ctx, params)
	if err != nil {
		return err
	}

	payload := &notify.UserPasswordChanged{
		Language: user.Language,
	}

	err = s.storage.PersonalAccessToken(opts...).ClearAllByUser(ctx, repo_personal_access_tokens.ClearAllByUserParams{
		TokenableID: user.ID,
		Token:       currentToken,
	})
	if err != nil {
		return err
	}

	go s.notificationService.SendUser(ctx, models.NotificationTypeUserPasswordChanged, user, payload, &models.NotificationArgs{UserID: &user.ID})

	return nil
}

func (s *Service) InitPasswordReset(ctx context.Context, email string) error {
	user, err := s.storage.Users().GetByEmail(ctx, email)
	if err != nil {
		s.logger.Error("fetch user by email", err)
		return nil
	}

	code, err := s.otpSvc.InitCode(ctx, "", email)
	if err != nil {
		s.logger.Error("init otp code", err)
		return errors.New("failed to init otp code")
	}

	payload := &notify.UserForgotPassword{
		Language: user.Language,
		Code:     strconv.Itoa(code),
	}

	go s.notificationService.SendUser(ctx, models.NotificationTypeUserForgotPassword, user, payload, &models.NotificationArgs{UserID: &user.ID})

	return nil
}

func (s *Service) ResetPassword(ctx context.Context, dto ResetPasswordDto) error {
	return repos.BeginTxFunc(ctx, s.storage.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		usr, err := s.storage.Users(repos.WithTx(tx)).GetByEmail(ctx, dto.Email)
		if err != nil {
			return errors.New("user not found")
		}

		if !usr.EmailVerifiedAt.Valid {
			return errors.New("email not verified")
		}

		if err = s.otpSvc.VerifyCode(ctx, dto.Code, "", usr.Email); err != nil {
			s.logger.Warn("failed to verify email", err)
			return errors.New("failed to verify email")
		}

		if dto.NewPassword != dto.ConfirmPassword {
			return fmt.Errorf("password mismatch")
		}

		if err = s.ChangeUserPassword(ctx, usr.ID, dto.NewPassword); err != nil {
			return err
		}

		payload := &notify.UserPasswordChanged{
			Language: usr.Language,
		}
		go s.notificationService.SendUser(ctx, models.NotificationTypeUserPasswordChanged, usr, payload, &models.NotificationArgs{UserID: &usr.ID})

		return nil
	})
}

func (s *Service) ChangeUserPassword(ctx context.Context, userID uuid.UUID, newPassword string) error {
	hashPassword, err := tools.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	if _, err := s.storage.Users().ChangePassword(ctx, repo_users.ChangePasswordParams{
		Password: hashPassword,
		ID:       userID,
	}); err != nil {
		return fmt.Errorf("failed to change user password: %w", err)
	}

	return nil
}

func (s *Service) InitEmailConfirmation(ctx context.Context, usr *models.User) error {
	if usr.EmailVerifiedAt.Valid {
		return errors.New("email is already confirmed")
	}

	code, err := s.otpSvc.InitCode(ctx, "", usr.Email)
	if err != nil {
		s.logger.Error("init otp code", err)
		return errors.New("failed to init otp code")
	}

	go s.sendUserEmailConfirmation(ctx, *usr, code)

	return nil
}

func (s *Service) ConfirmEmail(ctx context.Context, user *models.User, confirmationCode int) error {
	if err := s.otpSvc.VerifyCode(ctx, confirmationCode, "", user.Email); err != nil {
		return fmt.Errorf("incorrect confirmation code or email")
	}

	if _, err := s.storage.Users().UpdateEmailVerifiedAt(ctx, user.ID); err != nil {
		return fmt.Errorf("failed to update user verification status: %w", err)
	}

	payload := &notify.UserRegistration{
		Language: user.Language,
	}

	go s.notificationService.SendUser(ctx, models.NotificationTypeUserRegistration, user, payload, &models.NotificationArgs{UserID: &user.ID})

	return nil
}
