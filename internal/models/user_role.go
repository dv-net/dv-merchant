package models

type UserRole string // @name UserRole

func (o UserRole) String() string { return string(o) }

var validRoles = map[UserRole]struct{}{
	UserRoleRoot:           {},
	UserRoleDefault:        {},
	UserRoleSupport:        {},
	UserRoleFinanceManager: {},
}

func (o UserRole) Valid() bool {
	_, exists := validRoles[o]
	return exists
}

const (
	UserRoleDefault        UserRole = "user"
	UserRoleRoot           UserRole = "root"
	UserRoleSupport        UserRole = "support"
	UserRoleFinanceManager UserRole = "finance_manager"
)
