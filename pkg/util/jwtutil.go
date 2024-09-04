// util/jwtutil.go
package util

import (
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
)

type Claims struct {
	OrgName    string `json:"orgName"`
	DomainName string `json:"domain"`
	ExpiresAt  int64  `json:"exp"`
	FeatLimit  int64  `json:"featLimit"`
	jwt.StandardClaims
}

func DecodeToken(tokenString, secretKey string) (*Claims, error) {
	// Decode the token back
	decodedToken, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("Error decoding token: %w", err)
	}

	// Check if the token is valid and not expired
	if claims, ok := decodedToken.Claims.(*Claims); ok && decodedToken.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("Invalid token")
}

// IsValidToken checks if the token is valid and not expired
func IsValidToken(tokenString, secretKey string) bool {
	claims, err := DecodeToken(tokenString, secretKey)
	if err != nil {
		log.DefaultLogger.Error(err.Error())
		return false
	}

	// 	log.DefaultLogger.Info("OrgName:", claims.OrgName)
	// 	log.DefaultLogger.Info("Now:", time.Unix(time.Now().Unix(), 0))
	// 	log.DefaultLogger.Info("ExpiresAt:", time.Unix(claims.ExpiresAt, 0))

	return time.Now().Unix() < claims.ExpiresAt
}
