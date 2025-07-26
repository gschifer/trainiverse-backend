package firebase

import (
	"context"
	"net/http"
	"strings"

	firebase "firebase.google.com/go/v4"
	auth "firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
)

type FirebaseInterface interface {
	GetUserID(r *http.Request) string
}

type Authenticator interface {
	VerifyIDToken(ctx context.Context, idToken string) (*auth.Token, error)
}

type FirebaseClient struct{}

var Auth Authenticator

type contextKey string

var keyForUserIDs = contextKey("userID")

func StartFirebase() {
	opt := option.WithCredentialsFile("internal/firebase/serviceAccountKey.json")

	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		panic("could not start firebase: " + err.Error())
	}

	client, err := app.Auth(context.Background())
	if err != nil {
		panic("could not initialize firebase auth client: " + err.Error())
	}

	Auth = client
}

func FirebaseAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error":"missing Authorization header"}`, http.StatusUnauthorized)
			return
		}

		idToken := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := Auth.VerifyIDToken(r.Context(), idToken)
		if err != nil {
			http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), keyForUserIDs, token.UID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func (firebaseService FirebaseClient) GetUserID(r *http.Request) string {
	userID, _ := r.Context().Value(keyForUserIDs).(string)

	return userID
}
