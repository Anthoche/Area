package workflows

import (
	"encoding/json"
	"strings"

	"area/src/security"
)

var sensitiveKeys = map[string]struct{}{
	"bot_token": {},
	"token":     {},
	"api_key":   {},
}

// encryptTriggerConfig recursively encrypts sensitive fields in the trigger configuration.
func encryptTriggerConfig(raw json.RawMessage) json.RawMessage {
	if !security.Enabled() || len(raw) == 0 {
		return raw
	}
	var cfg map[string]any
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return raw
	}
	cfg = encryptSensitiveFields(cfg).(map[string]any)
	encoded, err := json.Marshal(cfg)
	if err != nil {
		return raw
	}
	return encoded
}

func decryptPayload(raw json.RawMessage) json.RawMessage {
	if !security.Enabled() || len(raw) == 0 {
		return raw
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return raw
	}
	payload = decryptSensitiveFields(payload).(map[string]any)
	encoded, err := json.Marshal(payload)
	if err != nil {
		return raw
	}
	return encoded
}

// encryptSensitiveFields recursively encrypts sensitive fields in the given value.
func encryptSensitiveFields(value any) any {
	switch v := value.(type) {
	case map[string]any:
		for key, val := range v {
			_, isSensitive := sensitiveKeys[strings.ToLower(key)]
			if isSensitive {
				if s, ok := val.(string); ok && s != "" && !strings.HasPrefix(s, "enc:") {
					if enc, err := security.EncryptString(s); err == nil {
						v[key] = enc
						continue
					}
				}
			}
			v[key] = encryptSensitiveFields(val)
		}
		return v
	case []any:
		for i, item := range v {
			v[i] = encryptSensitiveFields(item)
		}
		return v
	default:
		return v
	}
}

// decryptSensitiveFields recursively decrypts sensitive fields in the given value.
func decryptSensitiveFields(value any) any {
	switch v := value.(type) {
	case map[string]any:
		for key, val := range v {
			_, isSensitive := sensitiveKeys[strings.ToLower(key)]
			if isSensitive {
				if s, ok := val.(string); ok && strings.HasPrefix(s, "enc:") {
					if dec, err := security.DecryptString(s); err == nil {
						v[key] = dec
						continue
					}
				}
			}
			v[key] = decryptSensitiveFields(val)
		}
		return v
	case []any:
		for i, item := range v {
			v[i] = decryptSensitiveFields(item)
		}
		return v
	default:
		return v
	}
}
