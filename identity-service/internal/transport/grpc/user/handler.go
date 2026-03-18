package user

import (
	"context"
	"github.com/PaPaSmUrFiK/MarketFlow/identity-service/internal/domain"
	svcuser "github.com/PaPaSmUrFiK/MarketFlow/identity-service/internal/service/user"
	identityv1 "github.com/PaPaSmUrFiK/MarketFlow/marketplace-proto/gen/go/identity/v1"
	"github.com/google/uuid"
	"google.golang.org/grpc"
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
