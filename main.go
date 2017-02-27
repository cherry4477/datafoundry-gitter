package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/julienschmidt/httprouter"
	"github.com/zonesan/clog"
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

	router := httprouter.New()
	router.GET("/", handleMain)
	router.GET("/login", handleGitHubLogin)
	router.GET("/github_oauth_cb", handleGitHubCallback)

	clog.Debug("listening on port 7000 ...")
	clog.Fatal(http.ListenAndServe(":7000", router))

}
