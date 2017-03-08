package main

import (
	"github.com/zonesan/clog"
)

type GitLab struct {
}

func (gitlab *GitLab) ListPersonalRepos(user string)   { clog.Debug("called.") }
func (gitlab *GitLab) ListOrgRepos(org string)         { clog.Debug("called.") }
func (gitlab *GitLab) ListBranches(owner, repo string) { clog.Debug("called.") }
func (gitlab *GitLab) ListTags(owner, repo string)     { clog.Debug("called.") }
func (gitlab *GitLab) CreateWebhook(hook interface{})  { clog.Debug("called.") }
func (gitlab *GitLab) RemoveWebhook(hook interface{})  { clog.Debug("called.") }
func (gitlab *GitLab) CheckWebhook(hook interface{})   { clog.Debug("called.") }
