package main

import (
	"crypto/rand"
	"encoding/base64"
	"strings"
)

func setBaseUrl(urlStr string) string {
	// Make sure the given URL end with a slash
	if strings.HasSuffix(urlStr, "/") {
		return setBaseUrl(strings.TrimSuffix(urlStr, "/"))
	}
	return urlStr
}

func randToken() string {
	b := make([]byte, 40)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}
