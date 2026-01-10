package jwt

import (
	"Hyper/pkg/log"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

//var secret = []byte("z9tld5hG07Mgz1wm995xeH4yKYOprz9NALqQj2bBDUs=")

type Claims struct {
	UserID uint   `json:"user_id"`
	OpenID string `json:"openid"`
	Type   string `json:"type"`
	jwt.RegisteredClaims
}

func ShouldRotateRefreshToken(claims *Claims, refreshBuffer time.Duration) bool {
	if claims.ExpiresAt == nil {
		return false
	}
	return time.Until(claims.ExpiresAt.Time) <= refreshBuffer
}

func GenerateToken(secret []byte, userID uint, openid string, tokenType string, expire time.Duration) (string, error) {
	claims := Claims{
		UserID: userID,
		OpenID: openid,
		Type:   tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expire)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

//func ParseToken(secret []byte, tokenStr string) (*Claims, error) {
//	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
//		return secret, nil
//	})
//	if err != nil {
//		return nil, err
//	}
//	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
//		return claims, nil
//	}
//	return nil, jwt.ErrTokenInvalidClaims
//}

func ParseToken(secret []byte, expectedType string, tokenStr string) (*Claims, error) {

	fmt.Println(string(secret), tokenStr, expectedType, 55)
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return secret, nil
	})

	if err != nil {
		fmt.Println(err, 66)
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}
	log.L.Info("token parsed", zap.Any("claims", claims))

	if claims.Type != expectedType {
		return nil, errors.New("invalid token type")
	}

	return claims, nil
}
