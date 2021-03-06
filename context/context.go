package context

import (
	"context"
	"go_playground/models"
)

const (
	userKey privateKey = "user"
)

type privateKey string

func WithUser(ctx context.Context, user *models.User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

func User(ctx context.Context) *models.User {
	if tmp := ctx.Value(userKey); tmp != nil {
		if user, ok := tmp.(*models.User); ok {
			return user
		}
	}
	return nil
}
