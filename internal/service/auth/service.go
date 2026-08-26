package auth

import (
	"context"
	"errors"
	"fmt"
	"hash/crc32"
	"strings"
	"time"

	"github.com/dv-net/dv-merchant/internal/config"
	"github.com/dv-net/dv-merchant/internal/delivery/http/request/auth_request"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/notify"
	"github.com/dv-net/dv-merchant/internal/service/setting"
	"github.com/dv-net/dv-merchant/internal/service/user"
	"github.com/dv-net/dv-merchant/internal/service/wallet"
	"github.com/dv-net/dv-merchant/internal/storage"
	"github.com/dv-net/dv-merchant/internal/storage/repos"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_personal_access_tokens"
	"github.com/dv-net/dv-merchant/internal/tools"
	"github.com/dv-net/dv-merchant/internal/tools/hash"
	"github.com/dv-net/dv-merchant/internal/tools/str"
	"github.com/dv-net/dv-merchant/internal/util"
	"github.com/dv-net/dv-merchant/pkg/logger"
	"github.com/dv-net/dv-merchant/pkg/otp"
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
	AuthByWallet(ctx context.Context, w *models.Wallet) (*Token, error)
	GetWalletByToken(ctx context.Context, hashedToken string) (*models.Wallet, error)
	SendWalletVerificationCode(ctx context.Context, walletID, storeID uuid.UUID, email string) error
	VerifyWalletCode(ctx context.Context, walletID, storeID uuid.UUID, email, code string) (*Token, error)
}

type Service struct {
	cfg              *config.Config
	logger           logger.Logger
	storage          storage.IStorage
	userService      user.IUser
	userCredsService user.IUserCredentials
	notifyService    notify.INotificationService
	settingsService  setting.ISettingService
	walletService    wallet.IWalletService
	otpService       *otp.Service
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
	walletService wallet.IWalletService,
	otpService *otp.Service,
) *Service {
	return &Service{
		cfg:              cfg,
		logger:           logger,
		userService:      userService,
		userCredsService: userCredsService,
		storage:          storage,
		notifyService:    notifyService,
		settingsService:  settingsService,
		walletService:    walletService,
		otpService:       otpService,
	}
}

func (s *Service) RegisterUser(ctx context.Context, dto *user.CreateUserDTO) (*user.RegisterUserDTO, error) {
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

func (s *Service) Auth(ctx context.Context, dto auth_request.AuthRequest) (*Token, error) {
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

	token, err := s.createNewToken(ctx, "user", userForAuth.ID, "AuthToken", expiresAt)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func (s *Service) AuthByUser(ctx context.Context, user *models.User) (*Token, error) {
	return s.createNewToken(ctx, "user", user.ID, "AuthToken", util.Pointer(time.Now().Add(time.Hour*24)))
}

func (s *Service) GetUserByToken(ctx context.Context, hashedToken string) (*models.User, error) {
	token, err := s.resolveToken(ctx, hashedToken)
	if err != nil {
		return nil, err
	}

	u, err := s.userService.GetUserByID(ctx, token.TokenableID)
	if err != nil {
		return nil, errors.New("user not found")
	}
	return u, nil
}

func (s *Service) AuthByWallet(ctx context.Context, w *models.Wallet) (*Token, error) {
	return s.createNewToken(ctx, "wallet", w.ID, "WalletAuthToken", util.Pointer(time.Now().Add(time.Hour*3)))
}

func (s *Service) GetWalletByToken(ctx context.Context, hashedToken string) (*models.Wallet, error) {
	token, err := s.resolveToken(ctx, hashedToken)
	if err != nil {
		return nil, err
	}

	w, err := s.walletService.GetWallet(ctx, token.TokenableID)
	if err != nil {
		return nil, errors.New("wallet not found")
	}
	return w, nil
}

func (s *Service) SendWalletVerificationCode(ctx context.Context, walletID, storeID uuid.UUID, email string) error {
	w, err := s.matchWalletForAuth(ctx, walletID, storeID, email)
	if err != nil {
		return err
	}

	if err := s.storage.KeyValue().IncrementCounterWithLimit(ctx, walletCodeResendCooldownKey(walletID), 1,
		walletCodeResendCooldown); err != nil {
		return err
	}

	code, err := s.otpService.InitStringCode(ctx, "", walletOTPPurpose(walletID), generateWalletCode)
	if err != nil {
		return err
	}

	go s.notifyService.SendSystemEmail(ctx, models.NotificationTypeRefundVerificationCode, email, &notify.RefundVerificationCodeData{
		Language: w.Locale,
		Code:     code,
	}, &models.NotificationArgs{StoreID: &storeID})

	return nil
}

func (s *Service) VerifyWalletCode(ctx context.Context, walletID, storeID uuid.UUID, email, code string) (*Token, error) {
	w, err := s.matchWalletForAuth(ctx, walletID, storeID, email)
	if err != nil {
		return nil, err
	}

	if err := s.otpService.VerifyStringCode(ctx, strings.ToUpper(code), "", walletOTPPurpose(walletID)); err != nil {
		return nil, ErrInvalidOTPCode
	}

	return s.AuthByWallet(ctx, w)
}

func (s *Service) matchWalletForAuth(ctx context.Context, walletID, storeID uuid.UUID, email string) (*models.Wallet, error) {
	w, err := s.walletService.GetWallet(ctx, walletID)
	if err != nil {
		return nil, ErrWalletNotFound
	}
	if w.StoreID != storeID {
		return nil, ErrWalletNotFound
	}
	if !w.Email.Valid || !strings.EqualFold(w.Email.String, email) {
		return nil, ErrEmailMismatch
	}
	return w, nil
}

func (s *Service) resolveToken(ctx context.Context, hashedToken string) (*models.PersonalAccessToken, error) {
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

	return token, nil
}

func (s *Service) createNewToken(
	ctx context.Context,
	tType string,
	tID uuid.UUID,
	name string,
	expires *time.Time,
) (*Token, error) {
	token, err := generateTokenString()
	if err != nil {
		return nil, err
	}
	params := repo_personal_access_tokens.CreateParams{
		TokenableType: tType,
		TokenableID:   tID,
		Name:          name,
		Token:         hash.SHA256(token.FullToken),
		ExpiresAt:     expires,
	}

	_, err = s.storage.PersonalAccessToken().Create(ctx, params)
	if err != nil {
		return nil, err
	}

	return token, nil
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

func (s *Service) setDefaultUserSettings(ctx context.Context, user *models.User, opt repos.Option) error {
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
