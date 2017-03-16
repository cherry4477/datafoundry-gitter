package main

import (
	"os"

	"github.com/zonesan/clog"
	"golang.org/x/oauth2"
	githuboauth "golang.org/x/oauth2/github"
)

var (
	gitlabBaseURL     = setBaseUrl(os.Getenv("GITLAB_BASEURL"))
	gitHubCallBackURL = os.Getenv("GITHUB_CALLBACK_URL")
	gitLabCallBackURL = os.Getenv("GITLAB_CALLBACK_URL")
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
			AuthURL:  gitlabBaseURL + "/oauth/authorize",
			TokenURL: gitlabBaseURL + "/oauth/token"},
	}

	// random string for oauth2 API calls to protect against CSRF
	oauthStateString = randToken()
)

func init() {
	if len(oauthConf.ClientID) == 0 || len(oauthConf.ClientSecret) == 0 {
		clog.Fatal("GITHUB_CLIENT_ID GITHUB_CLIENT_SECRET must be specified via .")
		return
	}

	if len(oauthConfGitLab.ClientID) == 0 || len(oauthConfGitLab.ClientSecret) == 0 || len(gitlabBaseURL) == 0 {
		clog.Fatal("GITLAB_BASEURL GITLAB_APP_ID GITLAB_CLIENT_SECRET must be specified.")
		return
	}

	if len(gitHubCallBackURL) == 0 {
		clog.Fatal("GITHUB_CALLBACK_URL must be specified.")
		return
	}

	if len(gitLabCallBackURL) == 0 {
		clog.Fatal("GITLAB_CALLBACK_URL must be specified.")
		return
	}

	clog.Debug("random state string:", oauthStateString)
	clog.Debugf("gitlab: %+v", oauthConfGitLab.Endpoint)
	clog.Debug("gitlab callback url:", gitLabCallBackURL)
	clog.Debug("github callback url:", gitHubCallBackURL)
}
