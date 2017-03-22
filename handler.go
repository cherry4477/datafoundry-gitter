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

		authEnable := true

		if authEnable == false {
			handle(w, r, ps)
			return
		}

		token := r.Header.Get("Authorization")

		dfClient := NewDataFoundryTokenClient(token)

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
			// http.Redirect(w, r, "/authorize/github", http.StatusFound)
			RespError(w, ErrorNew(ErrCodeUnauthorized))
			return
		}
	case "gitlab":
		gitter, err = newLabGitter(user)
		if err != nil {
			// http.Redirect(w, r, "/authorize/gitlab", http.StatusFound)
			RespError(w, ErrorNew(ErrCodeUnauthorized))
			return
		}
	default:
		RespError(w, ErrorNew(ErrCodeNotFound))
		return
	}

	c := r.FormValue("cache")
	var cache bool = false
	if c == "true" {
		clog.Debug("using repos cache.")
		cache = true
	}

	// repos := gitter.ListPersonalRepos(user)
	repos := listPersonalRepos(gitter, cache)
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
			// http.Redirect(w, r, "/authorize/github", http.StatusFound)
			RespError(w, ErrorNew(ErrCodeUnauthorized))
			return
		}
		ns, repo = r.FormValue("ns"), r.FormValue("repo")
		if len(ns) == 0 || len(repo) == 0 {
			RespError(w, ErrorNewMsg(ErrCodeBadRequest, "ns or repo empty"))
			return
		}
	case "gitlab":
		gitter, err = newLabGitter(user)
		if err != nil {
			// http.Redirect(w, r, "/authorize/gitlab", http.StatusFound)
			RespError(w, ErrorNew(ErrCodeUnauthorized))
			return
		}
		repo = r.FormValue("id")
		if len(repo) == 0 {
			RespError(w, ErrorNewMsg(ErrCodeBadRequest, "id empty"))
			return
		}
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
			// http.Redirect(w, r, "/authorize/github", http.StatusFound)
			RespError(w, ErrorNew(ErrCodeUnauthorized))
			return
		}
	case "gitlab":
		gitter, err = newLabGitter(user)
		if err != nil {
			// http.Redirect(w, r, "/authorize/gitlab", http.StatusFound)
			RespError(w, ErrorNew(ErrCodeUnauthorized))
			return
		}
	default:
		RespError(w, ErrorNew(ErrCodeNotFound))
		return
	}

	token := r.Header.Get("Authorization")
	if len(token) == 0 || len(user) == 0 {
		// http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		RespError(w, ErrorNew(ErrCodeUnauthorized))
		return
	}

	if len(ns) == 0 {
		RespError(w, ErrorNewMsg(ErrCodeBadRequest, "ns empty"))
		return
	}

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
			// http.Redirect(w, r, "/authorize/github", http.StatusFound)
			RespError(w, ErrorNew(ErrCodeUnauthorized))
			return
		}
	case "gitlab":
		gitter, err = newLabGitter(user)
		if err != nil {
			// http.Redirect(w, r, "/authorize/gitlab", http.StatusFound)
			RespError(w, ErrorNew(ErrCodeUnauthorized))
			return
		}
	default:
		RespError(w, ErrorNew(ErrCodeNotFound))
		return
	}

	ns, bc := r.FormValue("ns"), r.FormValue("bc")
	if len(ns) == 0 || len(bc) == 0 {
		RespError(w, ErrorNewMsg(ErrCodeBadRequest, "ns or bc empty"))
		return
	}

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
			// http.Redirect(w, r, "/authorize/github", http.StatusFound)
			RespError(w, ErrorNew(ErrCodeUnauthorized))
			return
		}
	case "gitlab":
		gitter, err = newLabGitter(user)
		if err != nil {
			// http.Redirect(w, r, "/authorize/gitlab", http.StatusFound)
			RespError(w, ErrorNew(ErrCodeUnauthorized))
			return
		}
	default:
		RespError(w, ErrorNew(ErrCodeNotFound))
		return
	}

	hook := new(WebHook)
	if err := parseRequestBody(r, hook); err != nil {
		clog.Error("read request body error.", err)
		RespError(w, ErrorNew(ErrCodeInvalidParam))
		return
	}

	ns, bc := r.FormValue("ns"), r.FormValue("bc")
	if len(ns) == 0 || len(bc) == 0 {
		RespError(w, ErrorNewMsg(ErrCodeBadRequest, "ns or bc empty"))
		return
	}

	hook = createWebhook(gitter, ns, bc, hook)

	if hook == nil {
		RespError(w, ErrorNew(ErrCodeBadRequest))
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
			// http.Redirect(w, r, "/authorize/github", http.StatusFound)
			RespError(w, ErrorNew(ErrCodeUnauthorized))
			return
		}
	case "gitlab":
		gitter, err = newLabGitter(user)
		if err != nil {
			// http.Redirect(w, r, "/authorize/gitlab", http.StatusFound)
			RespError(w, ErrorNew(ErrCodeUnauthorized))
			return
		}
	default:
		RespError(w, ErrorNew(ErrCodeNotFound))
		return
	}

	ns, bc := r.FormValue("ns"), r.FormValue("bc")
	if len(ns) == 0 || len(bc) == 0 {
		RespError(w, ErrorNewMsg(ErrCodeBadRequest, "ns or bc empty"))
		return
	}

	err = removeWebhook(gitter, ns, bc, hookid)
	if err != nil {
		RespError(w, err)
		return
	}
	RespOK(w, nil)
}

func handleGitterAuthorize(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var url string

	source := ps.ByName("source")
	user := r.Header.Get("user")

	redirect_url := r.FormValue("redirect_url")
	if len(redirect_url) == 0 {
		RespError(w, ErrorNewMsg(ErrCodeBadRequest, "redirect_url empty"))
		return
	}

	queryStr := "?redirect_url=" + redirect_url + "&user=" + user
	clog.Debug("queryStr:", queryStr)

	switch source {
	case "github":
		// oauthConf.RedirectURL = gitHubCallBackURL + "?redirect_url=/repos/github&user=" + user
		oauthConf.RedirectURL = gitHubCallBackURL + queryStr
		url = oauthConf.AuthCodeURL(oauthStateString, oauth2.AccessTypeOnline)
	case "gitlab":
		// oauthConfGitLab.RedirectURL = gitLabCallBackURL + "?redirect_url=/repos/gitlab&user=" + user
		oauthConfGitLab.RedirectURL = gitLabCallBackURL + queryStr
		url = oauthConfGitLab.AuthCodeURL(oauthStateString, oauth2.AccessTypeOnline)
	default:
		RespError(w, ErrorNew(ErrCodeNotFound))
		return
	}
	//user and redirect_url must be set here.
	http.Redirect(w, r, url, http.StatusFound)
}
