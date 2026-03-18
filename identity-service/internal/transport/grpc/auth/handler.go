package auth

import (
	"context"
	"github.com/PaPaSmUrFiK/MarketFlow/identity-service/internal/jwt"
	"github.com/PaPaSmUrFiK/MarketFlow/identity-service/internal/service/auth"
	identityv1 "github.com/PaPaSmUrFiK/MarketFlow/marketplace-proto/gen/go/identity/v1"
	"github.com/google/uuid"
	"google.golang.org/grpc"
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

type authServerAPI struct {
	identityv1.UnimplementedAuthServiceServer
	auth Service
}

func Register(srv *grpc.Server, auth Service) {
	identityv1.RegisterAuthServiceServer(srv, &authServerAPI{auth: auth})
}
