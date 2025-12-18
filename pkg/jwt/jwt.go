package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

//var secret = []byte("z9tld5hG07Mgz1wm995xeH4yKYOprz9NALqQj2bBDUs=")

type Claims struct {
	UserID uint   `json:"user_id"`
	OpenID string `json:"openid"`
	jwt.RegisteredClaims
}

func GenerateToken(secret string, userID uint, openid string) (string, error) {
	claims := Claims{
		UserID: userID,
		OpenID: openid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

func ParseToken(secret string, tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, jwt.ErrTokenInvalidClaims
}
