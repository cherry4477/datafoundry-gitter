package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/zonesan/clog"
	"golang.org/x/oauth2"
)

const htmlIndex = `<html><body>
Logged in with <a href="/authorize/github">GitHub</a><br />
Logged in with <a href="/authorize/gitlab">GitLab</a><br />

</body></html>
`

// /
func handleMain(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(htmlIndex))
}

func handleRepos(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	source := ps.ByName("source")
	user := "zonesan"

	var git Gitter

	switch source {
	case "github":
		tok := loadGitHubToken(store, user)
		if tok == nil {
			clog.Errorf("can't load github token for user %v, need redirect to authorize.", user)
			http.Redirect(w, r, "/authorize/github", http.StatusFound)
			return
		}
		gitter := NewGitHub(tok)
		if gitter == nil {
			http.Redirect(w, r, "/authorize/github", http.StatusFound)
			return
		}
		git = gitter
	case "gitlab":
		tok := loadGitLabToken(store, user)
		if tok == nil {
			clog.Errorf("can't load gitlab token for user %v, need redirect to authorize.", user)
			http.Redirect(w, r, "/authorize/gitlab", http.StatusFound)
		}
		gitter := NewGitLab(tok)
		if gitter == nil {
			http.Redirect(w, r, "/authorize/gitlab", http.StatusFound)
			return
		}
		git = gitter
	default:
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(http.StatusText(http.StatusNotFound)))
		return
	}
	git.ListPersonalRepos(user)
}
func handleRepoBranches(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	handleMain(w, r, ps)
}
func handleCheckWebhook(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	handleMain(w, r, ps)
}
func handleCreateWebhook(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	handleMain(w, r, ps)
}
func handleRemoveWebhook(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	handleMain(w, r, ps)
}

func handleGitterAuthorize(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var url string

	source := ps.ByName("source")

	switch source {
	case "github":
		oauthConf.RedirectURL = gitHubCallBackURL + "?redirect_url=abcde&user=zonesan"
		url = oauthConf.AuthCodeURL(oauthStateString, oauth2.AccessTypeOnline)
	case "gitlab":
		oauthConfGitLab.RedirectURL = gitLabCallBackURL + "?redirect_url=abcde&user=zonesan"
		url = oauthConfGitLab.AuthCodeURL(oauthStateString, oauth2.AccessTypeOnline)
	default:
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(http.StatusText(http.StatusNotFound)))
		return
	}
	//user and redirect_url must be set here.
	http.Redirect(w, r, url, http.StatusFound)
}

// /login
// func handleGitHubLogin(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
// 	oauthConf.RedirectURL = "http://localhost:7000/github_oauth_cb?redirect_url=abcde&user=zonesan"
// 	url := oauthConf.AuthCodeURL(oauthStateString, oauth2.AccessTypeOnline)
// 	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
// }

// /github_oauth_cb. Called by github after authorization is granted
func handleGitHubCallback(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	redirect_url, user, token, err := callbackValidate(w, r, oauthConf)
	// token, err := exchangeToken(oauthConfGitLab, code)
	if err != nil {
		fmt.Printf("oauthConf.Exchange() failed with '%s'\n", err)
		http.Redirect(w, r, redirect_url, http.StatusFound)
		return
	}

	if err := saveGitHubToken(store, user, token); err != nil {
		clog.Error("save gitlab token error:", err)
	}

	http.Redirect(w, r, redirect_url, http.StatusFound)

	// go func() {
	// 	var user string = "zonesan"
	// 	UserProfile(client, user)
	// 	//ListPersonalRepos(client, user)
	// 	//ListOrgRepos(client)
	// }()

	// http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

// /github_oauth_cb. Called by github after authorization is granted
func handleGitLabCallback(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// redirect_url := r.FormValue("redirect_url")
	// user := r.FormValue("user")
	// state := r.FormValue("state")
	// code := r.FormValue("code")

	// clog.Debug("user:", user, "redirect_url:", redirect_url, "state:", state, "code:", code)

	// if state != oauthStateString {
	// 	fmt.Printf("invalid oauth state, expected '%s', got '%s'\n", oauthStateString, state)
	// 	http.Redirect(w, r, "/", http.StatusFound)
	// 	return
	// }

	redirect_url, user, token, err := callbackValidate(w, r, oauthConfGitLab)
	// token, err := exchangeToken(oauthConfGitLab, code)
	if err != nil {
		fmt.Printf("oauthConfGitLab.Exchange() failed with '%s'\n", err)
		http.Redirect(w, r, redirect_url, http.StatusFound)
		return
	}

	if err := saveGitLabToken(store, user, token); err != nil {
		clog.Error("save gitlab token error:", err)
	}

	http.Redirect(w, r, redirect_url, http.StatusFound)

	// oauthClient := oauthConfGitLab.Client(oauth2.NoContext, token)
	// client := gitlab.NewOAuthClient(oauthClient, token.AccessToken)
	// client.SetBaseURL(gitlabBaseUrl + "/api/v3")
	// //do something.

	// clog.Debug(token)

	// go func() {
	// 	a, b, c := client.Users.CurrentUser()
	// 	clog.Debug("user:", a, b, c)
	// 	//ListPersonalRepos(client, user)
	// 	//ListOrgRepos(client)
	// 	//opt := &gitlab.ListProjectsOptions{}
	// 	d, e, f := client.Projects.ListProjects(nil)
	// 	clog.Debug("project", d, e, f)

	// 	session, resp, err := client.Session.GetSession(nil)
	// 	clog.Debugf("session", session, resp, err)

	// }()

}

func callbackValidate(w http.ResponseWriter, r *http.Request, oauthConf *oauth2.Config) (string, string, *oauth2.Token, error) {
	redirect_url := r.FormValue("redirect_url")
	user := r.FormValue("user")
	state := r.FormValue("state")
	code := r.FormValue("code")

	clog.Debug("user:", user, "redirect_url:", redirect_url, "state:", state, "code:", code)

	if state != oauthStateString {
		clog.Errorf("invalid oauth state, expected '%s', got '%s'\n", oauthStateString, state)
		//http.Redirect(w, r, "/", http.StatusFound)
		return "", "", nil, errors.New("invalid oauth state.")
	}

	token, err := exchangeToken(oauthConf, code)
	// token, err := oauthConfGitLab.Exchange(oauth2.NoContext, code)
	if err != nil {
		fmt.Printf("oauthConf.Exchange() failed with '%s'\n", err)
		//http.Redirect(w, r, redirect_url, http.StatusFound)
	}

	return redirect_url, user, token, err
}
