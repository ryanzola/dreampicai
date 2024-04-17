package view

import (
	"context"
	"log/slog"
	"strconv"

	"dreampicai/types"
)

func AuthenticatedUser(ctx context.Context) types.AuthenticatedUser {
	user, ok := ctx.Value(types.UserContextKey).(types.AuthenticatedUser)
	if !ok {
		slog.Error("user not found in context")
		return types.AuthenticatedUser{}
	}

	return user
}

func String(i int) string {
	return strconv.Itoa(i)
}
