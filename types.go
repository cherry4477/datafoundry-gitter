package main

import (
	"time"
)

type Gitter interface {
	Source() string
	User() string
	ListPersonalRepos(cache bool) *[]Repositories
	ListOrgRepos(org string)
	ListBranches(owner, repo string) *[]Branch
	ListTags(owner, repo string)
	CreateWebhook(hook *WebHook) *WebHook
	RemoveWebhook(ns, bc string, id int) error
	CheckWebhook(ns, bc string) *WebHook
	CreateSecret(ns, secret string) (*Secret, error)
	CheckSecret(ns string) *Secret
	GetOauthToken() string
	GetBearerToken() string
	SetBearerToken(bearer string)
}

type OwnerInfo struct {
	Namespace string `json:"namespace"`
	Personal  bool   `json:"personal"`
}

type Repositories struct {
	OwnerInfo
	Repos []Repository `json:"repos"`
}

type Repository struct {
	ID        int    `json:"id,omitempty"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	CloneURL  string `json:"clone_url"`
	SSHURL    string `json:"ssh_clone_url,omitempty"`
	Private   bool   `json:"private"`
	//FullName  string `json:"full_name"`
	//Branches []Branch `json:"branches"`
	//Tags     []Tag    `json:"tags"`
}

type Tag struct {
	Name     string `json:"name"`
	CommitID string `json:"commitid"`
}

type Branch struct {
	Name     string `json:"name"`
	CommitID string `json:"commitid"`
}

type WebHook struct {
	ID int `json:"id"`
	// namespace + '/' + buildconfig
	Name      string `json:"name,omitempty"`
	Source    string `json:"source,omitempty"`
	hookParam `json:"params"`
}

type hookParam struct {
	Ns   string `json:"ns,omitempty"`
	Repo string `json:"repo,omitempty"`
	Pid  string `json:"id,omitempty"`
	URL  string `json:"url"`
}

type SSHKey struct {
	ID        int
	Owner     *string
	Pubkey    *string
	Privkey   *string
	CreatedAt *time.Time
}

type Secret struct {
	Ns        string `json:"namespace"`
	User      string `json:"user"`
	Secret    string `json:"secret"`
	Available bool   `json:"available"`
}

var (
	yes = true
	no  = false
)

func enable(b bool) *bool {
	return &b
}

type APIResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Reason  string `json:"reason,omitempty"`
	status  int    `json:"status,omitempty"`
}
