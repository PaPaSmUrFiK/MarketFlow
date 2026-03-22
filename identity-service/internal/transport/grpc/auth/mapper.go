package auth

import (
	"github.com/PaPaSmUrFiK/MarketFlow/identity-service/internal/service/auth"
	identityv1 "github.com/PaPaSmUrFiK/MarketFlow/marketplace-proto/gen/go/identity/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

func tokenPairToProto(pair *auth.TokenPair, accessTTL, refreshTTL time.Duration) *identityv1.AuthTokens {
	now := time.Now()
	return &identityv1.AuthTokens{
		AccessToken:      pair.AccessToken,
		RefreshToken:     pair.RefreshToken,
		AccessExpiresAt:  timestamppb.New(now.Add(accessTTL)),
		RefreshExpiresAt: timestamppb.New(now.Add(refreshTTL)),
	}
}
