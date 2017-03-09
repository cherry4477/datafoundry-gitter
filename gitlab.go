package main

import (
	gitlab "github.com/xanzy/go-gitlab"
	"github.com/zonesan/clog"
	"golang.org/x/oauth2"
)

type GitLab struct {
	client *gitlab.Client
}

func NewGitLab(tok *oauth2.Token) *GitLab {
	lab := new(GitLab)

	// token, err := oauthConfGitLab.TokenSource(oauth2.NoContext, tok).Token()
	// if err != nil {
	// 	clog.Error("wtf..", err)
	// 	token = tok
	// }

	oauthClient := oauthConfGitLab.Client(oauth2.NoContext, tok)

	client := gitlab.NewOAuthClient(oauthClient, tok.AccessToken)

	lab.client = client

	return lab

}

func (gitlab *GitLab) ListPersonalRepos(user string)   { clog.Debug("called.") }
func (gitlab *GitLab) ListOrgRepos(org string)         { clog.Debug("called.") }
func (gitlab *GitLab) ListBranches(owner, repo string) { clog.Debug("called.") }
func (gitlab *GitLab) ListTags(owner, repo string)     { clog.Debug("called.") }
func (gitlab *GitLab) CreateWebhook(hook interface{})  { clog.Debug("called.") }
func (gitlab *GitLab) RemoveWebhook(hook interface{})  { clog.Debug("called.") }
func (gitlab *GitLab) CheckWebhook(hook interface{})   { clog.Debug("called.") }
