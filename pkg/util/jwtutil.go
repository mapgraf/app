// util/jwtutil.go
package util

import (
	"crypto/ed25519"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	OrgName      string          `json:"org"`
	AllowedHosts HostPermissions `json:"hosts"`
	ExpiresAt    jwt.NumericDate `json:"exp"`
	jwt.RegisteredClaims
}

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
		return nil, fmt.Errorf("invalid signature token")
	}

	return claims, nil
}

func CheckHost(host string, claims *Claims) (HostPermission, error) {
	for _, h := range claims.AllowedHosts {
		if h.Domain == host {
			if !h.IsPower {
				return h, fmt.Errorf("host %s is not a Power host", host)
			}
			return h, nil
		}
	}

	return HostPermission{}, fmt.Errorf("host %s is not allowed", host)
}

func HasSomePowerHost(tokenString string, publicKey string) (bool, error) {
	claims, err := DecodeToken(tokenString, publicKey)
	if err != nil {
		return false, fmt.Errorf("failed to decode token: %w", err)
	}

	//if claims.ExpiresAt.Time.Before(time.Now()) {
	//	return false, fmt.Errorf("token globally expired at %v", claims.ExpiresAt.Time)
	//}

	for _, h := range claims.AllowedHosts {
		if h.IsPower && h.ExpiresAt.Time.After(time.Now()) {
			return true, nil
		}
	}

	return false, nil
}
