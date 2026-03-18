package admin

import (
	"context"
	"github.com/PaPaSmUrFiK/MarketFlow/identity-service/internal/domain"
	svcadmin "github.com/PaPaSmUrFiK/MarketFlow/identity-service/internal/service/admin"
	identityv1 "github.com/PaPaSmUrFiK/MarketFlow/marketplace-proto/gen/go/identity/v1"
	"github.com/google/uuid"
	"google.golang.org/grpc"
)

type Service interface {
	// Application
	CreateApplication(ctx context.Context, in svcadmin.CreateApplicationInput) (*domain.Application, error)

	//Users
	CreateUser(ctx context.Context, in svcadmin.CreateUserInput) (*domain.User, error)
	BlockUser(ctx context.Context, userID uuid.UUID) error
	UnblockUser(ctx context.Context, userID uuid.UUID) error
	DeleteUser(ctx context.Context, userID uuid.UUID) error
	ListUsers(ctx context.Context, appID uuid.UUID, page, size int32) ([]domain.User, int64, error)
	GetUserDetails(ctx context.Context, userID uuid.UUID, appID uuid.UUID) (*domain.User, error)

	//Roles
	CreateRole(ctx context.Context, in svcadmin.CreateRoleInput) (*domain.Role, error)
	DeleteRole(ctx context.Context, appID uuid.UUID, roleCode string) error
	AssignRoleToUser(ctx context.Context, in svcadmin.AssignRoleToUserInput) error
	RemoveRoleFromUser(ctx context.Context, in svcadmin.RemoveRoleFromUserInput) error

	// Permissions
	CreatePermission(ctx context.Context, in svcadmin.CreatePermissionInput) (*domain.Permission, error)
	DeletePermission(ctx context.Context, appID uuid.UUID, permCode string) error
	AssignPermissionToRole(ctx context.Context, in svcadmin.AssignPermissionToRoleInput) error
	RemovePermissionFromRole(ctx context.Context, in svcadmin.RemovePermissionFromRoleInput) error
}

type adminServerAPI struct {
	identityv1.UnimplementedAdminServiceServer
	admin Service
}

func Register(srv *grpc.Server, admin Service) {
	identityv1.RegisterAdminServiceServer(srv, &adminServerAPI{admin: admin})
}
