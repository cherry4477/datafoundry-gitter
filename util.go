package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
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

// func redirectUrl(oauthConf *oauth2.Config) string {
// 	return ""
// }

func debug(v interface{}) {
	d, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Printf("json.MarshlIndent() failed with %s\n", err)
	}
	fmt.Println(string(d))
}
