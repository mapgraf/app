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
	ExpiresAt    jwt.NumericDate `json:"expiresAt"`
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

	return claims, nil
}

func CheckHost(host string, claims *Claims) (HostPermission, error) {
	for _, h := range claims.AllowedHosts {
		if h.Domain == host {
			if !h.IsPower {
				return h, fmt.Errorf("host %s is not a power host", host)
			}
			// Check host-specific expiry
			if h.ExpiresAt.Time.Before(time.Now()) {
				return h, fmt.Errorf("host %s permission expired at %d", host, h.ExpiresAt.Time.Unix())
			}
			// Host is allowed, has power, and not expired
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

	if claims.ExpiresAt.Time.Before(time.Now()) {
		return false, fmt.Errorf("token expired at %v", claims.ExpiresAt.Time)
	}

	for _, h := range claims.AllowedHosts {
		if h.IsPower && h.ExpiresAt.Time.After(time.Now()) {
			return true, nil
		}
	}

	return false, nil
}
