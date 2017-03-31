package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"
	"time"

	gitlab "github.com/xanzy/go-gitlab"
	"github.com/zonesan/clog"
	"golang.org/x/crypto/ssh"
	"golang.org/x/oauth2"
)

type GitLab struct {
	client      *gitlab.Client
	source      string
	oauthtoken  string
	bearertoken string
	repoid      string
	ns          string
	bc          string
	user        string
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
	clog.Debug("token:", tok.AccessToken)

	client := gitlab.NewOAuthClient(oauthClient, tok.AccessToken)
	client.SetBaseURL(gitlabBaseURL + "/api/v3")

	lab.client = client
	lab.oauthtoken = tok.AccessToken

	return lab

}

func (lab *GitLab) ListPersonalRepos(cache bool) *[]Repositories {
	// clog.Debugf("list repos of %s called.", user)
	if cache {
		if repos, err := store.LoadReposGitlab(lab.User()); err == nil {
			if len(*repos) > 0 {
				return repos
			}
			clog.Warn("cache empty, fetching from remote server.")
		}
	}

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
		repo.CloneURL = v.HTTPURLToRepo
		repo.ID = v.ID
		repo.Name = v.Name
		repo.Namespace = v.Namespace.Name
		repo.Private = !v.Public
		repo.SSHURL = v.SSHURLToRepo
		repos[owner] = append(repos[owner], repo)
	}

	for k, v := range repos {
		repo := new(Repositories)
		repo.OwnerInfo = k
		repo.Repos = v

		*labRepos = append(*labRepos, *repo)
	}

	//debug(labRepos)

	go func() {
		if len(*labRepos) > 0 {
			store.SaveReposGitlab(lab.User(), labRepos)
		}
	}()

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
		return errors.New("lab hook not found")
	}
	if id != hook.ID {
		clog.Errorf("hook %v mismatch, want remvoe %v, and met %v", hook.Name, id, hook.ID)
		return errors.New("lab hook id mismatch")
	}

	if lab.source != hook.Source {
		clog.Errorf("hook %v (id %v) belongs to %v, and met %v", hook.Name, hook.ID, hook.Source, lab.source)
		return errors.New("invalid request")
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

func (lab *GitLab) deploySSHPubKey(pubkey string) (*gitlab.SSHKey, error) {
	title := "datafoundry-pull-secret"
	opt := &gitlab.AddSSHKeyOptions{
		Title: &title,
		Key:   &pubkey,
	}
	sshkey, resp, err := lab.client.Users.AddSSHKey(opt)
	_ = resp
	return sshkey, err
}

func (lab *GitLab) getSSHPrivateKey() (string, error) {
	return "", nil
}

func (lab *GitLab) generateSSHeKeyPair() (*SSHKey, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	err = priv.Validate()
	if err != nil {
		return nil, err
	}

	privDer := x509.MarshalPKCS1PrivateKey(priv)

	// pem.Block
	// blk pem.Block
	privBlk := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   privDer,
	}

	// Resultant private key in PEM format.
	// priv_pem string
	privateKey := string(pem.EncodeToMemory(&privBlk))
	//println("private:", privateKey)

	pub := priv.PublicKey

	sshpub, err := ssh.NewPublicKey(&pub)
	if err != nil {
		return nil, err
	}
	publicKey := string(ssh.MarshalAuthorizedKey(sshpub))
	publicKey = strings.TrimRight(publicKey, "\n")
	publicKey = fmt.Sprintf("%s rsa-key-%s", publicKey, time.Now().Format("20060102"))
	//println("public:", publicKey)

	key := new(SSHKey)
	key.Owner = &lab.user
	key.Pubkey = &publicKey
	key.Privkey = &privateKey

	return key, nil

}

func (lab *GitLab) CreateSecret(ns, name string) (*Secret, error) {
	token := lab.GetBearerToken()
	dfClient := NewDataFoundryTokenClient(token)

	key, err := store.LoadSSHKeyGitlab(lab.User())
	if err != nil {
		key, err = lab.generateSSHeKeyPair()
		if err != nil {
			clog.Error(err)
			return nil, err
		}
		sshkey, err := lab.deploySSHPubKey(*key.Pubkey)
		if err != nil {
			clog.Error(err)
			return nil, err
		}

		key.ID = sshkey.ID
		key.CreatedAt = sshkey.CreatedAt
		// clog.Debugf("sshkey: %#v", sshkey)
		clog.Debugf("key: %#v", key)

		store.SaveSSHKeyGitlab(lab.User(), key)
	}

	data := make(map[string]string)
	data["ssh-privatekey"] = *key.Privkey

	ksecret, err := dfClient.CreateSecret(ns, name, data)
	if err != nil {
		clog.Error(err)
		return nil, err
	}
	secret := new(Secret)
	secret.User = lab.User()
	secret.Secret = ksecret.Name
	secret.Ns = ns
	secret.Available = true
	store.SaveSecretGitlab(lab.User(), ns, secret)
	//clog.Debugf("%#v,%#v", ksecret, secret)

	return secret, nil
}

func (lab *GitLab) CheckSecret(ns string) *Secret {
	//key := hub.Source() + "/" + hub.User() + "/" + ns
	secret, _ := store.LoadSecretGitlab(lab.User(), ns)
	if secret == nil {
		clog.Warn("secret is nil")
	}
	return secret
}

func (lab *GitLab) Source() string {
	return lab.source
}

func (lab *GitLab) User() string {
	return lab.user
}

func (lab *GitLab) GetOauthToken() string {
	return lab.oauthtoken
}
func (lab *GitLab) GetBearerToken() string {
	return lab.bearertoken
}
func (lab *GitLab) SetBearerToken(bearer string) {
	lab.bearertoken = bearer
}
