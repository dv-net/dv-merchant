package converters

import (
	"github.com/dv-net/dv-merchant/internal/delivery/http/responses/admin_response"
	"github.com/dv-net/dv-merchant/internal/delivery/http/responses/user_response"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/admin"
)

func FromUserModelToInfoResponse(o *models.User, quickStartGuideStatus string, roles ...models.UserRole) *user_response.GetUserInfoResponse {
	userResponse := &user_response.GetUserInfoResponse{
		ID:                    o.ID,
		Email:                 o.Email,
		Location:              o.Location,
		Language:              o.Language,
		RateSource:            o.RateSource.String(),
		RateScale:             o.RateScale,
		ProcessingOwnerID:     o.ProcessingOwnerID,
		Roles:                 roles,
		CreatedAt:             o.CreatedAt.Time,
		UpdatedAt:             o.UpdatedAt.Time,
		QuickStartGuideStatus: quickStartGuideStatus,
	}
	if o.EmailVerifiedAt.Valid {
		userResponse.EmailVerifiedAt = &o.EmailVerifiedAt.Time
	}
	return userResponse
}

func FromInvitedUserToInviteResponse(user *admin.InvitedUser) admin_response.InviteUserResponse {
	return admin_response.InviteUserResponse{
		UserID: user.NewUser.ID,
		Email:  user.NewUser.Email,
		Token:  user.Token,
	}
}
