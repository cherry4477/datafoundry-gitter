package github

import (
	"github.com/google/go-github/github"

	"github.com/asiainfoLDP/datafoundry-gitter/api"
)

type Github struct {
	client *github.Client
	repo   *api.ReposityService
}

func (github Github) ListPersonalRepos() *api.ReposityService {}
