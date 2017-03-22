package main

import (
	"encoding/json"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/zonesan/clog"
	"golang.org/x/oauth2"
	"sync"
	"time"
)

func gitlabUserAuthTokenKey(user string) string {
	return "gitter://gitlab/oauthtoken/" + user
}
func githubUserAuthTokenKey(user string) string {
	return "gitter://github/oauthtoken/" + user
}
func webhookKey(key string) string {
	return "gitter://webhook/" + key
}
func gitlabUserSecretKey(namespace, user string) string {
	return "gitter://gitlab/secret/" + namespace + "/" + user
}
func githubUserSecretKey(namespace, user string) string {
	return "gitter://github/secret/" + namespace + "/" + user
}
func gitlabUserSSHKeyPairKey(user string) string {
	return "gitter://gitlab/sshkey/" + user
}
func githubUserReposKey(user string) string {
	return "gitter://github/repos/" + user
}
func gitlabUserReposKey(user string) string {
	return "gitter://gitlab/repos/" + user
}

//=======================================================
// Storage
//=======================================================

type storageError string

func (se storageError) Error() string {
	return string(se)
}

const (
	StorageErr_NotFound storageError = "key not found or value is nil"
)

type Storage interface {
	LoadTokenGitlab(user string) (*oauth2.Token, error)
	SaveTokenGitlab(user string, tok *oauth2.Token) error
	LoadTokenGithub(user string) (*oauth2.Token, error)
	SaveTokenGithub(user string, tok *oauth2.Token) error

	GetWebHook(webhook string) (*WebHook, error)
	CreateWebHook(webhook string, hook *WebHook) error
	DeleteWebHook(webhook string) error

	LoadSecretGitlab(user, ns string) (*Secret, error)
	SaveSecretGitlab(user, ns string, secret *Secret) error
	LoadSecretGithub(user, ns string) (*Secret, error)
	SaveSecretGithub(user, ns string, secret *Secret) error

	LoadSSHKeyGitlab(user string) (*SSHKey, error)
	SaveSSHKeyGitlab(user string, key *SSHKey) error

	LoadReposGitlab(user string) (*[]Repositories, error)
	SaveReposGitlab(user string, repos *[]Repositories) error
	LoadReposGithub(user string) (*[]Repositories, error)
	SaveReposGithub(user string, repos *[]Repositories) error
}

var _ Storage = &storage{}

type KeyValueStorager interface {
	Set(key string, value []byte) error
	Get(key string) ([]byte, error)
	//List(keyPrefix string)(<-chan struct{key, value []byte}, chan<- struct{})
	Delete(key string) error
}

var _ KeyValueStorager = &memoryStorager{}
var _ KeyValueStorager = &redisStorager{}

func NewStorage(kv KeyValueStorager) Storage {
	return &storage{kv}
}

//=======================================================
// storage
//=======================================================

type storage struct {
	KeyValueStorager
}

func (s *storage) Load(key string, into interface{}) error {
	data, err := s.Get(key)
	if err != nil {
		clog.Errorf("load (%s) error: %s", key, err)
		return err
	}

	// todo: ok?
	if data == nil {
		clog.Warn(StorageErr_NotFound)
		return StorageErr_NotFound
	}

	err = json.Unmarshal(data, into)
	if err != nil {
		clog.Errorf("unmarshal (%s) data (%s) error: %s", key, string(data), err)
		return err
	}

	return nil
}

func (s *storage) Save(key string, obj interface{}) error {
	data, err := json.Marshal(obj)
	if err != nil {
		clog.Errorf("marshal (%s) auth token (%v) error: %s", key, obj, err)
		return err
	}

	err = s.Set(key, data)
	if err != nil {
		clog.Errorf("save (%s) error: %s", key, err)
		return err
	}

	return nil
}

//========== implements Storage interface

func (s *storage) LoadTokenGitlab(user string) (*oauth2.Token, error) {
	var token = &oauth2.Token{}
	if err := s.Load(gitlabUserAuthTokenKey(user), token); err != nil {
		return nil, err
	}

	return token, nil
}

func (s *storage) SaveTokenGitlab(user string, tok *oauth2.Token) error {
	return s.Save(gitlabUserAuthTokenKey(user), tok)
}

func (s *storage) LoadTokenGithub(user string) (*oauth2.Token, error) {
	var token = &oauth2.Token{}
	if err := s.Load(githubUserAuthTokenKey(user), token); err != nil {
		return nil, err
	}

	return token, nil
}

func (s *storage) SaveTokenGithub(user string, tok *oauth2.Token) error {
	return s.Save(githubUserAuthTokenKey(user), tok)
}

func (s *storage) GetWebHook(webhook string) (*WebHook, error) {
	var hook = &WebHook{}
	if err := s.Load(webhookKey(webhook), hook); err != nil {
		return nil, err
	}

	return hook, nil
}

func (s *storage) CreateWebHook(webhook string, hook *WebHook) error {
	return s.Save(webhookKey(webhook), hook)
}

func (s *storage) DeleteWebHook(webhook string) error {
	return s.Delete(webhookKey(webhook))
}

func (s *storage) LoadSecretGitlab(user, ns string) (*Secret, error) {
	var secret = &Secret{}
	if err := s.Load(gitlabUserSecretKey(ns, user), secret); err != nil {
		return nil, err
	}

	return secret, nil
}

func (s *storage) SaveSecretGitlab(user, ns string, secret *Secret) error {
	return s.Save(gitlabUserSecretKey(ns, user), secret)
}

func (s *storage) LoadSecretGithub(user, ns string) (*Secret, error) {
	var secret = &Secret{}
	if err := s.Load(githubUserSecretKey(ns, user), secret); err != nil {
		return nil, err
	}

	return secret, nil
}

func (s *storage) SaveSecretGithub(user, ns string, secret *Secret) error {
	return s.Save(githubUserSecretKey(ns, user), secret)
}

func (s *storage) LoadSSHKeyGitlab(user string) (*SSHKey, error) {
	var key = &SSHKey{}
	if err := s.Load(gitlabUserSSHKeyPairKey(user), key); err != nil {
		return nil, err
	}

	return key, nil
}

func (s *storage) SaveSSHKeyGitlab(user string, key *SSHKey) error {
	return s.Save(gitlabUserSSHKeyPairKey(user), key)
}

func (s *storage) LoadReposGitlab(user string) (*[]Repositories, error) {
	repos := new([]Repositories)
	if err := s.Load(gitlabUserReposKey(user), repos); err != nil {
		return nil, err
	}
	return repos, nil
}

func (s *storage) SaveReposGitlab(user string, repos *[]Repositories) error {
	return s.Save(gitlabUserReposKey(user), repos)
}

func (s *storage) LoadReposGithub(user string) (*[]Repositories, error) {
	repos := new([]Repositories)
	if err := s.Load(githubUserReposKey(user), repos); err != nil {
		return nil, err
	}
	return repos, nil
}

func (s *storage) SaveReposGithub(user string, repos *[]Repositories) error {
	return s.Save(githubUserReposKey(user), repos)
}

//=======================================================
// memory kv
//=======================================================

type memoryStorager struct {
	sync.RWMutex
	m map[string][]byte
}

func NewMemoryKeyValueStorager() KeyValueStorager {
	return &memoryStorager{m: map[string][]byte{}}
}

func (ms *memoryStorager) Set(key string, value []byte) error {
	ms.Lock()
	ms.m[key] = value
	ms.Unlock()
	return nil
}

func (ms *memoryStorager) Get(key string) ([]byte, error) {
	ms.RLock()
	var value, present = ms.m[key]
	ms.RUnlock()
	if present {
		return value, nil
	} else {
		return nil, StorageErr_NotFound
	}
}

func (ms *memoryStorager) Delete(key string) error {
	ms.Lock()
	delete(ms.m, key)
	ms.Unlock()
	return nil
}

//=======================================================
// redis kv
//=======================================================

type redisStorager struct {
	pool *redis.Pool
}

// addr format is host:port.
// clusterName is blank means addr is the master address, otherwise addr is sentinel address.
func NewRedisKeyValueStorager(addr, clusterName, password string) KeyValueStorager {
	var p = &redis.Pool{
		MaxIdle:     10,
		MaxActive:   10,
		Wait:        true,
		IdleTimeout: 4 * time.Minute,
		Dial: func() (redis.Conn, error) {
			var masterAddr string
			if clusterName == "" {
				masterAddr = addr
			} else {
				// query master addr from sentinel
				err := func() error {
					conn, err := redis.DialTimeout("tcp", addr, time.Second*10, time.Second*10, time.Second*10)
					if err != nil {
						clog.Errorf("dial redis sentinel (%s) error: %s", addr, err)
						return err
					}
					defer conn.Close()

					redisMasterPair, err := redis.Strings(conn.Do("SENTINEL", "get-master-addr-by-name", clusterName))
					if err != nil {
						clog.Error("redis sentinel get-master-addr-by-name error:", err)
						return err
					}

					if len(redisMasterPair) != 2 {
						clog.Errorf("redis sentinel get-master-addr-by-name result invalid: %v", redisMasterPair)
						return fmt.Errorf("redis sentinel get-master-addr-by-name result invalid: %v", redisMasterPair)
					}

					masterAddr = redisMasterPair[0] + ":" + redisMasterPair[1]
					return nil
				}()

				if err != nil {
					return nil, err
				}
			}

			// dial master
			c, err := redis.Dial("tcp", masterAddr)
			if err != nil {
				clog.Errorf("dial redis master (%s) error: %s", addr, err)
				return nil, err
			}
			if password != "" {
				if _, err := c.Do("AUTH", password); err != nil {
					clog.Errorf("redis master (%s) AUTH (%s) error: %s", addr, password, err)
					c.Close()
					return nil, err
				}
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}

	return &redisStorager{pool: p}
}

func (rs *redisStorager) Set(key string, value []byte) error {
	c := rs.pool.Get()
	defer c.Close()

	if _, err := c.Do("SET", key, value); err != nil {
		clog.Error("[SET] err :", err)
		return err
	}

	return nil
}

func (rs *redisStorager) Get(key string) ([]byte, error) {
	c := rs.pool.Get()
	defer c.Close()

	b, err := redis.Bytes(c.Do("GET", key))
	if err != nil {
		if err == redis.ErrNil {
			return nil, StorageErr_NotFound
		}

		clog.Error("[GET] err:", err)
		return nil, err
	}

	return b, nil
}

func (rs *redisStorager) Delete(key string) error {
	c := rs.pool.Get()
	defer c.Close()

	if _, err := c.Do("DEL", key); err != nil {
		clog.Error("[DEL] err:", err)
		return err
	}

	return nil
}
