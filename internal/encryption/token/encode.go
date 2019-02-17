package token

import (
	"encoding/base64"
	"encoding/json"
)

func encodeJSONB64(v interface{}) string {
	if v == nil {
		return ""
	}
	bytes, err := json.Marshal(v)
	if err != nil || len(bytes) == 0 {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(bytes)
}

func decodeJSONB64(s string, v interface{}) bool {
	if s == "" {
		return false
	}
	bytes, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil || len(bytes) == 0 {
		return false
	}
	if err = json.Unmarshal(bytes, v); err != nil {
		return false
	}
	return true
}
