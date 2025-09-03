package admin

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/dv-net/dv-merchant/internal/service/notify"
	"github.com/dv-net/dv-merchant/internal/service/user"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_user_stores"

	"golang.org/x/text/language"

	"github.com/dv-net/dv-merchant/internal/config"
	"github.com/dv-net/dv-merchant/internal/delivery/http/request/admin_request"
	"github.com/dv-net/dv-merchant/internal/delivery/http/responses/admin_response"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/permission"
	"github.com/dv-net/dv-merchant/internal/storage"
	"github.com/dv-net/dv-merchant/internal/storage/repos"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_users"
	"github.com/dv-net/dv-merchant/internal/storage/storecmn"
	"github.com/dv-net/dv-merchant/internal/tools"
	"github.com/dv-net/dv-merchant/internal/tools/apierror"
	"github.com/dv-net/dv-merchant/internal/tools/str"
	"github.com/dv-net/dv-merchant/pkg/logger"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type IAdmin interface {
	GetAllUsersFiltered(ctx context.Context, req admin_request.GetUsersRequest) (*storecmn.FindResponseWithFullPagination[*admin_response.GetUsersResponse], error)
	BanUserByID(ctx context.Context, userID uuid.UUID) (*admin_response.BanUserResponse, error)
	UnbanUserByID(ctx context.Context, userID uuid.UUID) (*admin_response.UnbanUserResponse, error)
	InviteUserWithRole(ctx context.Context, request *admin_request.InviteUserWithRoleRequest) (*InvitedUser, error)
}

type Service struct {
	cfg               *config.Config
	storage           storage.IStorage
	logger            logger.Logger
	permissionService permission.IPermission
	userService       user.IUser
	notificationSvc   notify.INotificationService
}

var _ IAdmin = (*Service)(nil)

func New(
	cfg *config.Config,
	storage storage.IStorage,
	logger logger.Logger,
	permissionService permission.IPermission,
	userService user.IUser,
	notificationSvc notify.INotificationService,
) *Service {
	return &Service{
		cfg:               cfg,
		storage:           storage,
		logger:            logger,
		permissionService: permissionService,
		userService:       userService,
		notificationSvc:   notificationSvc,
	}
}

func (o *Service) GetAllUsersFiltered(ctx context.Context, req admin_request.GetUsersRequest) (*storecmn.FindResponseWithFullPagination[*admin_response.GetUsersResponse], error) {
	commonParams := storecmn.NewCommonFindParams()

	if req.PageSize != nil {
		commonParams.SetPageSize(req.PageSize)
	}
	if req.Page != nil {
		commonParams.SetPage(req.Page)
	}

	params := repo_users.GetAllFilteredParams{}

	if req.Roles != nil {
		roles := strings.Split(*req.Roles, ",")
		params.Roles = roles
	}

	users, err := o.storage.Users().GetAllFiltered(ctx, params)
	if err != nil {
		return nil, err
	}

	dto := make([]*admin_response.GetUsersResponse, 0, len(users.Items))

	for _, user := range users.Items {
		u := &admin_response.GetUsersResponse{
			Email:     user.Email,
			UserID:    user.ID,
			CreatedAt: user.CreatedAt.Time,
			Banned:    user.Banned.Bool,
		}
		if user.UserRoles.Valid {
			for _, role := range user.UserRoles.Elements {
				u.Roles = append(u.Roles, role.String)
			}
		}
		dto = append(dto, u)
	}

	return &storecmn.FindResponseWithFullPagination[*admin_response.GetUsersResponse]{
		Items:      dto,
		Pagination: users.Pagination,
	}, nil
}

func (o *Service) BanUserByID(ctx context.Context, userID uuid.UUID) (*admin_response.BanUserResponse, error) {
	roles, err := o.permissionService.UserRoles(userID.String())
	if err != nil {
		return nil, apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	if slices.Contains(roles, models.UserRoleRoot) {
		return nil, apierror.New().AddError(errors.New("banning root is forbidden")).SetHttpCode(fiber.StatusForbidden)
	}

	args := repo_users.UpdateBannedParams{
		ID:     userID,
		Banned: pgtype.Bool{Bool: true, Valid: true},
	}

	patch, err := o.storage.Users().UpdateBanned(ctx, args)
	if err != nil {
		return nil, err
	}

	return &admin_response.BanUserResponse{
		UserID: patch.ID,
		Banned: patch.Banned.Bool,
	}, nil
}

func (o *Service) UnbanUserByID(ctx context.Context, userID uuid.UUID) (*admin_response.UnbanUserResponse, error) {
	args := repo_users.UpdateBannedParams{
		ID:     userID,
		Banned: pgtype.Bool{Bool: false, Valid: true},
	}

	patch, err := o.storage.Users().UpdateBanned(ctx, args)
	if err != nil {
		return nil, err
	}

	return &admin_response.UnbanUserResponse{
		UserID: patch.ID,
		Banned: patch.Banned.Bool,
	}, nil
}

type InvitedUser struct {
	NewUser *models.User
	Token   uuid.UUID
}

func (o *Service) InviteUserWithRole(ctx context.Context, req *admin_request.InviteUserWithRoleRequest) (*InvitedUser, error) {
	res := &InvitedUser{}
	err := repos.BeginTxFunc(ctx, o.storage.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		_, err := o.storage.Users(repos.WithTx(tx)).GetByEmail(ctx, req.Email)
		if err == nil {
			return fmt.Errorf("user with this email already exists")
		}
		if err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				return err
			}
		}

		randomPassword, err := str.RandomString(32)
		if err != nil {
			return err
		}

		hashPassword, err := tools.HashPassword(randomPassword)
		if err != nil {
			return fmt.Errorf("can't hash password: %w", err)
		}

		regInfo, err := o.userService.StoreUser(ctx, &user.CreateUserDTO{
			Email:      req.Email,
			Language:   language.English.String(),
			Password:   hashPassword,
			RateSource: models.RateSourceBinance,
			Location:   "America/New_York",
			Mnemonic:   req.Mnemonic,
		}, repos.WithTx(tx))
		if err != nil {
			return err
		}

		for _, storeID := range req.StoreIDs {
			_, err = o.storage.UserStores(repos.WithTx(tx)).Create(ctx, repo_user_stores.CreateParams{
				UserID:  regInfo.User.ID,
				StoreID: storeID,
			})
			if err != nil {
				return err
			}
		}

		token, err := uuid.NewUUID()
		if err != nil {
			return err
		}

		err = o.storage.KeyValue().Set(ctx, regInfo.User.Email, token.String(), time.Hour*24)
		if err != nil {
			return err
		}

		res = &InvitedUser{
			NewUser: regInfo.User,
			Token:   token,
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}
