package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/google/go-github/github"
	"github.com/zonesan/clog"
	"golang.org/x/oauth2"
)

type GitHub struct {
	client      *github.Client
	source      string
	oauthtoken  string
	bearertoken string
	owner       string
	repo        string
	ns          string
	bc          string
	user        string
}

func NewGitHub(tok *oauth2.Token) *GitHub {
	hub := new(GitHub)
	hub.source = "github"

	// token, err := oauthConfGitLab.TokenSource(oauth2.NoContext, tok).Token()
	// if err != nil {
	// 	clog.Error("wtf..", err)
	// 	token = tok
	// }

	if tok == nil {
		clog.Error("not authorized yet.")
		return nil
	}

	oauthClient := oauthConf.Client(oauth2.NoContext, tok)

	clog.Debug("token:", tok.AccessToken)

	client := github.NewClient(oauthClient)

	hub.client = client
	hub.oauthtoken = tok.AccessToken

	return hub
}

func (hub *GitHub) ListPersonalRepos(cache bool) *[]Repositories {

	if cache {
		if repos, err := store.LoadReposGithub(hub.User()); err == nil {
			if len(*repos) > 0 {
				return repos
			}
			clog.Warn("cache empty, fetching from remote server.")
		}
	}

	var allRepos []*github.Repository

	opt := &github.RepositoryListOptions{
		ListOptions: github.ListOptions{PerPage: 30},
	}

	for {
		repos, resp, err := hub.client.Repositories.List("", opt)
		if err != nil {
			log.Println(err)
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

	hubRepos := new([]Repositories)

	repos := make(map[OwnerInfo][]Repository)

	for _, v := range allRepos {
		repo := Repository{}
		owner := OwnerInfo{}

		owner.Namespace = *v.Owner.Login
		if *v.Owner.Type == "User" {
			owner.Personal = true
		}
		repo.CloneUrl = *v.CloneURL
		repo.ID = *v.ID
		repo.Name = *v.Name
		repo.Namespace = *v.Owner.Login
		repo.Private = *v.Private
		repo.SshUrl = *v.SSHURL
		repos[owner] = append(repos[owner], repo)
	}

	for k, v := range repos {
		repo := new(Repositories)
		repo.OwnerInfo = k
		repo.Repos = v

		*hubRepos = append(*hubRepos, *repo)
	}

	// debug(hubRepos)

	go func() {
		if len(*hubRepos) > 0 {
			store.SaveReposGithub(hub.User(), hubRepos)
		}
	}()

	return hubRepos

	// for idx, repo := range allRepos {
	// 	fmt.Println(idx, *repo.Owner.Login, *repo.Name, *repo.CloneURL)
	// 	go ListBranches(client, *repo.Owner.Login, *repo.Name)
	// }

}

func (hub *GitHub) ListOrgRepos(org string) { clog.Debug("called.") }

func (hub *GitHub) ListBranches(owner, repo string) *[]Branch {
	var allBranches []*github.Branch
	opt := &github.ListOptions{PerPage: 30}
	for {
		branches, resp, err := hub.client.Repositories.ListBranches(owner, repo, opt)
		if err != nil {
			clog.Error(err)
			return nil
		}
		allBranches = append(allBranches, branches...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
		fmt.Printf("fetch next %v branches, page %v\n", opt.PerPage, resp.NextPage)
	}

	clog.Debugf("Total %d branches.\n", len(allBranches))

	hubBranches := new([]Branch)
	for _, v := range allBranches {
		branch := new(Branch)
		branch.Name = *v.Name
		branch.CommitID = *v.Commit.SHA
		*hubBranches = append(*hubBranches, *branch)
	}

	return hubBranches
}

func (hub *GitHub) ListTags(owner, repo string) { clog.Debug("called.") }
func (hub *GitHub) CreateWebhook(hook *WebHook) *WebHook {

	clog.Debugf("hook info: %#v", hook)

	exist, err := store.GetWebHook(hook.Name)
	if exist != nil {
		return exist
	}

	hook.Source = hub.source
	hook.Pid = ""

	hookname := "web"

	hubhook := new(github.Hook)

	hubhook.Active = enable(yes)
	hubhook.Name = &hookname
	hubhook.Config = make(map[string]interface{})
	hubhook.Config["url"] = hook.URL
	hubhook.Config["content_type"] = "json"
	hubhook.Config["insecure_ssl"] = "1"

	hubhook, resp, err := hub.client.Repositories.CreateHook(hook.Ns, hook.Repo, hubhook)
	_ = resp
	if err != nil {
		clog.Error(err)
		return nil
	}
	hook.ID = *hubhook.ID
	{
		store.CreateWebHook(hook.Name, hook)
	}
	clog.Debugf("created hook github.com/%v (hook id %v)", hook.Name, hook.ID)
	return hook
}

func (hub *GitHub) RemoveWebhook(ns, bc string, id int) error {
	key := ns + "/" + bc
	hook, err := store.GetWebHook(key)
	if hook == nil {
		return errors.New("hook not found.")
	}

	if id != hook.ID {
		clog.Errorf("hook %v mismatch, want remvoe %v, and met %v", hook.Name, id, hook.ID)
		return errors.New("hook id mismatch.")
	}

	if hub.source != hook.Source {
		clog.Errorf("hook %v (id %v) belongs to %v, and met %v", hook.Name, hook.ID, hook.Source, hub.source)
		return errors.New("invalid request.")
	}

	clog.Debugf("remove github.com/%v/%v hook %v (hook id %v)", hook.Ns, hook.Repo, key, hook.ID)
	resp, err := hub.client.Repositories.DeleteHook(hook.Ns, hook.Repo, hook.ID)
	if err != nil {
		clog.Error(err)
		return err
	}
	_ = resp
	return store.DeleteWebHook(key)
}

func (hub *GitHub) CheckWebhook(ns, bc string) *WebHook {
	key := ns + "/" + bc
	hook, err := store.GetWebHook(key)
	if hook == nil {
		clog.Error("hook is nil", err)
		return nil
	}

	/*
		TODO
		should call git api to check if the hook is really exist,
		and if not, remove record from db.
	*/
	return hook
}

func (hub *GitHub) CreateSecret(ns, name string) (*Secret, error) {
	token := hub.GetBearerToken()

	dfClient := NewDataFoundryTokenClient(token)

	data := make(map[string]string)
	data["password"] = hub.GetOauthToken()

	ksecret, err := dfClient.CreateSecret(ns, name, data)
	if err != nil {
		clog.Error(err)
		return nil, err
	}

	secret := new(Secret)
	secret.User = hub.User()
	secret.Secret = ksecret.Name
	secret.Ns = ns
	secret.Available = true

	store.SaveSecretGithub(hub.User(), ns, secret)
	clog.Debugf("%#v,%#v", ksecret, secret)

	return secret, nil
}

func (hub *GitHub) CheckSecret(ns string) *Secret {
	//key := hub.Source() + "/" + hub.User() + "/" + ns
	secret, _ := store.LoadSecretGithub(hub.User(), ns)
	if secret == nil {
		clog.Warn("secret is nil")
	}
	return secret
}

func (hub *GitHub) Source() string {
	return hub.source
}

func (hub *GitHub) User() string {
	return hub.user
}

func (hub *GitHub) GetOauthToken() string {
	return hub.oauthtoken
}

func (hub *GitHub) GetBearerToken() string {
	return hub.bearertoken
}
func (hub *GitHub) SetBearerToken(bearer string) {
	hub.bearertoken = bearer
}

func ListPersonalRepos(client *github.Client, user string) error {

	var allRepos []*github.Repository
	opt := &github.RepositoryListOptions{
		ListOptions: github.ListOptions{PerPage: 30},
	}
	for {
		repos, resp, err := client.Repositories.List("", opt)
		if err != nil {
			log.Println(err)
			return err
		}
		allRepos = append(allRepos, repos...)
		if resp.NextPage == 0 {
			break
		}
		opt.ListOptions.Page = resp.NextPage
		//fmt.Printf("fetch next %d repos\n", resp.NextPage)
	}
	fmt.Printf("Total %d repos.\n", len(allRepos))

	return nil

	// for idx, repo := range allRepos {
	// 	fmt.Println(idx, *repo.Owner.Login, *repo.Name, *repo.CloneURL)
	// 	go ListBranches(client, *repo.Owner.Login, *repo.Name)
	// }

}

func ListOrgRepos(client *github.Client) error {
	var allOrgs []*github.Organization
	opt := &github.ListOptions{PerPage: 30}
	for {
		orgs, resp, err := client.Organizations.List("", opt)
		if err != nil {
			log.Println(err)
			return err
		}
		allOrgs = append(allOrgs, orgs...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
		//fmt.Printf("fetch next %d repos\n", resp.NextPage)
	}
	fmt.Printf("\nTotal %d organization(s).\n", len(allOrgs))
	orgOpt := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: 30},
	}
	var allRepos []*github.Repository
	for _, org := range allOrgs {
		for {
			repos, resp, err := client.Repositories.ListByOrg(*org.Login, orgOpt)
			if err != nil {
				log.Println(err)
				return err
			}
			allRepos = append(allRepos, repos...)
			if resp.NextPage == 0 {
				break
			}
			orgOpt.ListOptions.Page = resp.NextPage
		}
	}

	for idx, repo := range allRepos {
		fmt.Println(idx, *repo.CloneURL)
	}

	d, err := json.MarshalIndent(allOrgs, "", "  ")
	if err != nil {
		fmt.Printf("json.MarshlIndent(allOrgs) failed with %s\n", err)
		return err
	}

	fmt.Printf("Organizations:\n%s\n", string(d))
	return nil
}

func ListBranches(client *github.Client, owner, repo string) error {
	var allBranches []*github.Branch
	opt := &github.ListOptions{PerPage: 30}
	for {
		branches, resp, err := client.Repositories.ListBranches(owner, repo, opt)
		if err != nil {
			log.Println(err)
			return err
		}
		allBranches = append(allBranches, branches...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
		//fmt.Printf("fetch next %d branches\n", resp.NextPage)
	}
	fmt.Printf("\nbranches of %s/%s:\n", owner, repo)
	for _, branch := range allBranches {
		fmt.Println(*branch.Name)
	}
	return ListTags(client, owner, repo)
}

func ListTags(client *github.Client, owner, repo string) error {
	var allTags []*github.RepositoryTag
	opt := &github.ListOptions{PerPage: 30}
	for {
		tags, resp, err := client.Repositories.ListTags(owner, repo, opt)
		if err != nil {
			log.Println(err)
			return err
		}
		allTags = append(allTags, tags...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
		//fmt.Printf("fetch next %d tags\n", resp.NextPage)
	}
	fmt.Printf("\ntags of %s/%s:\n", owner, repo)
	for _, tag := range allTags {
		fmt.Println(*tag.Name)
	}
	return nil
}
func UserProfile(client *github.Client, username string) error {
	user, _, err := client.Users.Get("")
	if err != nil {
		fmt.Printf("client.Users.Get() faled with '%s'\n", err)
		//http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return err
	}
	fmt.Printf("Logged in as GitHub user: %s\n", *user.Login)

	d, err := json.MarshalIndent(user, "", "  ")
	if err != nil {
		fmt.Printf("json.MarshlIndent(user) failed with %s\n", err)
		return err
	}

	fmt.Printf("User:\n%s\n", string(d))
	return nil

}
