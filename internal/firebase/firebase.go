package firebase

import (
	"context"
	"net/http"
	"strings"

	firebase "firebase.google.com/go/v4"
	auth "firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
)

var AuthClient *auth.Client

type contextKey string

var keyForUserIds = contextKey("userID")

func StartFirebase() {
	opt := option.WithCredentialsFile("../internal/firebase/serviceAccountKey.json")

	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		panic("could not start firebase: " + err.Error())
	}

	AuthClient, err = app.Auth(context.Background())
	if err != nil {
		panic("could not initialize firebase auth client: " + err.Error())
	}
}

func FirebaseAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error":"missing Authorization header"}`, http.StatusUnauthorized)
			return
		}

		idToken := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := AuthClient.VerifyIDToken(r.Context(), idToken)
		if err != nil {
			http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), keyForUserIds, token.UID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}


func GetUserID(r *http.Request) string {
	userID, _ := r.Context().Value(keyForUserIds).(string)
	return userID
}

