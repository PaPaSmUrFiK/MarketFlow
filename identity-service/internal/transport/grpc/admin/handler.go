package admin

import (
	"context"
	"github.com/PaPaSmUrFiK/MarketFlow/identity-service/internal/domain"
	svcadmin "github.com/PaPaSmUrFiK/MarketFlow/identity-service/internal/service/admin"
	grpcerrors "github.com/PaPaSmUrFiK/MarketFlow/identity-service/internal/transport/grpc"
	"github.com/PaPaSmUrFiK/MarketFlow/identity-service/internal/transport/grpc/common"
	identityv1 "github.com/PaPaSmUrFiK/MarketFlow/marketplace-proto/gen/go/identity/v1"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

func (s *adminServerAPI) CreateApplication(ctx context.Context, req *identityv1.CreateApplicationRequest) (*identityv1.CreateApplicationResponse, error) {
	if req.GetCode() == "" || req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "code and name are required")
	}

	app, err := s.admin.CreateApplication(ctx, svcadmin.CreateApplicationInput{
		Code: req.GetCode(),
		Name: req.GetName(),
	})
	if err != nil {
		return nil, grpcerrors.DomainErrToStatus(err)
	}

	return &identityv1.CreateApplicationResponse{
		Application: common.ApplicationToProto(app),
	}, nil
}

func (s *adminServerAPI) CreateUser(ctx context.Context, req *identityv1.CreateUserRequest) (*identityv1.CreateUserResponse, error) {
	appID, err := uuid.Parse(req.GetAppId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid app_id")
	}

	if req.GetEmail() == "" || req.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "email and password are required")
	}

	user, err := s.admin.CreateUser(ctx, svcadmin.CreateUserInput{
		AppID:    appID,
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
	})
	if err != nil {
		return nil, grpcerrors.DomainErrToStatus(err)
	}

	return &identityv1.CreateUserResponse{
		User: common.UserToProto(user),
	}, nil
}

func (s *adminServerAPI) BlockUser(ctx context.Context, req *identityv1.BlockUserRequest) (*identityv1.BlockUserResponse, error) {
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	if err := s.admin.BlockUser(ctx, userID); err != nil {
		return nil, grpcerrors.DomainErrToStatus(err)
	}

	return &identityv1.BlockUserResponse{Success: true}, nil
}

func (s *adminServerAPI) UnblockUser(ctx context.Context, req *identityv1.UnblockUserRequest) (*identityv1.UnblockUserResponse, error) {
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	if err := s.admin.UnblockUser(ctx, userID); err != nil {
		return nil, grpcerrors.DomainErrToStatus(err)
	}

	return &identityv1.UnblockUserResponse{Success: true}, nil
}

func (s *adminServerAPI) DeleteUser(ctx context.Context, req *identityv1.DeleteUserRequest) (*identityv1.DeleteUserResponse, error) {
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	if err := s.admin.DeleteUser(ctx, userID); err != nil {
		return nil, grpcerrors.DomainErrToStatus(err)
	}

	return &identityv1.DeleteUserResponse{Success: true}, nil
}

func (s *adminServerAPI) ListUsers(ctx context.Context, req *identityv1.ListUsersRequest) (*identityv1.ListUsersResponse, error) {
	appID, err := uuid.Parse(req.GetAppId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid app_id")
	}

	page := int32(1)
	size := int32(20)
	if p := req.GetPagination(); p != nil {
		if p.GetPage() > 0 {
			page = p.GetPage()
		}
		if p.GetSize() > 0 {
			size = p.GetSize()
		}
	}

	users, total, err := s.admin.ListUsers(ctx, appID, page, size)
	if err != nil {
		return nil, grpcerrors.DomainErrToStatus(err)
	}

	protoUsers := make([]*identityv1.User, 0, len(users))
	for _, u := range users {
		u := u
		protoUsers = append(protoUsers, common.UserToProto(&u))
	}

	return &identityv1.ListUsersResponse{
		Users:    protoUsers,
		PageInfo: &identityv1.PageInfo{Total: total},
	}, nil
}

func (s *adminServerAPI) GetUserDetails(ctx context.Context, req *identityv1.GetUserDetailsRequest) (*identityv1.GetUserDetailsResponse, error) {
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	appID, err := uuid.Parse(req.GetAppId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid app_id")
	}

	user, err := s.admin.GetUserDetails(ctx, userID, appID)
	if err != nil {
		return nil, grpcerrors.DomainErrToStatus(err)
	}

	return &identityv1.GetUserDetailsResponse{
		Profile: common.UserProfileToProto(user),
	}, nil
}

func (s *adminServerAPI) CreateRole(ctx context.Context, req *identityv1.CreateRoleRequest) (*identityv1.CreateRoleResponse, error) {
	appID, err := uuid.Parse(req.GetAppId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid app_id")
	}

	if req.GetCode() == "" {
		return nil, status.Error(codes.InvalidArgument, "code is required")
	}

	role, err := s.admin.CreateRole(ctx, svcadmin.CreateRoleInput{
		AppID:       appID,
		Code:        req.GetCode(),
		Description: req.GetDescription(),
	})
	if err != nil {
		return nil, grpcerrors.DomainErrToStatus(err)
	}

	return &identityv1.CreateRoleResponse{
		Role: common.RoleToProto(*role),
	}, nil
}

func (s *adminServerAPI) DeleteRole(ctx context.Context, req *identityv1.DeleteRoleRequest) (*identityv1.DeleteRoleResponse, error) {
	appID, err := uuid.Parse(req.GetAppId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid app_id")
	}

	if req.GetRoleCode() == "" {
		return nil, status.Error(codes.InvalidArgument, "role_code is required")
	}

	if err := s.admin.DeleteRole(ctx, appID, req.GetRoleCode()); err != nil {
		return nil, grpcerrors.DomainErrToStatus(err)
	}

	return &identityv1.DeleteRoleResponse{Success: true}, nil
}

func (s *adminServerAPI) AssignRoleToUser(ctx context.Context, req *identityv1.AssignRoleToUserRequest) (*identityv1.AssignRoleToUserResponse, error) {
	appID, err := uuid.Parse(req.GetAppId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid app_id")
	}

	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	if req.GetRoleCode() == "" {
		return nil, status.Error(codes.InvalidArgument, "role_code is required")
	}

	if err := s.admin.AssignRoleToUser(ctx, svcadmin.AssignRoleToUserInput{
		AppID:    appID,
		UserID:   userID,
		RoleCode: req.GetRoleCode(),
	}); err != nil {
		return nil, grpcerrors.DomainErrToStatus(err)
	}

	return &identityv1.AssignRoleToUserResponse{Success: true}, nil
}

func (s *adminServerAPI) RemoveRoleFromUser(ctx context.Context, req *identityv1.RemoveRoleFromUserRequest) (*identityv1.RemoveRoleFromUserResponse, error) {
	appID, err := uuid.Parse(req.GetAppId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid app_id")
	}

	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	if req.GetRoleCode() == "" {
		return nil, status.Error(codes.InvalidArgument, "role_code is required")
	}

	if err := s.admin.RemoveRoleFromUser(ctx, svcadmin.RemoveRoleFromUserInput{
		AppID:    appID,
		UserID:   userID,
		RoleCode: req.GetRoleCode(),
	}); err != nil {
		return nil, grpcerrors.DomainErrToStatus(err)
	}

	return &identityv1.RemoveRoleFromUserResponse{Success: true}, nil
}

func (s *adminServerAPI) CreatePermission(ctx context.Context, req *identityv1.CreatePermissionRequest) (*identityv1.CreatePermissionResponse, error) {
	appID, err := uuid.Parse(req.GetAppId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid app_id")
	}

	if req.GetCode() == "" {
		return nil, status.Error(codes.InvalidArgument, "code is required")
	}

	perm, err := s.admin.CreatePermission(ctx, svcadmin.CreatePermissionInput{
		AppID:       appID,
		Code:        req.GetCode(),
		Description: req.GetDescription(),
	})
	if err != nil {
		return nil, grpcerrors.DomainErrToStatus(err)
	}

	return &identityv1.CreatePermissionResponse{
		Permission: common.PermissionToProto(*perm),
	}, nil
}

func (s *adminServerAPI) DeletePermission(ctx context.Context, req *identityv1.DeletePermissionRequest) (*identityv1.DeletePermissionResponse, error) {
	appID, err := uuid.Parse(req.GetAppId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid app_id")
	}

	if req.GetPermissionCode() == "" {
		return nil, status.Error(codes.InvalidArgument, "permission_code is required")
	}

	if err := s.admin.DeletePermission(ctx, appID, req.GetPermissionCode()); err != nil {
		return nil, grpcerrors.DomainErrToStatus(err)
	}

	return &identityv1.DeletePermissionResponse{Success: true}, nil
}

func (s *adminServerAPI) AssignPermissionToRole(ctx context.Context, req *identityv1.AssignPermissionToRoleRequest) (*identityv1.AssignPermissionToRoleResponse, error) {
	appID, err := uuid.Parse(req.GetAppId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid app_id")
	}

	if req.GetRoleCode() == "" || req.GetPermissionCode() == "" {
		return nil, status.Error(codes.InvalidArgument, "role_code and permission_code are required")
	}

	if err := s.admin.AssignPermissionToRole(ctx, svcadmin.AssignPermissionToRoleInput{
		AppID:          appID,
		RoleCode:       req.GetRoleCode(),
		PermissionCode: req.GetPermissionCode(),
	}); err != nil {
		return nil, grpcerrors.DomainErrToStatus(err)
	}

	return &identityv1.AssignPermissionToRoleResponse{Success: true}, nil
}

func (s *adminServerAPI) RemovePermissionFromRole(ctx context.Context, req *identityv1.RemovePermissionFromRoleRequest) (*identityv1.RemovePermissionFromRoleResponse, error) {
	appID, err := uuid.Parse(req.GetAppId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid app_id")
	}

	if req.GetRoleCode() == "" || req.GetPermissionCode() == "" {
		return nil, status.Error(codes.InvalidArgument, "role_code and permission_code are required")
	}

	if err := s.admin.RemovePermissionFromRole(ctx, svcadmin.RemovePermissionFromRoleInput{
		AppID:          appID,
		RoleCode:       req.GetRoleCode(),
		PermissionCode: req.GetPermissionCode(),
	}); err != nil {
		return nil, grpcerrors.DomainErrToStatus(err)
	}

	return &identityv1.RemovePermissionFromRoleResponse{Success: true}, nil
}
