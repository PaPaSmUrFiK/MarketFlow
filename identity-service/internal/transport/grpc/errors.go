package grpc

import (
	"errors"
	"github.com/PaPaSmUrFiK/MarketFlow/identity-service/internal/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// DomainErrToStatus конвертирует domain ошибку в gRPC статус.
// Важно: не раскрываем внутренние детали клиенту —
// все неизвестные ошибки превращаются в Internal.
func DomainErrToStatus(err error) error {
	switch {
	// 400 — невалидный запрос
	case errors.Is(err, domain.ErrEmailRequired),
		errors.Is(err, domain.ErrPasswordRequired),
		errors.Is(err, domain.ErrInvalidEmailFormat),
		errors.Is(err, domain.ErrPasswordTooWeak),
		errors.Is(err, domain.ErrAppCodeRequired),
		errors.Is(err, domain.ErrInvalidIPAddress):
		return status.Error(codes.InvalidArgument, err.Error())

	// 401 — неверные credentials
	case errors.Is(err, domain.ErrInvalidCredentials):
		return status.Error(codes.Unauthenticated, err.Error())

	// 401 — токен истёк (клиент должен сделать Refresh)
	case errors.Is(err, domain.ErrRefreshTokenExpired):
		return status.Error(codes.Unauthenticated, "token expired")

	// 403 — аккаунт заблокирован
	case errors.Is(err, domain.ErrUserBlocked),
		errors.Is(err, domain.ErrUserLocked):
		return status.Error(codes.PermissionDenied, err.Error())

	case errors.Is(err, domain.ErrPermissionDenied):
		return status.Error(codes.PermissionDenied, err.Error())

	// 404 — сущность не найдена
	case errors.Is(err, domain.ErrUserNotFound):
		return status.Error(codes.NotFound, "user not found")

	case errors.Is(err, domain.ErrApplicationNotFound):
		return status.Error(codes.NotFound, "application not found")

	case errors.Is(err, domain.ErrRoleNotFound):
		return status.Error(codes.NotFound, "role not found")

	case errors.Is(err, domain.ErrPermissionNotFound):
		return status.Error(codes.NotFound, "permission not found")

	case errors.Is(err, domain.ErrSessionNotFound),
		errors.Is(err, domain.ErrSessionRevoked):
		return status.Error(codes.NotFound, "session not found or revoked")

	// 409 — конфликт (дубликат)
	case errors.Is(err, domain.ErrEmailAlreadyExists):
		return status.Error(codes.AlreadyExists, "email already exists")

	case errors.Is(err, domain.ErrAppAlreadyExists):
		return status.Error(codes.AlreadyExists, "application already exists")

	case errors.Is(err, domain.ErrRoleOrPermissionNotFound):
		return status.Error(codes.NotFound, "role or permission not found")

	case errors.Is(err, domain.ErrPermissionAlreadyAssigned):
		return status.Error(codes.AlreadyExists, "permission already assigned")

	case errors.Is(err, domain.ErrIdentityAlreadyExists):
		return status.Error(codes.AlreadyExists, "identity already linked")

	case errors.Is(err, domain.ErrIdentityNotFound):
		return status.Error(codes.NotFound, "identity not found")

	// 501 — не реализовано
	case errors.Is(err, domain.ErrNotImplemented):
		return status.Error(codes.Unimplemented, "not implemented")

	// 500 — всё остальное скрываем
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
