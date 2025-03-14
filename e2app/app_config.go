package e2app

import (
	"encoding/base64"
)

type AppConfig struct {
	Name       string         `mapstructure:"name"`
	ExtraProps map[string]any `mapstructure:",remain"`
}

func (ap *AppConfig) Get(key string) any {
	if val, exists := ap.ExtraProps[key]; exists {
		return val
	}
	return nil
}

func (ap *AppConfig) GetString(key string) string {
	if val, ok := ap.ExtraProps[key].(string); ok {
		return val
	}
	return ""
}

func (ap *AppConfig) GetInt(key string) int {
	if val, ok := ap.ExtraProps[key].(int); ok {
		return val
	}
	return 0
}

func (ap *AppConfig) GetFloat(key string) float64 {
	if val, ok := ap.ExtraProps[key].(float64); ok {
		return val
	}
	return 0.0
}

func (ap *AppConfig) GetBool(key string) bool {
	if val, ok := ap.ExtraProps[key].(bool); ok {
		return val
	}
	return false
}

// GetStringSlice tags
func (ap *AppConfig) GetStringSlice(key string) []string {
	if val, ok := ap.ExtraProps[key].([]any); ok {
		var result []string
		for _, v := range val {
			if str, ok := v.(string); ok {
				result = append(result, str)
			}
		}
		return result
	}
	return []string{}
}

// GetStringMap [app.settings]
func (ap *AppConfig) GetStringMap(key string) map[string]any {
	if val, ok := ap.ExtraProps[key].(map[string]any); ok {
		return val
	}
	return map[string]any{}
}

func (ap *AppConfig) GetBytesFromBase64(key string) []byte {
	if val, ok := ap.ExtraProps[key].(string); ok {
		decoded, err := base64.StdEncoding.DecodeString(val)
		if err == nil {
			return decoded
		}
	}
	return nil
}
