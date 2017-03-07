package main

import (
	"net/http"
	"os"

	"github.com/julienschmidt/httprouter"
	"github.com/zonesan/clog"
	"golang.org/x/oauth2"
	githuboauth "golang.org/x/oauth2/github"
)

var (
	gitlabBaseUrl = setBaseUrl(os.Getenv("GITLAB_BASEURL"))
	// You must register the app at https://github.com/settings/applications
	// Set callback to http://127.0.0.1:7000/github_oauth_cb
	// Set ClientId and ClientSecret to
	oauthConf = &oauth2.Config{
		ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		Scopes:       []string{"user:email", "repo"},
		Endpoint:     githuboauth.Endpoint,
	}

	oauthConfGitLab = &oauth2.Config{
		ClientID:     os.Getenv("GITLAB_APP_ID"),
		ClientSecret: os.Getenv("GITLAB_CLIENT_SECRET"),
		Scopes:       []string{"api"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  gitlabBaseUrl + "/oauth/authorize",
			TokenURL: gitlabBaseUrl + "/oauth/token"},
	}

	// random string for oauth2 API calls to protect against CSRF
	oauthStateString = "ashdkjahiweakdaiirhfljskaowr"
)

// http://localhost:18080/?code=958b1416f6362d24229ea051debeaa5256db9539ff12655f55aa0afc989429af&state=your_unique_state_hash
// https://gitlab.example.com/oauth/authorize?client_id=APP_ID&redirect_uri=REDIRECT_URI&response_type=code&state=your_unique_state_hash
func main() {

	if len(oauthConf.ClientID) == 0 || len(oauthConf.ClientSecret) == 0 {
		clog.Fatal("GITHUB_CLIENT_ID GITHUB_CLIENT_SECRET must be specified via .")
		return
	}

	if len(oauthConfGitLab.ClientID) == 0 || len(oauthConfGitLab.ClientSecret) == 0 || len(gitlabBaseUrl) == 0 {
		clog.Fatal("GITLAB_BASEURL GITLAB_APP_ID GITLAB_CLIENT_SECRET must be specified.")
		return
	}
	//clog.Debugf("github: %+v", oauthConf.Endpoint)
	clog.Debugf("gitlab: %+v", oauthConfGitLab.Endpoint)

	router := httprouter.New()
	router.GET("/", handleMain)

	//github
	router.GET("/login", handleGitHubLogin)
	router.GET("/github_oauth_cb", handleGitHubCallback)

	//gitlab
	router.GET("/authorize", handleGitLabAuthorize)
	router.GET("/gitlab_oauth_cb", handleGitLabCallback)

	clog.Debug("listening on port 7000 ...")
	clog.Fatal(http.ListenAndServe(":7000", router))

}
