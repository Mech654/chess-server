package auth

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func JoinHandler(w http.ResponseWriter, r *http.Request) {

	username := r.URL.Query().Get("username")
	if username == "" {
		username = "Guest_" + fmt.Sprint(time.Now().UnixNano())[10:]
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
	})

	tokenString, _ := token.SignedString([]byte(os.Getenv("TOKEN_KEY")))

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    tokenString,
		HttpOnly: true,
		Path:     "/",
	})
}

func GetUsernameFromToken(r *http.Request) (string, error) {
	cookie, err := r.Cookie("token")

	if err != nil {
		return "", fmt.Errorf("missing token cookie")
	}

	secret := os.Getenv("TOKEN_KEY")

	if secret == "" {
		return "", fmt.Errorf("missing TOKEN_KEY env var")
	}

	token, err := jwt.Parse(cookie.Value, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(secret), nil
	})

	if err != nil {
		return "", err
	}

	if !token.Valid {
		return "", fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("invalid token claims")
	}

	username, ok := claims["username"].(string)
	if !ok {
		return "", fmt.Errorf("username not found in token")
	}

	return username, nil
}
