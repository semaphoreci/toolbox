package metrics

import (
	"os"
	"strings"
)

func CacheUsername() string {
	return os.Getenv("SEMAPHORE_CACHE_USERNAME")
}

func CacheServerIP() string {
	cacheURL := os.Getenv("SEMAPHORE_CACHE_URL")
	if cacheURL == "" {
		return ""
	}

	ipAndPort := strings.Split(cacheURL, ":")
	if len(ipAndPort) != 2 {
		return ""
	}

	return ipAndPort[0]
}
