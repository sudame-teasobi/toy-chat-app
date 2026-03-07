package middleware

import (
	"context"
	"fmt"
	"net/http"
)

type userIDCtxKey struct{}

func SetUserID(ctx context.Context, val string) context.Context {
	return context.WithValue(ctx, userIDCtxKey{}, val)
}

func GetUserID(ctx context.Context) (string, error) {
	val, ok := ctx.Value(userIDCtxKey{}).(string)
	if !ok {
		return "", fmt.Errorf("failed to get user ID from context")
	}

	return val, nil
}

// AuthMiddleware は認証情報をコンテキストに乗せる
// 学習用プロジェクトなので、HTTPヘッダに直接乗せたユーザーのIDを取得するだけ
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get("x-user-id")
		if userID == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := SetUserID(r.Context(), userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
