package util

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

type HostPermission struct {
	Domain string
	IsFull bool
}

type HostPermissions []HostPermission

func (hp *HostPermissions) UnmarshalJSON(data []byte) error {
	// decode flat array of alternating string/bool
	var flat []interface{}
	if err := json.Unmarshal(data, &flat); err != nil {
		return fmt.Errorf("failed to unmarshal allowed_hosts: %w", err)
	}

	if len(flat)%2 != 0 {
		return fmt.Errorf("allowed_hosts array must have even length")
	}

	*hp = make([]HostPermission, 0, len(flat)/2)

	for i := 0; i < len(flat); i += 2 {
		// domain
		domain, ok := flat[i].(string)
		if !ok {
			return fmt.Errorf("expected string at index %d, got %T", i, flat[i])
		}

		// allow
		isFull, ok := flat[i+1].(bool)
		if !ok {
			return fmt.Errorf("expected bool at index %d, got %T", i+1, flat[i+1])
		}

		*hp = append(*hp, HostPermission{Domain: domain, IsFull: isFull})
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
