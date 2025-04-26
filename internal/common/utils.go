package common

import (
	"os"
)

func GetEnv(key string, defaultValue ...string) string {
	val := os.Getenv(key)
	if val == "" {
		if len(defaultValue) == 0 {
			return ""
		}
		return defaultValue[0]
	}
	return val
}

func GetEndpointURL(endpoint string) string {
	return os.Getenv("HOST") + ":" + os.Getenv("PORT") + endpoint
}
