package user

import (
	"context"
	"github.com/PaPaSmUrFiK/MarketFlow/identity-service/internal/domain"
	svcuser "github.com/PaPaSmUrFiK/MarketFlow/identity-service/internal/service/user"
	grpcerrors "github.com/PaPaSmUrFiK/MarketFlow/identity-service/internal/transport/grpc"
	"github.com/PaPaSmUrFiK/MarketFlow/identity-service/internal/transport/grpc/common"
	identityv1 "github.com/PaPaSmUrFiK/MarketFlow/marketplace-proto/gen/go/identity/v1"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	_ "time"
)

type Service interface {
	GetMe(ctx context.Context, userID uuid.UUID, appID uuid.UUID) (*domain.User, error)
	GetUserById(ctx context.Context, userID uuid.UUID, appID uuid.UUID) (*domain.User, error)
	ChangePassword(ctx context.Context, userID uuid.UUID, in svcuser.ChangePasswordInput) error
	ListSessions(ctx context.Context, userID uuid.UUID, appID uuid.UUID) ([]domain.Session, error)
	RevokeSession(ctx context.Context, userID uuid.UUID, sessionID uuid.UUID) error
	LinkIdentity(ctx context.Context, userID uuid.UUID, in svcuser.LinkIdentityInput) error
	UnlinkIdentity(ctx context.Context, userID uuid.UUID, provider string) error
}

type userServerAPI struct {
	identityv1.UnimplementedUserServiceServer
	user Service
}

func Register(srv *grpc.Server, user Service) {
	identityv1.RegisterUserServiceServer(srv, &userServerAPI{user: user})
}

func (s *userServerAPI) GetMe(ctx context.Context, req *identityv1.GetMeRequest) (*identityv1.GetMeResponse, error) {
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	appID, err := uuid.Parse(req.GetAppId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid app_id")
	}

	user, err := s.user.GetMe(ctx, userID, appID)
	if err != nil {
		return nil, grpcerrors.DomainErrToStatus(err)
	}

	return &identityv1.GetMeResponse{
		Profile: common.UserProfileToProto(user),
	}, nil
}

func (s *userServerAPI) GetUserById(ctx context.Context, req *identityv1.GetUserByIdRequest) (*identityv1.GetUserByIdResponse, error) {
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	appID, err := uuid.Parse(req.GetAppId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid app_id")
	}

	user, err := s.user.GetUserById(ctx, userID, appID)
	if err != nil {
		return nil, grpcerrors.DomainErrToStatus(err)
	}

	return &identityv1.GetUserByIdResponse{
		Profile: common.UserProfileToProto(user),
	}, nil
}

func (s *userServerAPI) ChangePassword(ctx context.Context, req *identityv1.ChangePasswordRequest) (*identityv1.ChangePasswordResponse, error) {
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	if req.GetOldPassword() == "" || req.GetNewPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "old_password and new_password are required")
	}

	if err := s.user.ChangePassword(ctx, userID, svcuser.ChangePasswordInput{
		OldPassword: req.GetOldPassword(),
		NewPassword: req.GetNewPassword(),
	}); err != nil {
		return nil, grpcerrors.DomainErrToStatus(err)
	}

	return &identityv1.ChangePasswordResponse{Success: true}, nil
}

func (s *userServerAPI) ListSessions(ctx context.Context, req *identityv1.ListSessionsRequest) (*identityv1.ListSessionsResponse, error) {
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	appID, err := uuid.Parse(req.GetAppId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid app_id")
	}

	sessions, err := s.user.ListSessions(ctx, userID, appID)
	if err != nil {
		return nil, grpcerrors.DomainErrToStatus(err)
	}

	protoSessions := make([]*identityv1.Session, 0, len(sessions))
	for _, sess := range sessions {
		protoSessions = append(protoSessions, common.SessionToProto(sess))
	}

	return &identityv1.ListSessionsResponse{Sessions: protoSessions}, nil
}

func (s *userServerAPI) RevokeSession(ctx context.Context, req *identityv1.RevokeSessionRequest) (*identityv1.RevokeSessionResponse, error) {
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	sessionID, err := uuid.Parse(req.GetSessionId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid session_id")
	}

	if err := s.user.RevokeSession(ctx, userID, sessionID); err != nil {
		return nil, grpcerrors.DomainErrToStatus(err)
	}

	return &identityv1.RevokeSessionResponse{Success: true}, nil
}

func (s *userServerAPI) LinkIdentity(ctx context.Context, req *identityv1.LinkIdentityRequest) (*identityv1.LinkIdentityResponse, error) {
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	if req.GetProvider() == "" || req.GetCode() == "" {
		return nil, status.Error(codes.InvalidArgument, "provider and code are required")
	}

	if err := s.user.LinkIdentity(ctx, userID, svcuser.LinkIdentityInput{
		Provider:    req.GetProvider(),
		Code:        req.GetCode(),
		RedirectURI: req.GetRedirectUri(),
	}); err != nil {
		return nil, grpcerrors.DomainErrToStatus(err)
	}

	return &identityv1.LinkIdentityResponse{Success: true}, nil
}

func (s *userServerAPI) UnlinkIdentity(ctx context.Context, req *identityv1.UnlinkIdentityRequest) (*identityv1.UnlinkIdentityResponse, error) {
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	if req.GetProvider() == "" {
		return nil, status.Error(codes.InvalidArgument, "provider is required")
	}

	if err := s.user.UnlinkIdentity(ctx, userID, req.GetProvider()); err != nil {
		return nil, grpcerrors.DomainErrToStatus(err)
	}

	return &identityv1.UnlinkIdentityResponse{Success: true}, nil
}
