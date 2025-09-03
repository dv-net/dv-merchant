package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/dv-net/dv-merchant/internal/config"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/notify"
	"github.com/dv-net/dv-merchant/internal/service/permission"
	"github.com/dv-net/dv-merchant/internal/service/processing"
	"github.com/dv-net/dv-merchant/internal/service/setting"
	"github.com/dv-net/dv-merchant/internal/service/store"
	"github.com/dv-net/dv-merchant/internal/storage"
	"github.com/dv-net/dv-merchant/internal/storage/repos"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_stores"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_users"
	"github.com/dv-net/dv-merchant/internal/tools"
	"github.com/dv-net/dv-merchant/pkg/admin_gateway"
	admin_requests "github.com/dv-net/dv-merchant/pkg/admin_gateway/requests"
	"github.com/dv-net/dv-merchant/pkg/logger"
	"github.com/dv-net/dv-merchant/pkg/otp"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
)

type IUser interface { //nolint:interfacebloat
	IUserSettings
	TelegramAccount

	GetAllUsers(ctx context.Context, page int32) ([]*models.User, error)
	GetUserByID(ctx context.Context, ID uuid.UUID) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	StoreUser(ctx context.Context, dto *CreateUserDTO, opts ...repos.Option) (*RegisterUserDTO, error)
	UpdateUser(ctx context.Context, usr *models.User, dto *UpdateUserDTO, opts ...repos.Option) (*models.User, error)
	RegisterUserInProcessing(ctx context.Context, user *models.User, mnemonic string, options ...repos.Option) (*processing.RegisterOwnerInfo, error)
	AcceptInvite(ctx context.Context, email, token, password string) error
	UpdateRate(ctx context.Context, user *models.User, source models.RateSource, scale decimal.Decimal) error
	InitOwnerRegistration(ctx context.Context, user *models.User) (link string, err error)
	GetOwnerDataFromAdmin(ctx context.Context, user *models.User) (OwnerData, error)
	GenerateTgLink(ctx context.Context, user *models.User) (string, error)
}

type Service struct {
	cfg                 *config.Config
	storage             storage.IStorage
	logger              logger.Logger
	storeService        store.IStore
	permissionService   permission.IPermission
	processingService   processing.IProcessingOwner
	notificationService notify.INotificationService
	settingsService     setting.ISettingService
	dvAuth              admin_gateway.IOwner
	otpSvc              *otp.Service
}

const PageSize = 10

var _ IUser = (*Service)(nil)

func New(
	cfg *config.Config,
	storage storage.IStorage,
	storeService store.IStore,
	permissionService permission.IPermission,
	processingService processing.IProcessingOwner,
	notificationService notify.INotificationService,
	logger logger.Logger,
	settingsService setting.ISettingService,
	dvAuth admin_gateway.IOwner,
	otpSvc *otp.Service,
) *Service {
	return &Service{
		cfg:                 cfg,
		storage:             storage,
		storeService:        storeService,
		permissionService:   permissionService,
		processingService:   processingService,
		notificationService: notificationService,
		logger:              logger,
		settingsService:     settingsService,
		dvAuth:              dvAuth,
		otpSvc:              otpSvc,
	}
}

func (s *Service) GetAllUsers(ctx context.Context, page int32) ([]*models.User, error) {
	if page < 1 {
		return nil, fmt.Errorf("page number must be greater than 0")
	}
	params := repo_users.GetAllParams{
		Limit:  PageSize,
		Offset: (page - 1) * PageSize,
	}

	users, err := s.storage.Users().GetAll(ctx, params)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (s *Service) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	user, err := s.storage.Users().GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Service) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	user, err := s.storage.Users().GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Service) StoreUser(ctx context.Context, dto *CreateUserDTO, opts ...repos.Option) (*RegisterUserDTO, error) {
	// Hash the password
	hashPassword, err := tools.HashPassword(dto.Password)
	if err != nil {
		return nil, fmt.Errorf("can't hash password: %w", err)
	}
	dto.Password = hashPassword

	// Validate parameters
	if err := dto.Validate(); err != nil {
		return nil, fmt.Errorf("validate params error: %w", err)
	}

	// Check if user already exists
	existingUser, err := s.GetUserByEmail(ctx, dto.Email)
	if err == nil && existingUser != nil {
		return nil, errors.New("user with this email already exists")
	}

	var rUserInfo = &RegisterUserDTO{}
	// Start transaction
	err = repos.BeginTxFunc(ctx, s.storage.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		// Create the user
		user, createErr := s.storage.Users(repos.WithTx(tx)).Create(ctx, dto.ToDbCreateParams())
		if createErr != nil {
			return createErr
		}

		// Create the store
		if _, createErr = s.storeService.CreateStore(ctx, store.CreateStore{Name: "First Store"}, user, repos.WithTx(tx)); createErr != nil {
			return createErr
		}
		rUserInfo.User = user

		return nil
	}, opts...)
	if err != nil {
		return nil, fmt.Errorf("store user error: %w", err)
	}

	// Add user role
	if _, err = s.permissionService.AddUserRole(rUserInfo.User.ID.String(), dto.Role); err != nil {
		return nil, err
	}

	return rUserInfo, nil
}

func (s *Service) RegisterUserInProcessing(
	ctx context.Context,
	user *models.User,
	mnemonic string,
	options ...repos.Option,
) (*processing.RegisterOwnerInfo, error) {
	if _, err := s.processingService.ProcessingSettings(ctx); err != nil {
		return nil, err
	}

	pOwnerInfo, err := s.processingService.CreateOwner(ctx, user.ID.String(), mnemonic)
	if err != nil {
		return nil, fmt.Errorf("create owner in processing: %w", err)
	}

	if _, err := s.UpdateUserProcessingID(ctx, pOwnerInfo.OwnerID, user.ID, options...); err != nil {
		return nil, fmt.Errorf("update user processing ID: %w", err)
	}

	return pOwnerInfo, nil
}

func (s *Service) UpdateUser(ctx context.Context, usr *models.User, dto *UpdateUserDTO, opts ...repos.Option) (*models.User, error) {
	params := dto.ToUpdateParams()

	if dto.ExchangeSlug != nil {
		params.ExchangeSlug = dto.ExchangeSlug
	} else {
		params.ExchangeSlug = usr.ExchangeSlug
	}
	params.ID = usr.ID

	user, err := s.storage.Users(opts...).Update(ctx, params)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Service) UpdateUserProcessingID(ctx context.Context, processingOwnerID uuid.UUID, userID uuid.UUID, opts ...repos.Option) (*models.User, error) {
	params := repo_users.UpdateProcessingOwnerIdParams{
		ProcessingOwnerID: uuid.NullUUID{UUID: processingOwnerID, Valid: true},
		ID:                userID,
	}
	user, err := s.storage.Users(opts...).UpdateProcessingOwnerId(ctx, params)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Service) UpdateRate(ctx context.Context, user *models.User, source models.RateSource, scale decimal.Decimal) error {
	err := repos.BeginTxFunc(ctx, s.storage.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		userParams := repo_users.UpdateRateParams{
			RateSource: source,
			RateScale:  scale,
			ID:         user.ID,
		}
		err := s.storage.Users(repos.WithTx(tx)).UpdateRate(ctx, userParams)
		if err != nil {
			return fmt.Errorf("failed to update user rate scale: %w", err)
		}
		storeParams := repo_stores.UpdateRateParams{
			RateSource: source,
			RateScale:  scale,
			UserID:     user.ID,
		}
		err = s.storage.Stores(repos.WithTx(tx)).UpdateRate(ctx, storeParams)
		if err != nil {
			return fmt.Errorf("failed to update user rate scale: %w", err)
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) AcceptInvite(ctx context.Context, email string, token string, password string) error {
	storageToken, err := s.storage.KeyValue().Get(ctx, email)
	if err != nil {
		return fmt.Errorf("failed to get user invite token: %w", err)
	}

	if storageToken.String() != token {
		return fmt.Errorf("failed to validate user invite token")
	}

	user, err := s.storage.Users().GetByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("failed to get user by email: %w", err)
	}

	hashPassword, err := tools.HashPassword(password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	if _, err := s.storage.Users().ChangePassword(ctx, repo_users.ChangePasswordParams{
		Password: hashPassword,
		ID:       user.ID,
	}); err != nil {
		return fmt.Errorf("failed to change user password: %w", err)
	}

	if err := s.storage.KeyValue().Delete(ctx, email); err != nil {
		return fmt.Errorf("failed to delete user invite token: %w", err)
	}

	return nil
}

func (s *Service) InitOwnerRegistration(ctx context.Context, user *models.User) (link string, err error) {
	if !user.ProcessingOwnerID.Valid {
		return "", ErrOwnerIDIsNotSet
	}

	ctx, err = s.prepareDvNetContext(ctx)
	if err != nil {
		return "", err
	}

	resp, err := s.dvAuth.GetAuthCode(ctx, admin_requests.InitAuthRequest{
		OwnerID: user.ProcessingOwnerID.UUID,
	})
	if err != nil {
		return "", err
	}

	if err = s.storage.Users().UpdateDvToken(ctx, repo_users.UpdateDvTokenParams{
		ID:         user.ID,
		DvnetToken: pgtype.Text{String: resp.Token, Valid: true},
	}); err != nil {
		return "", fmt.Errorf("failed to update user token: %w", err)
	}

	return resp.Link, nil
}

func (s *Service) GetOwnerDataFromAdmin(ctx context.Context, user *models.User) (OwnerData, error) {
	if !user.DvnetToken.Valid {
		return OwnerData{}, nil
	}

	ctx, err := s.prepareDvNetContext(ctx)
	if err != nil {
		return OwnerData{}, err
	}

	res, err := s.dvAuth.GetOwnerData(ctx, user.DvnetToken.String)
	if errors.Is(err, admin_gateway.ErrUnauthenticated) {
		return OwnerData{}, nil
	}

	if err != nil {
		return OwnerData{}, err
	}

	return OwnerData{
		IsAuthorized: true,
		Balance:      res.Balance,
		OwnerID:      user.ProcessingOwnerID.UUID.String(),
		Telegram:     res.Telegram,
	}, err
}

func (s *Service) sendUserEmailConfirmation(ctx context.Context, usr models.User, code int) {
	payload := &notify.UserVerificationData{
		Language: usr.Language,
		Code:     code,
	}

	go s.notificationService.SendUser(ctx, models.NotificationTypeUserVerification, &usr, payload, &models.NotificationArgs{UserID: &usr.ID})
}

func (s *Service) prepareDvNetContext(ctx context.Context) (context.Context, error) {
	clID, err := s.settingsService.GetRootSetting(ctx, setting.ProcessingClientID)
	if err != nil || clID == nil {
		return nil, ErrClientIDNotFound
	}

	adminSecret, err := s.settingsService.GetRootSetting(ctx, setting.DvAdminSecretKey)
	if err != nil || adminSecret == nil {
		// admin secret uninitialized
		return nil, ErrAdminSecretNotFound
	}

	return admin_gateway.PrepareServiceContext(ctx, adminSecret.Value, clID.Value), nil
}

func (s *Service) ensureDVAuthDataValid(user *models.User) error {
	if !user.DvnetToken.Valid {
		return ErrDvTokenNotSet
	}
	if !user.ProcessingOwnerID.Valid {
		return ErrOwnerIDIsNotSet
	}

	return nil
}
