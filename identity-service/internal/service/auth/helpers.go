package auth

import "github.com/PaPaSmUrFiK/MarketFlow/identity-service/internal/domain"

func extractRoleCodes(user *domain.User) []string {
	roleCodes := make([]string, 0, len(user.Roles))
	for _, r := range user.Roles {
		roleCodes = append(roleCodes, r.Code)
	}
	return roleCodes
}
