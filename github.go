package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/go-github/github"
	"github.com/zonesan/clog"
	"golang.org/x/oauth2"
)

type GitHub struct {
	client *github.Client
}

func NewGitHub(tok *oauth2.Token) *GitHub {
	hub := new(GitHub)

	// token, err := oauthConfGitLab.TokenSource(oauth2.NoContext, tok).Token()
	// if err != nil {
	// 	clog.Error("wtf..", err)
	// 	token = tok
	// }

	oauthClient := oauthConf.Client(oauth2.NoContext, tok)

	client := github.NewClient(oauthClient)

	hub.client = client

	return hub
}

func (hub *GitHub) ListPersonalRepos(user string) *[]Repositories {
	clog.Debug("called.")

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
		clog.Debugf("fetch next %d repos\n", resp.NextPage)
	}
	fmt.Printf("Total %d repos.\n", len(allRepos))

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

	debug(hubRepos)

	return hubRepos

	// for idx, repo := range allRepos {
	// 	fmt.Println(idx, *repo.Owner.Login, *repo.Name, *repo.CloneURL)
	// 	go ListBranches(client, *repo.Owner.Login, *repo.Name)
	// }

}

func (hub *GitHub) ListOrgRepos(org string)         { clog.Debug("called.") }
func (hub *GitHub) ListBranches(owner, repo string) { clog.Debug("called.") }
func (hub *GitHub) ListTags(owner, repo string)     { clog.Debug("called.") }
func (hub *GitHub) CreateWebhook(hook interface{})  { clog.Debug("called.") }
func (hub *GitHub) RemoveWebhook(hook interface{})  { clog.Debug("called.") }
func (hub *GitHub) CheckWebhook(hook interface{})   { clog.Debug("called.") }

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
