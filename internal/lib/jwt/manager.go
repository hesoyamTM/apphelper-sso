package jwt

import (
	"crypto/ecdsa"
	"sso/internal/models"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
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

func Verify(token string, publicKey *ecdsa.PublicKey) (int64, error) {
	parsed, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		return publicKey, nil
	})
	if err != nil {
		return 0, err
	}

	claims := parsed.Claims.(jwt.MapClaims)
	return claims["uid"].(int64), nil
}
