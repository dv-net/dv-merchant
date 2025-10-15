package auth

import (
	"context"
	"errors"
	"fmt"
	"hash/crc32"
	"time"

	"github.com/dv-net/dv-merchant/internal/config"
	"github.com/dv-net/dv-merchant/internal/delivery/http/request/auth_request"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/notify"
	"github.com/dv-net/dv-merchant/internal/service/setting"
	"github.com/dv-net/dv-merchant/internal/service/user"
	"github.com/dv-net/dv-merchant/internal/storage"
	"github.com/dv-net/dv-merchant/internal/storage/repos"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_personal_access_tokens"
	"github.com/dv-net/dv-merchant/internal/tools"
	"github.com/dv-net/dv-merchant/internal/tools/hash"
	"github.com/dv-net/dv-merchant/internal/tools/str"
	"github.com/dv-net/dv-merchant/internal/util"
	"github.com/dv-net/dv-merchant/pkg/logger"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// ClientInfo contains information about the client device/browser
type ClientInfo struct {
	IP        string // IP address of the client
	UserAgent string // User-Agent header from the client
}

type IAuth interface {
	RegisterUser(ctx context.Context, dto *user.CreateUserDTO) (*user.RegisterUserDTO, error)
	Auth(ctx context.Context, dto auth_request.AuthRequest) (*Token, error)
	GetUserByToken(ctx context.Context, hashedToken string) (*models.User, error)
	AuthByUser(ctx context.Context, user *models.User) (*Token, error)
}

type Service struct {
	cfg              *config.Config
	logger           logger.Logger
	userService      user.IUser
	userCredsService user.IUserCredentials
	storage          storage.IStorage
	notifyService    notify.INotificationService
	settingsService  setting.ISettingService
}

type Token struct {
	TokenEntropy string
	CRC32BHash   string
	FullToken    string
}

func New(
	cfg *config.Config,
	logger logger.Logger,
	storage storage.IStorage,
	userService user.IUser,
	userCredsService user.IUserCredentials,
	notifyService notify.INotificationService,
	settingsService setting.ISettingService,
) *Service {
	return &Service{
		cfg:              cfg,
		logger:           logger,
		userService:      userService,
		userCredsService: userCredsService,
		storage:          storage,
		notifyService:    notifyService,
		settingsService:  settingsService,
	}
}

func (s Service) RegisterUser(ctx context.Context, dto *user.CreateUserDTO) (*user.RegisterUserDTO, error) {
	var rUserDto *user.RegisterUserDTO
	err := repos.BeginTxFunc(ctx, s.storage.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		registeredUser, err := s.userService.StoreUser(ctx, dto, repos.WithTx(tx))
		if err != nil {
			return err
		}
		rUserDto = registeredUser

		if err := s.setDefaultUserSettings(ctx, registeredUser.User, repos.WithTx(tx)); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return rUserDto, nil
}

func (s Service) Auth(ctx context.Context, dto auth_request.AuthRequest) (*Token, error) {
	userForAuth, err := s.userService.GetUserByEmail(ctx, dto.Email)
	if err != nil {
		return nil, err
	}

	if userForAuth.Banned.Bool {
		return nil, err
	}

	if !tools.CheckPasswordHash(dto.Password, userForAuth.Password) {
		return nil, err
	}

	var expiresAt *time.Time
	if !dto.RememberMe {
		expiresAt = util.Pointer(time.Now().Add(time.Hour * 24))
	}

	token, err := s.createNewToken(ctx, expiresAt, userForAuth.ID)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func (s Service) createNewToken(ctx context.Context, expires *time.Time, userID uuid.UUID) (*Token, error) {
	token, err := generateTokenString()
	if err != nil {
		return nil, err
	}
	params := repo_personal_access_tokens.CreateParams{
		TokenableType: "user",
		TokenableID:   userID,
		Name:          "AuthToken",
		Token:         hash.SHA256(token.FullToken),
		ExpiresAt:     expires,
	}

	_, err = s.storage.PersonalAccessToken().Create(ctx, params)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func (s Service) AuthByUser(ctx context.Context, user *models.User) (*Token, error) {
	token, err := generateTokenString()
	if err != nil {
		return nil, err
	}

	params := repo_personal_access_tokens.CreateParams{
		TokenableType: "user",
		TokenableID:   user.ID,
		Name:          "AuthToken",
		Token:         hash.SHA256(token.FullToken),
		ExpiresAt:     util.Pointer(time.Now().Add(time.Hour * 24)),
	}

	_, err = s.storage.PersonalAccessToken().Create(ctx, params)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (s Service) GetUserByToken(ctx context.Context, hashedToken string) (*models.User, error) {
	token, err := s.storage.PersonalAccessToken().GetByToken(ctx, hashedToken)
	if err != nil || token == nil {
		return nil, ErrTokenExpired
	}

	if token.ExpiresAt != nil && token.ExpiresAt.Before(time.Now()) {
		if removeErr := s.storage.PersonalAccessToken().Delete(ctx, token.ID); removeErr != nil {
			s.logger.Errorw("remove expired token error", "error", removeErr)
		}

		return nil, ErrTokenExpired
	}

	u, err := s.userService.GetUserByID(ctx, token.TokenableID)
	if err != nil {
		return nil, errors.New("user not found")
	}
	return u, nil
}

func generateTokenString() (*Token, error) {
	tokenEntropy, err := str.RandomString(40)
	if err != nil {
		return nil, err
	}
	crc32bHash := fmt.Sprintf("%08x", crc32.ChecksumIEEE([]byte(tokenEntropy)))

	fullToken := fmt.Sprintf("%s%s", tokenEntropy, crc32bHash)

	return &Token{
		TokenEntropy: tokenEntropy,
		CRC32BHash:   crc32bHash,
		FullToken:    fullToken,
	}, nil
}

func (s Service) setDefaultUserSettings(ctx context.Context, user *models.User, opt repos.Option) error {
	settings := []setting.UpdateDTO{
		{
			Name:  setting.TransferType,
			Value: setting.TransferByResource.String(),
			Model: setting.IModelSetting(user),
		},
		{
			Name:  setting.TransfersStatus,
			Value: setting.FlagValueEnabled,
			Model: setting.IModelSetting(user),
		},
		{
			Name:  setting.QuickStartGuideStatus,
			Value: setting.FlagValueIncompleted,
			Model: setting.IModelSetting(user),
		},
		{
			Name:  setting.WithdrawFromProcessing,
			Value: setting.FlagValueEnabled,
			Model: setting.IModelSetting(user),
		},
	}

	for _, sDTO := range settings {
		if err := s.settingsService.SetModelSetting(ctx, sDTO, opt); err != nil {
			return err
		}
	}

	return nil
}
