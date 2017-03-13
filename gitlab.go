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

func (lab *GitLab) ListPersonalRepos(user string) *[]Repositories {
	// clog.Debugf("list repos of %s called.", user)

	var allRepos []*gitlab.Project

	opt := &gitlab.ListProjectsOptions{
		ListOptions: gitlab.ListOptions{PerPage: 30},
	}

	for {
		repos, resp, err := lab.client.Projects.ListProjects(opt)
		if err != nil {
			clog.Error(err)
			return nil
		}
		allRepos = append(allRepos, repos...)
		if resp.NextPage == 0 {
			break
		}
		opt.ListOptions.Page = resp.NextPage
		clog.Debugf("fetch next %v repos, page %v\n", opt.ListOptions.PerPage, resp.NextPage)
	}
	clog.Debugf("Total %d repos.\n", len(allRepos))

	labRepos := new([]Repositories)

	repos := make(map[OwnerInfo][]Repository)

	for _, v := range allRepos {
		repo := Repository{}
		owner := OwnerInfo{}

		owner.Namespace = v.Namespace.Name
		if v.Owner != nil {
			owner.Personal = true
		}
		repo.CloneUrl = v.HTTPURLToRepo
		repo.ID = v.ID
		repo.Name = v.Name
		repo.Namespace = v.Namespace.Name
		repo.Private = !v.Public
		repo.SshUrl = v.SSHURLToRepo
		repos[owner] = append(repos[owner], repo)
	}

	for k, v := range repos {
		repo := new(Repositories)
		repo.OwnerInfo = k
		repo.Repos = v

		*labRepos = append(*labRepos, *repo)
	}

	//debug(labRepos)

	return labRepos

}

func (lab *GitLab) ListOrgRepos(org string) { clog.Debug("called.") }

func (lab *GitLab) ListBranches(owner, repo string) *[]Branch {
	branches, resp, err := lab.client.Branches.ListBranches(repo)
	_ = resp
	if err != nil {
		clog.Error(err)
		return nil
	}
	clog.Debugf("total %v branches.", len(branches))

	labBranches := new([]Branch)
	for _, v := range branches {
		branch := new(Branch)
		branch.Name = v.Name
		branch.CommitID = v.Commit.ID
		*labBranches = append(*labBranches, *branch)
	}

	return labBranches

}

func (lab *GitLab) ListTags(owner, repo string)    { clog.Debug("called.") }
func (lab *GitLab) CreateWebhook(hook interface{}) { clog.Debug("called.") }
func (lab *GitLab) RemoveWebhook(hook interface{}) { clog.Debug("called.") }
func (lab *GitLab) CheckWebhook(hook interface{})  { clog.Debug("called.") }
