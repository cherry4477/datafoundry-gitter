package main

import (
	"errors"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/zonesan/clog"
	"golang.org/x/oauth2"
)

// /github_oauth_cb. Called by github after authorization is granted
func handleGitHubCallback(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	clog.Info("from", r.RemoteAddr, r.Method, r.URL.RequestURI(), r.Proto)

	redirect_url, user, token, err := callbackValidate(w, r, oauthConf)
	// token, err := exchangeToken(oauthConfGitLab, code)
	if err != nil {
		clog.Errorf("oauthConf.Exchange() failed with '%s'\n", err)
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
	clog.Info("from", r.RemoteAddr, r.Method, r.URL.RequestURI(), r.Proto)

	redirect_url, user, token, err := callbackValidate(w, r, oauthConfGitLab)
	// token, err := exchangeToken(oauthConfGitLab, code)
	if err != nil {
		clog.Errorf("oauthConfGitLab.Exchange() failed with '%s'\n", err)
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
		clog.Errorf("oauthConf.Exchange() failed with '%s'\n", err)
		//http.Redirect(w, r, redirect_url, http.StatusFound)
	}

	return redirect_url, user, token, err
}
