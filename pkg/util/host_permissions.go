package util

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type HostPermission struct {
	Domain    string
	ExpiresAt jwt.NumericDate
	IsPower   bool
}

type HostPermissions []HostPermission

func (hp *HostPermissions) UnmarshalJSON(data []byte) error {
	// decode flat array of alternating string/number/bool
	var flat []interface{}
	if err := json.Unmarshal(data, &flat); err != nil {
		return fmt.Errorf("failed to unmarshal allowed hosts: %w", err)
	}

	if len(flat)%3 != 0 {
		return fmt.Errorf("allowed_hosts array length must be multiple of 3")
	}

	*hp = make([]HostPermission, 0, len(flat)/3)

	for i := 0; i < len(flat); i += 3 {
		// domain
		domain, ok := flat[i].(string)
		if !ok {
			return fmt.Errorf("expected string at index %d, got %T", i, flat[i])
		}

		// expiresAt (number -> NumericDate)
		var expiresAt jwt.NumericDate
		switch v := flat[i+1].(type) {
		case float64:
			// JSON numbers are float64
			expiresAt = jwt.NumericDate{Time: time.Unix(int64(v), 0)}
		default:
			return fmt.Errorf("expected number at index %d, got %T", i+1, flat[i+1])
		}

		// isPower
		isPower, ok := flat[i+2].(bool)
		if !ok {
			return fmt.Errorf("expected bool at index %d, got %T", i+2, flat[i+2])
		}

		*hp = append(*hp, HostPermission{
			Domain:    domain,
			ExpiresAt: expiresAt,
			IsPower:   isPower,
		})
	}

	return nil
}

// decodePublicKey decodes URL-safe Base64 Ed25519 keys with optional padding
func decodePublicKey(pubKey string) ([]byte, error) {
	s := strings.TrimSpace(pubKey)
	if m := len(s) % 4; m != 0 {
		s += strings.Repeat("=", 4-m)
	}
	return base64.URLEncoding.DecodeString(s)
}
