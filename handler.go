package main

import (
	"fmt"
	"net/http"

	"github.com/google/go-github/github"
	"github.com/julienschmidt/httprouter"
	gitlab "github.com/xanzy/go-gitlab"
	"github.com/zonesan/clog"
	"golang.org/x/oauth2"
)

const htmlIndex = `<html><body>
Logged in with <a href="/login">GitHub</a><br />
Logged in with <a href="/authorize">GitLab</a><br />

</body></html>
`

// /
func handleMain(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(htmlIndex))
}

// /login
func handleGitHubLogin(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	oauthConf.RedirectURL = "http://localhost:7000/github_oauth_cb?redirect_url=abcde&user=zonesan"
	url := oauthConf.AuthCodeURL(oauthStateString, oauth2.AccessTypeOnline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// /github_oauth_cb. Called by github after authorization is granted
func handleGitHubCallback(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	redirect_url := r.FormValue("redirect_url")
	user := r.FormValue("user")
	clog.Debug("user:", user, "redirect_url:", redirect_url)

	state := r.FormValue("state")
	clog.Debug("state:", state)

	if state != oauthStateString {
		fmt.Printf("invalid oauth state, expected '%s', got '%s'\n", oauthStateString, state)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	code := r.FormValue("code")
	clog.Debug("code:", code)

	token, err := oauthConf.Exchange(oauth2.NoContext, code)
	if err != nil {
		fmt.Printf("oauthConf.Exchange() failed with '%s'\n", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	clog.Debug(token)

	oauthClient := oauthConf.Client(oauth2.NoContext, token)
	client := github.NewClient(oauthClient)
	//do something.

	go func() {
		var user string = "zonesan"
		UserProfile(client, user)
		//ListPersonalRepos(client, user)
		//ListOrgRepos(client)
	}()

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func handleGitLabAuthorize(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	oauthConfGitLab.RedirectURL = "http://localhost:7000/gitlab_oauth_cb?redirect_url=abcde&user=zonesan"
	url := oauthConfGitLab.AuthCodeURL(oauthStateString, oauth2.AccessTypeOnline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// /github_oauth_cb. Called by github after authorization is granted
func handleGitLabCallback(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	redirect_url := r.FormValue("redirect_url")
	user := r.FormValue("user")
	clog.Debug("user:", user, "redirect_url:", redirect_url)

	state := r.FormValue("state")
	clog.Debug("state:", state)

	if state != oauthStateString {
		fmt.Printf("invalid oauth state, expected '%s', got '%s'\n", oauthStateString, state)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	code := r.FormValue("code")
	clog.Debug("code:", code)

	token, err := oauthConfGitLab.Exchange(oauth2.NoContext, code)
	if err != nil {
		fmt.Printf("oauthConfGitLab.Exchange() failed with '%s'\n", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	oauthClient := oauthConfGitLab.Client(oauth2.NoContext, token)
	client := gitlab.NewOAuthClient(oauthClient, token.AccessToken)
	client.SetBaseURL(gitlabBaseUrl + "/api/v3")
	//do something.

	clog.Debug(token)

	go func() {
		a, b, c := client.Users.CurrentUser()
		clog.Debug("user:", a, b, c)
		//ListPersonalRepos(client, user)
		//ListOrgRepos(client)
		//opt := &gitlab.ListProjectsOptions{}
		d, e, f := client.Projects.ListProjects(nil)
		clog.Debug("project", d, e, f)

		session, resp, err := client.Session.GetSession(nil)
		clog.Debugf("session", session, resp, err)

	}()

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}
