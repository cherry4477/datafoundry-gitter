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

func handleMain(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	clog.Info("from", r.RemoteAddr, r.Method, r.URL.RequestURI(), r.Proto)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(htmlIndex))
}

func handleRepos(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	clog.Info("from", r.RemoteAddr, r.Method, r.URL.RequestURI(), r.Proto)
	source := ps.ByName("source")
	user := "zonesan"

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
	repos := listPersonalRepos(gitter, user)
	RespOK(w, repos)
}

func handleRepoBranches(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	clog.Info("from", r.RemoteAddr, r.Method, r.URL.RequestURI(), r.Proto)
	source := ps.ByName("source")
	user := "zonesan"

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

func handleCheckWebhook(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	clog.Info("from", r.RemoteAddr, r.Method, r.URL.RequestURI(), r.Proto)
	source := ps.ByName("source")
	user := "zonesan"

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
	user := "zonesan"

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

	user := "zonesan"

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
