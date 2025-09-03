package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func CheckHashPassword(password, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return err
	}
	return nil
}

func MakeJWT (userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{Issuer: "chirpy", IssuedAt: jwt.NewNumericDate(time.Now().UTC()), ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn).UTC()), Subject: userID.String()})
	signedToken, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", err
	}
	return signedToken, nil
}

func ValidateJWT (tokenString, tokenSecret string) (uuid.UUID, error) {
	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {return []byte(tokenSecret), nil})
	if err != nil {
		return uuid.Nil, err
	}
	if !token.Valid {
		return uuid.Nil, jwt.ErrTokenInvalidClaims
	}
	subject, err := claims.GetSubject()
	if err != nil {
		return uuid.Nil, err
	}
	userId, err := uuid.Parse(subject)
	if err != nil {
		return uuid.Nil, err
	}
	return userId, nil
}

func GetBearerToken (headers http.Header) (string, error) {
	header := headers.Get("Authorization")
	if header == "" || !strings.HasPrefix(header, "Bearer ") {
		return "", fmt.Errorf("invalid authorization header")
	}
	token := strings.TrimPrefix(header, "Bearer ")
	token = strings.TrimLeft(token, " ")
	token = strings.TrimRight(token, " ")
	return token, nil	
}

func MakeRefreshToken () (string, error) {
	token := make([]byte, 32)
	_, err := rand.Read(token)
	if err != nil {
		return "", err
	}
	refreshToken := hex.EncodeToString([]byte(token))
	return refreshToken, nil
}

func GetAPIKey(headers http.Header) (string, error) {
	header := headers.Get("Authorization")
	if header == "" || !strings.HasPrefix(header, "ApiKey ") {
		return "", fmt.Errorf("invalid authorization header")
	}
	token := strings.TrimPrefix(header, "ApiKey ")
	token = strings.TrimLeft(token, " ")
	token = strings.TrimRight(token, " ")
	return token, nil
}