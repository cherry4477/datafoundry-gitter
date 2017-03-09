package main

import (
	"github.com/zonesan/clog"
	"golang.org/x/oauth2"
)

func listPersonalRepos(gitter Gitter, user string) {
	clog.Debug("listPersonalRepos interface")
	gitter.ListPersonalRepos(user)
}

func listOrgRepos(gitter Gitter, org string) {
	clog.Debug("listOrgRepos interface")
	gitter.ListOrgRepos(org)
}

func listBranches(gitter Gitter, owner, repo string) {
	clog.Debug("listBranches interface")
	gitter.ListBranches(owner, repo)
}

func listTags(gitter Gitter, owner, repo string) {
	clog.Debug("listTags interface")
	gitter.ListTags(owner, repo)
}

func createWebhook(gitter Gitter, hook interface{}) {
	clog.Debug("createWebhook interface")
	gitter.CreateWebhook(hook)
}

func removeWebhook(gitter Gitter, hook interface{}) {
	clog.Debug("removeWebhook interface")
	gitter.RemoveWebhook(hook)
}

func checkWebhook(gitter Gitter, hook interface{}) {
	clog.Debug("checkWebhook interface")
	gitter.CheckWebhook(hook)
}

// func loadToken(gitter Gitter) (*oauth2.Token, error) {
// 	clog.Debug("loadToken interface")
// 	return gitter.LoadToken()
// }

// func saveToken(gitter Gitter, tok *oauth2.Token) error {
// 	clog.Debug("saveToken interface")
// 	return gitter.SaveToken(tok)
// }

func loadGitLabToken(store Storage, user string) {
	clog.Debug("loadGitLabToken interface")
	store.LoadTokenGitlab(user)
}

func saveGitLabToken(store Storage, user string, tok *oauth2.Token) error {
	clog.Debug("saveGitLabToken interface")
	store.SaveTokenGitlab(user, tok)
	return nil
}

func loadGitHubToken(store Storage, user string) {
	clog.Debug("loadGitHubToken interface")
	store.LoadTokenGithub(user)
}

func saveGitHubToken(store Storage, user string, tok *oauth2.Token) {
	clog.Debug("saveGitHubToken interface")
	store.SaveTokenGithub(user, tok)
}

func exchangeToken(oauthConf *oauth2.Config, code string) (*oauth2.Token, error) {
	return oauthConf.Exchange(oauth2.NoContext, code)
}
