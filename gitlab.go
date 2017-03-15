package main

import (
	"errors"

	gitlab "github.com/xanzy/go-gitlab"
	"github.com/zonesan/clog"
	"golang.org/x/oauth2"
)

type GitLab struct {
	client *gitlab.Client
	source string
	repoid string
	ns     string
	bc     string
}

func NewGitLab(tok *oauth2.Token) *GitLab {
	lab := new(GitLab)
	lab.source = "gitlab"

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

func (lab *GitLab) ListTags(owner, repo string) { clog.Debug("called.") }
func (lab *GitLab) CreateWebhook(hook *WebHook) *WebHook {
	clog.Debugf("hook info: %#v", hook)

	exist, err := store.GetWebHook(hook.Name)
	if exist != nil {
		return exist
	}

	hook.Source = lab.source
	hook.Ns = ""
	hook.Repo = ""

	opt := &gitlab.AddProjectHookOptions{
		URL:                   &hook.URL,
		PushEvents:            enable(yes),
		TagPushEvents:         enable(yes),
		EnableSSLVerification: enable(no),
	}
	labhook, resp, err := lab.client.Projects.AddProjectHook(hook.Pid, opt)
	_ = resp
	if err != nil {
		clog.Error(err)
		return nil
	}
	hook.ID = labhook.ID
	{
		store.CreateWebHook(hook.Name, hook)
	}
	return hook
}

func (lab *GitLab) RemoveWebhook(ns, bc string, id int) error {
	key := ns + "/" + bc
	hook, err := store.GetWebHook(key)
	if hook == nil {
		return errors.New("hook not found.")
	}
	if id != hook.ID {
		clog.Errorf("hook %v mismatch, want remvoe %v, and met %v", hook.Name, id, hook.ID)
		return errors.New("hook id mismatch.")
	}

	if lab.source != hook.Source {
		clog.Errorf("hook %v (id %v) belongs to %v, and met %v", hook.Name, hook.ID, hook.Source, lab.source)
		return errors.New("invalid request.")
	}

	clog.Debugf("remove gitlab project %v hook %v (hook id %v)", hook.Pid, key, hook.ID)
	resp, err := lab.client.Projects.DeleteProjectHook(hook.Pid, hook.ID)
	// resp, err := hub.client.Repositories.DeleteHook(hook.Ns, hook.Repo, hook.ID)
	if err != nil {
		clog.Error(err)
		return err
	}
	_ = resp
	return store.DeleteWebHook(key)
}

func (lab *GitLab) CheckWebhook(ns, bc string) *WebHook {

	key := ns + "/" + bc
	hook, err := store.GetWebHook(key)
	clog.Debugf("checking %v hook of %v", lab.source, key)
	if hook == nil {
		clog.Error("hook is nil:", err)
		return nil
	}
	clog.Debugf("%v %v, id: %v", hook.Source, hook.Name, hook.ID)
	/*
		TODO
		should call git api to check if the hook is really exist,
		and if not, remove record from db.
		should check err of store.GetWebHook(key) ?
	*/
	return hook
}
