package data

import (
	"context"
	"strings"
)

func EmailExists(ctx context.Context, email string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}

	storeMu.RLock()
	defer storeMu.RUnlock()

	_, ok := usersByEmail[strings.ToLower(email)]
	return ok, nil
}
