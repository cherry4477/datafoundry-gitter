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
	var err error

	switch source {
	case "github":
		git, err = newHubGitter(user)
		if err != nil {
			http.Redirect(w, r, "/authorize/github", http.StatusFound)
			return
		}
	case "gitlab":
		git, err = newLabGitter(user)
		if err != nil {
			http.Redirect(w, r, "/authorize/gitlab", http.StatusFound)
			return
		}
	default:
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(http.StatusText(http.StatusNotFound)))
		return
	}
	repos := git.ListPersonalRepos(user)
	RespOK(w, repos)
}

func handleGitHubRepos(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	user := "zonesan"

	git, err := newHubGitter(user)
	if err != nil {
		http.Redirect(w, r, "/authorize/github", http.StatusFound)
		return
	}

	repos := git.ListPersonalRepos(user)
	RespOK(w, repos)

}

func handleGitLabRepos(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	user := "zonesan"

	git, err := newLabGitter(user)
	if err != nil {
		http.Redirect(w, r, "/authorize/gitlab", http.StatusFound)
		return
	}

	repos := git.ListPersonalRepos(user)
	RespOK(w, repos)

}

func handleGitLabRepoBranches(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	user := "zonesan"
	id := r.FormValue("id")

	git, err := newLabGitter(user)
	if err != nil {
		http.Redirect(w, r, "/authorize/gitlab", http.StatusFound)
		return
	}

	branches := git.ListBranches("", id)
	RespOK(w, branches)

}

func handleGitHubRepoBranches(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	user := "zonesan"
	namespace, repo := r.FormValue("ns"), r.FormValue("repo")

	git, err := newHubGitter(user)
	if err != nil {
		http.Redirect(w, r, "/authorize/github", http.StatusFound)
		return
	}

	branches := git.ListBranches(namespace, repo)

	RespOK(w, branches)
}

func handleGitHubCheckWebhook(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	user := "zonesan"
	// namespace, repo := r.FormValue("namespace"), r.FormValue("repo")
	// if len(namespace) == 0 || len(repo) == 0 {
	// 	w.WriteHeader(http.StatusNotFound)
	// 	w.Write([]byte(http.StatusText(http.StatusNotFound)))
	// 	return
	// }
	ns, bc := r.FormValue("ns"), r.FormValue("bc")

	git, err := newHubGitter(user)
	if err != nil {
		http.Redirect(w, r, "/authorize/github", http.StatusFound)
		return
	}
	hook := git.CheckWebhook(ns, bc)
	RespOK(w, hook)
}

func handleGitHubCreateWebhook(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	handleMain(w, r, ps)
}
func handleGitHubRemoveWebhook(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	handleMain(w, r, ps)
}

func handleGitLabCheckWebhook(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ns, bc := r.FormValue("ns"), r.FormValue("bc")
	if len(ns) == 0 || len(bc) == 0 {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(http.StatusText(http.StatusNotFound)))
		return
	}
	handleMain(w, r, ps)
}
func handleGitLabCreateWebhook(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	handleMain(w, r, ps)
}
func handleGitLabRemoveWebhook(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	handleMain(w, r, ps)
}

func handleGitterAuthorize(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var url string

	source := ps.ByName("source")

	switch source {
	case "github":
		oauthConf.RedirectURL = gitHubCallBackURL + "?redirect_url=/repos/github&user=zonesan"
		url = oauthConf.AuthCodeURL(oauthStateString, oauth2.AccessTypeOnline)
	case "gitlab":
		oauthConfGitLab.RedirectURL = gitLabCallBackURL + "?redirect_url=/repos/gitlab&user=zonesan"
		url = oauthConfGitLab.AuthCodeURL(oauthStateString, oauth2.AccessTypeOnline)
	default:
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(http.StatusText(http.StatusNotFound)))
		return
	}
	//user and redirect_url must be set here.
	http.Redirect(w, r, url, http.StatusFound)
}

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

}

// /github_oauth_cb. Called by github after authorization is granted
func handleGitLabCallback(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

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
