package auth

import (
	"context"
	"github.com/PaPaSmUrFiK/MarketFlow/identity-service/internal/jwt"
	"github.com/PaPaSmUrFiK/MarketFlow/identity-service/internal/service/auth"
	grpcerrors "github.com/PaPaSmUrFiK/MarketFlow/identity-service/internal/transport/grpc"
	identityv1 "github.com/PaPaSmUrFiK/MarketFlow/marketplace-proto/gen/go/identity/v1"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

// Auth defines the contract for the transport layer.
// Used for dependency injection in handlers.
type Service interface {
	RegisterNewUser(ctx context.Context, in auth.RegisterInput) (*auth.TokenPair, error)
	Login(ctx context.Context, in auth.LoginInput) (*auth.TokenPair, error)
	LoginWithOAuth(ctx context.Context, in auth.OAuthLoginInput) (*auth.TokenPair, error)
	Refresh(ctx context.Context, rawRefresh string) (*auth.TokenPair, error)
	Logout(ctx context.Context, sessionID uuid.UUID) error
	LogoutAll(ctx context.Context, userID, appID uuid.UUID) error
	ValidateAccessToken(ctx context.Context, token string) (*jwt.AccessClaims, error)
}

type jwtTTLProvider interface {
	AccessTTL() time.Duration
	RefreshTTL() time.Duration
}

type authServerAPI struct {
	identityv1.UnimplementedAuthServiceServer
	auth Service
	jwt  jwtTTLProvider
}

func Register(srv *grpc.Server, auth Service, jwtManager jwtTTLProvider) {
	identityv1.RegisterAuthServiceServer(srv, &authServerAPI{
		auth: auth,
		jwt:  jwtManager,
	})
}

func (s *authServerAPI) Register(ctx context.Context, req *identityv1.RegisterRequest) (*identityv1.RegisterResponse, error) {
	if req.GetAppId() == "" {
		return nil, status.Error(codes.InvalidArgument, "app_id is required")
	}

	pair, err := s.auth.RegisterNewUser(ctx, auth.RegisterInput{
		AppCode:   req.GetAppId(),
		Email:     req.GetEmail(),
		Password:  req.GetPassword(),
		UserAgent: req.GetUserAgent(),
		IPAddress: req.GetIpAddress(),
	})
	if err != nil {
		return nil, grpcerrors.DomainErrToStatus(err)
	}

	return &identityv1.RegisterResponse{
		Tokens: tokenPairToProto(pair, s.jwt.AccessTTL(), s.jwt.RefreshTTL()),
	}, nil
}

func (s *authServerAPI) Login(ctx context.Context, req *identityv1.LoginRequest) (*identityv1.LoginResponse, error) {
	if req.GetAppId() == "" {
		return nil, status.Error(codes.InvalidArgument, "app_id is required")
	}

	pair, err := s.auth.Login(ctx, auth.LoginInput{
		AppCode:   req.GetAppId(),
		Email:     req.GetEmail(),
		Password:  req.GetPassword(),
		UserAgent: req.GetUserAgent(),
		IPAddress: req.GetIpAddress(),
	})
	if err != nil {
		return nil, grpcerrors.DomainErrToStatus(err)
	}

	return &identityv1.LoginResponse{
		Tokens: tokenPairToProto(pair, s.jwt.AccessTTL(), s.jwt.RefreshTTL()),
	}, nil
}

func (s *authServerAPI) OAuthLogin(ctx context.Context, req *identityv1.OAuthLoginRequest) (*identityv1.OAuthLoginResponse, error) {
	if req.GetAppId() == "" || req.GetProvider() == "" || req.GetCode() == "" {
		return nil, status.Error(codes.InvalidArgument, "app_id, provider and code are required")
	}

	pair, err := s.auth.LoginWithOAuth(ctx, auth.OAuthLoginInput{
		AppCode:     req.GetAppId(),
		Provider:    req.GetProvider(),
		Code:        req.GetCode(),
		RedirectURI: req.GetRedirectUri(),
		UserAgent:   req.GetUserAgent(),
		IPAddress:   req.GetIpAddress(),
	})
	if err != nil {
		return nil, grpcerrors.DomainErrToStatus(err)
	}

	return &identityv1.OAuthLoginResponse{
		Tokens: tokenPairToProto(pair, s.jwt.AccessTTL(), s.jwt.RefreshTTL()),
	}, nil
}

func (s *authServerAPI) Refresh(ctx context.Context, req *identityv1.RefreshRequest) (*identityv1.RefreshResponse, error) {
	if req.GetRefreshToken() == "" {
		return nil, status.Error(codes.InvalidArgument, "refresh_token is required")
	}

	pair, err := s.auth.Refresh(ctx, req.GetRefreshToken())
	if err != nil {
		return nil, grpcerrors.DomainErrToStatus(err)
	}

	return &identityv1.RefreshResponse{
		Tokens: tokenPairToProto(pair, s.jwt.AccessTTL(), s.jwt.RefreshTTL()),
	}, nil
}

func (s *authServerAPI) Logout(ctx context.Context, req *identityv1.LogoutRequest) (*identityv1.LogoutResponse, error) {
	sessionID, err := uuid.Parse(req.GetSessionId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid session_id")
	}

	if err := s.auth.Logout(ctx, sessionID); err != nil {
		return nil, grpcerrors.DomainErrToStatus(err)
	}

	return &identityv1.LogoutResponse{Success: true}, nil
}

func (s *authServerAPI) LogoutAll(ctx context.Context, req *identityv1.LogoutAllRequest) (*identityv1.LogoutAllResponse, error) {
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	appID, err := uuid.Parse(req.GetAppId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid app_id")
	}

	if err := s.auth.LogoutAll(ctx, userID, appID); err != nil {
		return nil, grpcerrors.DomainErrToStatus(err)
	}

	return &identityv1.LogoutAllResponse{Success: true}, nil
}

func (s *authServerAPI) Validate(ctx context.Context, req *identityv1.ValidateRequest) (*identityv1.ValidateResponse, error) {
	if req.GetAccessToken() == "" {
		return nil, status.Error(codes.InvalidArgument, "access_token is required")
	}

	claims, err := s.auth.ValidateAccessToken(ctx, req.GetAccessToken())
	if err != nil {
		return nil, grpcerrors.DomainErrToStatus(err)
	}

	return &identityv1.ValidateResponse{
		UserId: claims.UserID.String(),
		AppId:  claims.AppID.String(),
		Roles:  claims.Roles,
	}, nil
}
