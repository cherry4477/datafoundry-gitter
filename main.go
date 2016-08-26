package main

import (
	"fmt"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	githuboauth "golang.org/x/oauth2/github"
)

var (
	// You must register the app at https://github.com/settings/applications
	// Set callback to http://127.0.0.1:7000/github_oauth_cb
	// Set ClientId and ClientSecret to
	oauthConf = &oauth2.Config{
		ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		Scopes:       []string{"user:email", "repo"},
		Endpoint:     githuboauth.Endpoint,
	}
	// random string for oauth2 API calls to protect against CSRF
	oauthStateString = "ashdkjahiweakdaiirhfljskaowr"
)

func main() {

	if len(oauthConf.ClientID) == 0 || len(oauthConf.ClientSecret) == 0 {
		fmt.Println("clientID or clientSecret must be specified.")
		return
	}

	http.HandleFunc("/", handleMain)
	http.HandleFunc("/login", handleGitHubLogin)
	http.HandleFunc("/github_oauth_cb", handleGitHubCallback)

	fmt.Println("Started running on http://127.0.0.1:7000")
	fmt.Println(http.ListenAndServe(":7000", nil))
}
