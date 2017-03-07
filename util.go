package main

import (
	"strings"
)

func setBaseUrl(urlStr string) string {
	// Make sure the given URL end with a slash
	if strings.HasSuffix(urlStr, "/") {
		return setBaseUrl(strings.TrimSuffix(urlStr, "/"))
	}
	return urlStr
}
