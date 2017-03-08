package main

import (
	"github.com/zonesan/clog"
	"golang.org/x/oauth2"
)

type Storage interface {
	LoadTokenGitlab(user string) (*oauth2.Token, error)
	SaveTokenGitlab(tok *oauth2.Token) error
	LoadTokenGithub(user string) (*oauth2.Token, error)
	SaveTokenGithub(tok *oauth2.Token) error
}

type RedisStore struct {
}

func (rs *RedisStore) LoadTokenGitlab(user string) (*oauth2.Token, error) {
	clog.Debug("called.")
	return nil, nil
}

func (rs *RedisStore) SaveTokenGitlab(tok *oauth2.Token) error {
	clog.Debug("called.")
	return nil
}

func (rs *RedisStore) LoadTokenGithub(user string) (*oauth2.Token, error) {
	clog.Debug("called.")
	return nil, nil
}

func (rs *RedisStore) SaveTokenGithub(tok *oauth2.Token) error {
	clog.Debug("called.")
	return nil
}
