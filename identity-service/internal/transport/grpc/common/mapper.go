package common

import (
	"github.com/PaPaSmUrFiK/MarketFlow/identity-service/internal/domain"
	identityv1 "github.com/PaPaSmUrFiK/MarketFlow/marketplace-proto/gen/go/identity/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func UserToProto(u *domain.User) *identityv1.User {
	return &identityv1.User{
		Id:        u.ID.String(),
		Status:    string(u.Status),
		CreatedAt: timestamppb.New(u.CreatedAt),
		UpdatedAt: timestamppb.New(u.UpdatedAt),
	}
}

func RoleToProto(r domain.Role) *identityv1.Role {
	return &identityv1.Role{
		Id:          r.ID.String(),
		AppId:       r.AppID.String(),
		Code:        r.Code,
		Description: r.Description,
	}
}

func PermissionToProto(p domain.Permission) *identityv1.Permission {
	return &identityv1.Permission{
		Id:          p.ID.String(),
		AppId:       p.AppID.String(),
		Code:        p.Code,
		Description: p.Description,
	}
}

func SessionToProto(s domain.Session) *identityv1.Session {
	return &identityv1.Session{
		Id:        s.ID.String(),
		AppId:     s.AppID.String(),
		UserAgent: s.UserAgent,
		IpAddress: s.IPAddress,
		CreatedAt: timestamppb.New(s.CreatedAt),
		ExpiresAt: timestamppb.New(s.ExpiresAt),
	}
}

func ApplicationToProto(a *domain.Application) *identityv1.Application {
	return &identityv1.Application{
		Id:        a.ID.String(),
		Code:      a.Code,
		Name:      a.Name,
		Active:    a.Active,
		CreatedAt: timestamppb.New(a.CreatedAt),
	}
}

// UserProfileToProto маппит пользователя с ролями и пермишнами в UserProfile.
// email не хранится в domain.User — его нужно передать отдельно.
func UserProfileToProto(u *domain.User) *identityv1.UserProfile {
	roles := make([]*identityv1.Role, 0, len(u.Roles))
	var permissions []*identityv1.Permission

	for _, r := range u.Roles {
		roles = append(roles, RoleToProto(r))
		for _, p := range r.Permissions {
			permissions = append(permissions, PermissionToProto(p))
		}
	}

	return &identityv1.UserProfile{
		Id:          u.ID.String(),
		Email:       u.Email,
		Status:      string(u.Status),
		Roles:       roles,
		Permissions: permissions,
		CreatedAt:   timestamppb.New(u.CreatedAt),
		UpdatedAt:   timestamppb.New(u.UpdatedAt),
	}
}
