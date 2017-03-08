package main

type Gitter interface {
	ListPersonalRepos(user string)
	ListOrgRepos(org string)
	ListBranches(owner, repo string)
	ListTags(owner, repo string)
	CreateWebhook(hook interface{})
	RemoveWebhook(hook interface{})
	CheckWebhook(hook interface{})
}

type ReposityService struct {
	Login     string     `json:"login"`
	AvatorUrl string     `json:"avator_url"`
	Type      string     `json:"type"`
	Reposites []Reposity `json:"repos"`
}

type Reposity struct {
	Name     string `json:"repo_name"`
	CloneUrl string `json:"clone_url"`
	FullName string `json:"full_name"`
	Private  bool   `json:"private"`
	//Branches []Branch `json:"branches"`
	//Tags     []Tag    `json:"tags"`
}

type Tag struct {
	Name     string `json:"tag"`
	CommitID string `json:"commitid"`
}

type Branch struct {
	Name     string `json:"branch"`
	CommitID string `json:"commitid"`
}
