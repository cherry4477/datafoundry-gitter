package main

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/zonesan/clog"
	"golang.org/x/oauth2"
)

func listPersonalRepos(gitter Gitter, cache bool) *[]Repositories {
	clog.Debug("listPersonalRepos interface")
	return gitter.ListPersonalRepos(cache)
}

func listOrgRepos(gitter Gitter, org string) {
	clog.Debug("listOrgRepos interface")
	gitter.ListOrgRepos(org)
}

func listBranches(gitter Gitter, owner, repo string) *[]Branch {
	clog.Debug("listBranches interface")
	return gitter.ListBranches(owner, repo)
}

func listTags(gitter Gitter, owner, repo string) {
	clog.Debug("listTags interface")
	gitter.ListTags(owner, repo)
}

func createWebhook(gitter Gitter, ns, bc string, hook *WebHook) *WebHook {
	clog.Debug("createWebhook interface")
	hook.Name = ns + "/" + bc
	return gitter.CreateWebhook(hook)
}

func checkWebhook(gitter Gitter, ns, bc string) *WebHook {
	clog.Debug("checkWebhook interface")
	return gitter.CheckWebhook(ns, bc)
}

func checkSecret(gitter Gitter, ns string) *Secret {
	clog.Debug("checkSecret interface")
	var err error

	secret := gitter.CheckSecret(ns)
	if secret == nil {
		secretName := gitter.Source() + "-" + gitter.User() + "-" + randomStr(8)
		if secret, err = gitter.CreateSecret(ns, secretName); err != nil {
			clog.Error(err)
			return nil
		}
	}
	return secret
}

func removeWebhook(gitter Gitter, ns, bc, hookid string) error {
	clog.Debug("removeWebhook interface")
	id, err := strconv.Atoi(hookid)
	if err != nil {
		clog.Error(err)
		return err
	}
	return gitter.RemoveWebhook(ns, bc, id)
}

func loadGitLabToken(store Storage, user string) *oauth2.Token {
	clog.Debug("loadGitLabToken interface")
	tok, err := store.LoadTokenGitlab(user)
	if err != nil {
		clog.Error(err)
		return nil
	}
	return tok
}

func saveGitLabToken(store Storage, user string, tok *oauth2.Token) error {
	clog.Debug("saveGitLabToken interface")
	store.SaveTokenGitlab(user, tok)
	return nil
}

func loadGitHubToken(store Storage, user string) *oauth2.Token {
	clog.Debug("loadGitHubToken interface")
	tok, err := store.LoadTokenGithub(user)
	if err != nil {
		clog.Error(err)
		return nil
	}
	return tok
}

func saveGitHubToken(store Storage, user string, tok *oauth2.Token) error {
	clog.Debug("saveGitHubToken interface")
	return store.SaveTokenGithub(user, tok)
}

func exchangeToken(oauthConf *oauth2.Config, code string) (*oauth2.Token, error) {
	return oauthConf.Exchange(oauth2.NoContext, code)
}

func newLabGitter(user string) (Gitter, error) {
	tok := loadGitLabToken(store, user)
	if tok == nil {
		errStr := fmt.Sprintf("can't load gitlab token for user %v, need redirect to authorize.", user)
		clog.Error(errStr)
		return nil, errors.New(errStr)
	}

	gitter := NewGitLab(tok)
	if gitter == nil {
		errStr := fmt.Sprintf("empty Gitter returned, need authoriza.")
		clog.Error(errStr)
		return nil, errors.New(errStr)
	}

	// NEVER FORGET THIS
	gitter.user = user

	return gitter, nil
}

func newHubGitter(user string) (Gitter, error) {
	tok := loadGitHubToken(store, user)
	if tok == nil {
		errStr := fmt.Sprintf("can't load gitlab token for user %v, need redirect to authorize.", user)
		clog.Error(errStr)
		return nil, errors.New(errStr)
	}

	gitter := NewGitHub(tok)
	if gitter == nil {
		errStr := fmt.Sprintf("empty Gitter returned, need authoriza.")
		clog.Error(errStr)
		return nil, errors.New(errStr)
	}

	// NEVER FORGET THIS
	gitter.user = user
	return gitter, nil
}
