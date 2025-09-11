// util/jwtutil.go
package util

import (
	"crypto/ed25519"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
)

type Claims struct {
	OrgName      string          `json:"org_name"`
	AllowedHosts HostPermissions `json:"allowed_hosts"`
	ExpiresAt    int64           `json:"expires"`
	jwt.RegisteredClaims
}

//func GenerateJWT(orgName, domain, secretKey string) (string, int64, error) {
//	// Set the expiration time for the token (e.g., 7 days from now)
//	expirationTime := time.Now().AddDate(0, 0, 14).Unix()
//
//	// Create JWT claims with SomeKey and expiration time
//	claims := &Claims{
//		OrgName:      orgName,
//		ExpiresAt:    expirationTime,
//		AllowedHosts: domain,
//		RegisteredClaims: jwt.RegisteredClaims{
//			ExpiresAt: expirationTime,
//		},
//	}
//
//	// Create JWT token with the claims
//	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
//
//	// Sign the token with the secret key
//	tokenString, err := token.SignedString([]byte(secretKey))
//	if err != nil {
//		return "", 0, err
//	}
//
//	return tokenString, expirationTime, nil
//}

// DecodeToken verifies an EdDSA JWT with a given public key
func DecodeToken(tokenString string, publicKey string) (*Claims, error) {
	pubKeyBytes, err := decodePublicKey(publicKey)
	if err != nil {
		return nil, fmt.Errorf("invalid public key: %w", err)
	}

	if len(pubKeyBytes) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("invalid public key length: %d bytes", len(pubKeyBytes))
	}

	edPubKey := ed25519.PublicKey(pubKeyBytes)

	decodedToken, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodEd25519); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return edPubKey, nil
	})
	if err != nil {
		return nil, fmt.Errorf("error decoding token: %w", err)
	}

	claims, ok := decodedToken.Claims.(*Claims)
	if !ok || !decodedToken.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Manual expiry check for custom 'expires' field
	if claims.ExpiresAt > 0 && time.Now().Unix() > claims.ExpiresAt {
		return nil, fmt.Errorf("token expired at %d", claims.ExpiresAt)
	}

	return claims, nil
}

// IsValidToken checks if the token is valid and not expired
func IsValidToken(tokenString string, publicKey string) bool {
	claims, err := DecodeToken(tokenString, publicKey)
	if err != nil {
		log.DefaultLogger.Error(err.Error())
		return false
	}

	// 	log.DefaultLogger.Info("OrgName:", claims.OrgName)
	// 	log.DefaultLogger.Info("Now:", time.Unix(time.Now().Unix(), 0))
	// 	log.DefaultLogger.Info("ExpiresAt:", time.Unix(claims.ExpiresAt, 0))

	return time.Now().Unix() < claims.ExpiresAt
}
