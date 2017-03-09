package main

import (
	"encoding/json"
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

	if tok == nil {
		clog.Error("not authorized yet.")
		return nil
	}

	oauthClient := oauthConfGitLab.Client(oauth2.NoContext, tok)

	client := gitlab.NewOAuthClient(oauthClient, tok.AccessToken)
	client.SetBaseURL(gitlabBaseURL + "/api/v3")

	lab.client = client

	return lab

}

func (lab *GitLab) ListPersonalRepos(user string) {
	clog.Debugf("list repos of %s called. on progress, nothing to show.", user)

	var allRepos []*gitlab.Project

	opt := &gitlab.ListProjectsOptions{
		ListOptions: gitlab.ListOptions{PerPage: 30},
	}

	for {
		repos, resp, err := lab.client.Projects.ListProjects(opt)
		if err != nil {
			clog.Error(err)
			return
		}
		allRepos = append(allRepos, repos...)
		if resp.NextPage == 0 {
			break
		}
		opt.ListOptions.Page = resp.NextPage
		//fmt.Printf("fetch next %d repos\n", resp.NextPage)
	}
	clog.Infof("Total %d repos.\n", len(allRepos))

	d, err := json.MarshalIndent(allRepos, "", "  ")
	if err != nil {
		clog.Error("json.MarshlIndent(allRepos) failed with %s\n", err)
		return
	}

	clog.Printf("Repos:\n%s\n", string(d))
	_ = d

}

func (lab *GitLab) ListOrgRepos(org string)         { clog.Debug("called.") }
func (lab *GitLab) ListBranches(owner, repo string) { clog.Debug("called.") }
func (lab *GitLab) ListTags(owner, repo string)     { clog.Debug("called.") }
func (lab *GitLab) CreateWebhook(hook interface{})  { clog.Debug("called.") }
func (lab *GitLab) RemoveWebhook(hook interface{})  { clog.Debug("called.") }
func (lab *GitLab) CheckWebhook(hook interface{})   { clog.Debug("called.") }
