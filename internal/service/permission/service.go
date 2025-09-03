package permission

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/dv-net/dv-merchant/internal/config"
	"github.com/dv-net/dv-merchant/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
	pgadapter "github.com/pckhoi/casbin-pgx-adapter/v3"

	fiber_casbin "github.com/dv-net/dv-merchant/third_party/fiber_casbin"

	"github.com/casbin/casbin/v2"

	"slices"

	"github.com/gofiber/fiber/v3"
)

var ErrUserNotFoundInLocals = errors.New("user not found in fiber context")

const (
	DomainData string = "data"
)

type IPermission interface { //nolint:interfacebloat
	// ClearPolicy deletes all permission policies
	ClearPolicy()
	// LoadPolicies load permission policies from CSV-file
	LoadPolicies(policiesFilePath string) error
	// AllUsers returns list of all users in permission policies
	AllUsers() ([]string, error)
	// AllRoles returns list of all roles in permission policies
	AllRoles() ([]models.UserRole, error)
	// UserRoles returns list of all roles of given user
	UserRoles(id string) ([]models.UserRole, error)
	// RoleUsers returns list of all users for given role
	RoleUsers(role models.UserRole) ([]string, error)
	// AddUserRole add permission role to given user.
	// Returns false if the user already has the roles (aka not affected).
	AddUserRole(id string, role models.UserRole) (bool, error)
	// DeleteUserRole deletes permission role for given user.
	// Returns false if the user does not have the role (aka not affected).
	DeleteUserRole(id string, role models.UserRole) (bool, error)
	// FiberMiddleware return casbin fiber middleware
	FiberMiddleware(roles ...models.UserRole) fiber.Handler
	// Enforcer return casbin interface
	Enforcer() *casbin.Enforcer
	// IsRoot checks if the user is root
	IsRoot(id string) (bool, error)
}

func New(conf *config.Config, conn *pgxpool.Pool) (srv Service, err error) {
	if srv.adapter, err = pgadapter.NewAdapter(conf.Postgres.DSN(), pgadapter.WithConnectionPool(conn)); err != nil {
		err = fmt.Errorf("create casbin pg adapter failed: %w", err)
		return srv, err
	}
	if srv.enforcer, err = casbin.NewEnforcer(conf.RolesModelPath, srv.adapter); err != nil {
		err = fmt.Errorf("create casbin enforcer failed: %w", err)
		return srv, err
	}
	srv.fiberCasbin = fiber_casbin.New(fiber_casbin.Config{
		ModelFilePath: conf.RolesModelPath,
		PolicyAdapter: srv.adapter,
		Enforcer:      srv.enforcer,
		Lookup: func(c fiber.Ctx) (subject string) {
			if v, ok := c.Locals("user").(*models.User); ok {
				subject = v.ID.String()
			} else {
				err = ErrUserNotFoundInLocals
			}
			return subject
		},
	})
	return srv, err
}

type Service struct {
	adapter     *pgadapter.Adapter
	enforcer    *casbin.Enforcer
	fiberCasbin *fiber_casbin.Middleware
}

var _ IPermission = (*Service)(nil)

func (s Service) Enforcer() *casbin.Enforcer {
	return s.enforcer
}

func (s Service) FiberMiddleware(roles ...models.UserRole) fiber.Handler {
	roleStrings := make([]string, len(roles))
	for i, role := range roles {
		roleStrings[i] = role.String()
	}
	return s.fiberCasbin.RequiresRoles(roleStrings)
}

func (s Service) ClearPolicy() {
	s.enforcer.ClearPolicy()
}

func (s Service) LoadPolicies(policiesFilePath string) error {
	var (
		err error
		f   io.ReadCloser
	)
	if f, err = os.Open(policiesFilePath); err != nil {
		return fmt.Errorf("open permission policies file failed: %w", err)
	}

	defer func() { _ = f.Close() }()
	r := csv.NewReader(f)
	var record []string

	for record, err = r.Read(); err == nil; record, err = r.Read() {
		if len(record) < 3 {
			return fmt.Errorf("invalid permission policy: %v", record)
		}
		if _, err = s.enforcer.AddNamedPolicy(record[0], record[1:]); err != nil {
			return fmt.Errorf("add permission policy {%v} failed: %w", record, err)
		}
	}

	if !errors.Is(err, io.EOF) {
		return fmt.Errorf("parse permission policies file failed: %w", err)
	}

	return nil
}

func (s Service) AllUsers() ([]string, error) {
	return s.enforcer.GetAllUsersByDomain(DomainData)
}

func (s Service) AllRoles() ([]models.UserRole, error) {
	allRoles, err := s.enforcer.GetAllRoles()
	if err != nil {
		return nil, err
	}
	roles := make([]models.UserRole, 0, len(allRoles))
	for _, role := range allRoles {
		if !models.UserRole(role).Valid() {
			return nil, fmt.Errorf("invalid role from enforcer: %s", role)
		}
		roles = append(roles, models.UserRole(role))
	}
	return roles, nil
}

func (s Service) UserRoles(id string) ([]models.UserRole, error) {
	userRoles, err := s.enforcer.GetRolesForUser(id)
	if err != nil {
		return nil, err
	}
	roles := make([]models.UserRole, 0, len(userRoles))
	for _, role := range userRoles {
		if !models.UserRole(role).Valid() {
			return nil, fmt.Errorf("invalid user role from enforcer: %s", role)
		}
		roles = append(roles, models.UserRole(role))
	}
	return roles, nil
}

func (s Service) RoleUsers(role models.UserRole) ([]string, error) {
	return s.enforcer.GetUsersForRole(role.String(), DomainData)
}

func (s Service) AddUserRole(id string, role models.UserRole) (bool, error) {
	return s.enforcer.AddRolesForUser(id, []string{role.String()}, DomainData)
}

func (s Service) DeleteUserRole(id string, role models.UserRole) (bool, error) {
	return s.enforcer.DeleteRoleForUser(id, role.String(), DomainData)
}

func (s Service) IsRoot(id string) (bool, error) {
	roles, err := s.UserRoles(id)
	if err != nil {
		return false, err
	}
	if slices.Contains(roles, models.UserRoleRoot) {
		return true, nil
	}
	return false, nil
}
