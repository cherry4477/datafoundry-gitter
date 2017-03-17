package main

import (
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

type Mux struct{}

func (m *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	clog.Info("from", r.RemoteAddr, r.Method, r.URL.RequestURI(), r.Proto)
	//RespError(w, ErrorNew(ErrCodeNotFound), http.StatusNotFound)
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("not found"))
}

func authorize(handle httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		clog.Info("from", r.RemoteAddr, r.Method, r.URL.RequestURI(), r.Proto)

		authEnable := false

		if authEnable == false {
			handle(w, r, ps)
			return
		}

		token := r.Header.Get("Authorization")

		dfClient := NewDataFoundryTokenClient("https://10.1.130.134:8443", token)

		user := new(User)
		err := dfClient.OGet("/users/~", user)
		if err != nil {
			code := http.StatusBadRequest
			if e, ok := err.(*StatusError); ok {
				code = int(e.ErrStatus.Code)
			}
			http.Error(w, err.Error(), code)
			return
		}

		r.Header.Set("user", user.Name)

		ok := true
		if ok {
			handle(w, r, ps)
		} else {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		}
	}
}

func handleMain(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	clog.Info("from", r.RemoteAddr, r.Method, r.URL.RequestURI(), r.Proto)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(htmlIndex))
}

func handleRepos(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	//clog.Info("from", r.RemoteAddr, r.Method, r.URL.RequestURI(), r.Proto)
	source := ps.ByName("source")
	user := r.Header.Get("user")

	var gitter Gitter
	var err error

	switch source {
	case "github":
		gitter, err = newHubGitter(user)
		if err != nil {
			http.Redirect(w, r, "/authorize/github", http.StatusFound)
			return
		}
	case "gitlab":
		gitter, err = newLabGitter(user)
		if err != nil {
			http.Redirect(w, r, "/authorize/gitlab", http.StatusFound)
			return
		}
	default:
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(http.StatusText(http.StatusNotFound)))
		return
	}
	// repos := gitter.ListPersonalRepos(user)
	repos := listPersonalRepos(gitter)
	RespOK(w, repos)
}

func handleRepoBranches(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	clog.Info("from", r.RemoteAddr, r.Method, r.URL.RequestURI(), r.Proto)
	source := ps.ByName("source")
	user := r.Header.Get("user")

	var gitter Gitter
	var err error
	var ns, repo string

	switch source {
	case "github":
		gitter, err = newHubGitter(user)
		if err != nil {
			http.Redirect(w, r, "/authorize/github", http.StatusFound)
			return
		}
		ns, repo = r.FormValue("ns"), r.FormValue("repo")
	case "gitlab":
		gitter, err = newLabGitter(user)
		if err != nil {
			http.Redirect(w, r, "/authorize/gitlab", http.StatusFound)
			return
		}
		repo = r.FormValue("id")
	default:
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(http.StatusText(http.StatusNotFound)))
		return
	}

	// branches := git.ListBranches(ns, repo)
	branches := listBranches(gitter, ns, repo)
	RespOK(w, branches)

}

func handleSecret(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	source := ps.ByName("source")
	user := r.Header.Get("user")
	ns := r.FormValue("ns")

	var gitter Gitter
	var err error

	switch source {
	case "github":
		gitter, err = newHubGitter(user)
		if err != nil {
			http.Redirect(w, r, "/authorize/github", http.StatusFound)
			return
		}
	case "gitlab":
		gitter, err = newLabGitter(user)
		if err != nil {
			http.Redirect(w, r, "/authorize/gitlab", http.StatusFound)
			return
		}
	default:
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(http.StatusText(http.StatusNotFound)))
		return
	}
	token := r.Header.Get("Authorization")
	gitter.SetBearerToken(token)
	secret := checkSecret(gitter, ns)
	RespOK(w, secret)
}

func handleCheckWebhook(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	clog.Info("from", r.RemoteAddr, r.Method, r.URL.RequestURI(), r.Proto)
	source := ps.ByName("source")
	user := r.Header.Get("user")

	var gitter Gitter
	var err error

	switch source {
	case "github":
		gitter, err = newHubGitter(user)
		if err != nil {
			http.Redirect(w, r, "/authorize/github", http.StatusFound)
			return
		}
	case "gitlab":
		gitter, err = newLabGitter(user)
		if err != nil {
			http.Redirect(w, r, "/authorize/gitlab", http.StatusFound)
			return
		}
	default:
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(http.StatusText(http.StatusNotFound)))
		return
	}

	ns, bc := r.FormValue("ns"), r.FormValue("bc")
	hook := checkWebhook(gitter, ns, bc)
	RespOK(w, hook)
}

func handleCreateWebhook(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	clog.Info("from", r.RemoteAddr, r.Method, r.URL.RequestURI(), r.Proto)
	source := ps.ByName("source")
	user := r.Header.Get("user")

	var gitter Gitter
	var err error

	switch source {
	case "github":
		gitter, err = newHubGitter(user)
		if err != nil {
			http.Redirect(w, r, "/authorize/github", http.StatusFound)
			return
		}
	case "gitlab":
		gitter, err = newLabGitter(user)
		if err != nil {
			http.Redirect(w, r, "/authorize/gitlab", http.StatusFound)
			return
		}
	default:
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(http.StatusText(http.StatusNotFound)))
		return
	}

	hook := new(WebHook)
	if err := parseRequestBody(r, hook); err != nil {
		clog.Error("read request body error.", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(http.StatusText(http.StatusBadRequest)))
		return
	}

	ns, bc := r.FormValue("ns"), r.FormValue("bc")

	hook = createWebhook(gitter, ns, bc, hook)

	if hook == nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(http.StatusText(http.StatusBadRequest)))
		return
	}

	RespOK(w, hook)
}

func handleRemoveWebhook(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	clog.Info("from", r.RemoteAddr, r.Method, r.URL.RequestURI(), r.Proto)
	source, hookid := ps.ByName("source"), ps.ByName("hookid")

	user := r.Header.Get("user")

	var gitter Gitter
	var err error

	switch source {
	case "github":
		gitter, err = newHubGitter(user)
		if err != nil {
			http.Redirect(w, r, "/authorize/github", http.StatusFound)
			return
		}
	case "gitlab":
		gitter, err = newLabGitter(user)
		if err != nil {
			http.Redirect(w, r, "/authorize/gitlab", http.StatusFound)
			return
		}
	default:
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(http.StatusText(http.StatusNotFound)))
		return
	}
	ns, bc := r.FormValue("ns"), r.FormValue("bc")
	err = removeWebhook(gitter, ns, bc, hookid)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	RespOK(w, nil)
}

func handleGitterAuthorize(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var url string

	source := ps.ByName("source")
	user := r.Header.Get("user")

	switch source {
	case "github":
		oauthConf.RedirectURL = gitHubCallBackURL + "?redirect_url=/repos/github&user=" + user
		url = oauthConf.AuthCodeURL(oauthStateString, oauth2.AccessTypeOnline)
	case "gitlab":
		oauthConfGitLab.RedirectURL = gitLabCallBackURL + "?redirect_url=/repos/gitlab&user=" + user
		url = oauthConfGitLab.AuthCodeURL(oauthStateString, oauth2.AccessTypeOnline)
	default:
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(http.StatusText(http.StatusNotFound)))
		return
	}
	//user and redirect_url must be set here.
	http.Redirect(w, r, url, http.StatusFound)
}
