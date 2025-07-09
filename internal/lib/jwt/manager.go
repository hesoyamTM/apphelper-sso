package jwt

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/hesoyamTM/apphelper-sso/internal/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrUnauthorized = errors.New("unauthorized")
)

func NewTokens(user models.UserInfo, duration time.Duration, prKey *ecdsa.PrivateKey) (models.JWTokens, error) {
	token := jwt.New(jwt.SigningMethodES256)

	claims := token.Claims.(jwt.MapClaims)
	claims["uid"] = user.Id
	claims["name"] = user.Name
	claims["surname"] = user.Surname
	claims["exp"] = time.Now().Add(duration).Unix()

	tokenString, err := token.SignedString(prKey)
	if err != nil {
		return models.JWTokens{}, err
	}

	return models.JWTokens{
		AccessToken:  tokenString,
		RefreshToken: uuid.NewString(),
	}, nil
}

func VerifyBearerToken(bearerToken string, publicKey *ecdsa.PublicKey) (string, error) {
	const op = "jwt.VerifyBearerToken"

	token := strings.Split(bearerToken, " ")[1]

	parsed, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		return publicKey, nil
	})
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	claims := parsed.Claims.(jwt.MapClaims)

	if int64(claims["exp"].(float64)) < time.Now().Unix() {
		return "", fmt.Errorf("%s: %w", op, ErrUnauthorized)
	}

	return claims["uid"].(string), nil
}
