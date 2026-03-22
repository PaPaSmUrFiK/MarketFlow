// Package sl предоставляет утилиты для логирования ошибок с контекстом операции.
package sl

import (
	"fmt"
	"log/slog"
)

// Err оборачивает ошибку, добавляя имя операции.
// Если err == nil, возвращает nil.
//
// Пример использования:
//
//	const op = "auth.Service.Login"
//	if err != nil {
//	    return sl.Err(op, err)
//	}
func Err(op string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", op, err)
}

func ErrAttr(err error) slog.Attr {
	return slog.String("error", err.Error())
}
