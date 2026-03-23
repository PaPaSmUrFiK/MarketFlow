package admin

import (
	"errors"
	"github.com/PaPaSmUrFiK/MarketFlow/identity-service/internal/domain"
)

func isNotFound(err error) bool {
	return err != nil && (errors.Is(err, domain.ErrRoleNotFound) ||
		errors.Is(err, domain.ErrUserNotFound) ||
		errors.Is(err, domain.ErrApplicationNotFound) ||
		errors.Is(err, domain.ErrPermissionNotFound))
}

func isAlreadyExists(err error) bool {
	return err != nil && errors.Is(err, domain.ErrEmailAlreadyExists)
}
