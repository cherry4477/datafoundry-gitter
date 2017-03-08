package main

import (
	"github.com/zonesan/clog"
	"golang.org/x/oauth2"
)

// ListPersonalRepos(user string)
// ListOrgRepos(org string)
// ListBranches(owner, repo string)
// ListTags(owner, repo string)
// CreateWebhook(hook interface{})
// RemoveWebhook(hook interface{})
// CheckWebhook(hook interface{})

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

func loadGitLabToken(store Storage, user string) {
	clog.Debug("loadGitLabToken interface")
	store.LoadTokenGitlab(user)
}

func saveGitLabToken(store Storage, tok *oauth2.Token) {
	clog.Debug("saveGitLabToken interface")
	store.SaveTokenGitlab(tok)
}

func loadGitHubToken(store Storage, user string) {
	clog.Debug("loadGitHubToken interface")
	store.LoadTokenGithub(user)
}

func saveGitHubToken(store Storage, tok *oauth2.Token) {
	clog.Debug("saveGitHubToken interface")
	store.SaveTokenGithub(tok)
}
