package main

import (
	"fmt"
	"time"
	"encoding/json"
	"github.com/zonesan/clog"
	"golang.org/x/oauth2"
	"github.com/garyburd/redigo/redis"
)

func gitlabUserAuthTokenKey(user string) string {
	return "gitlab/oauthtoken/" + user
}
func githubUserAuthTokenKey(user string) string {
	return "github/oauthtoken/" + user
}

//=======================================================
// Storage
//=======================================================

type Storage interface {
	LoadTokenGitlab(user string) *oauth2.Token
	SaveTokenGitlab(user string, tok *oauth2.Token) error
	LoadTokenGithub(user string) *oauth2.Token
	SaveTokenGithub(user string, tok *oauth2.Token) error
}

type KeyValueStorager interface {
	Set(key string, value []byte) error
	Get(key string) ([]byte, error)
	//List(keyPrefix string)(<-chan struct{key, value []byte}, chan<- struct{})
	Delete(key string) error
}

func NewStorage(kv KeyValueStorager) Storage {
	return &storage{kv}
}

//=======================================================
// KeyValueStorage
//=======================================================

type storage struct {
	KeyValueStorager
}

func (s *storage) LoadTokenGitlab(user string) *oauth2.Token {
	userKey := gitlabUserAuthTokenKey(user)

	data, err := s.Get(userKey)
	if err == nil {
		clog.Debugf("get gitlab user (%s) auth token error: %s", user, err)
		return nil
	}

	var t = &oauth2.Token{}
	err = json.Unmarshal(data, t)
	if err == nil {
		clog.Debugf("unmarshal gitlab user (%s) auth token (%s) error: %s", user, string(data), err)
		return nil
	}

	return t
}

func (s *storage) SaveTokenGitlab(user string, tok *oauth2.Token) error {
	data, err := json.Marshal(tok)
	if err == nil {
		clog.Debugf("marshal gitlab user (%s) auth token (%v) error: %s", user, tok, err)
		return err
	}

	userKey := gitlabUserAuthTokenKey(user)

	err = s.Set(userKey, data)
	if err != nil {
		clog.Debugf("set gitlab user (%s) auth token error: %s", user, err)
		return err
	}

	return nil
}

func (s *storage) LoadTokenGithub(user string) *oauth2.Token {
	userKey := githubUserAuthTokenKey(user)

	data, err := s.Get(userKey)
	if err == nil {
		clog.Debugf("get gitlab user (%s) auth token error: %s", user, err)
		return nil
	}

	var t = &oauth2.Token{}
	err = json.Unmarshal(data, t)
	if err == nil {
		clog.Debugf("unmarshal gitlab user (%s) auth token (%s) error: %s", user, string(data), err)
		return nil
	}

	return t
}

func (s *storage) SaveTokenGithub(user string, tok *oauth2.Token) error {
	data, err := json.Marshal(tok)
	if err == nil {
		clog.Debugf("marshal gitlab user (%s) auth token (%v) error: %s", user, tok, err)
		return err
	}

	userKey := githubUserAuthTokenKey(user)

	err = s.Set(userKey, data)
	if err != nil {
		clog.Debugf("set gitlab user (%s) auth token error: %s", user, err)
		return err
	}

	return nil
}

//=======================================================
// memory kv
//=======================================================

type memoryStorager struct {
	m map[string][]byte
}

func NewMemoryKeyValueStorager() KeyValueStorager {
	return &memoryStorager{map[string][]byte{}}
}

func (ms *memoryStorager) Set(key string, value []byte) error {
	ms.m[key] = value
	return nil
}

func (ms *memoryStorager) Get(key string) ([]byte, error) {
	return ms.m[key], nil
}

func (ms *memoryStorager) Delete(key string) error {
	delete(ms.m, key)
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

func (ms *redisStorager) Set(key string, value []byte) error {
	c := ms.pool.Get()
	defer c.Close()

	if _, err := c.Do("SET", key, value); err != nil {
		clog.Error("[SET] err :", err)
		return err
	}

	return nil
}

func (ms *redisStorager) Get(key string) ([]byte, error) {
	c := ms.pool.Get()
	defer c.Close()

	b, err := redis.Bytes(c.Do("GET", key))
	if err != nil && err != redis.ErrNil {
		clog.Error("[GET] err:", err)
		return nil, err
	}

	return b, nil
}

func (ms *redisStorager) Delete(key string) error {
	c := ms.pool.Get()
	defer c.Close()

	if _, err := c.Do("DEL", key); err != nil {
		clog.Error("[DEL] err:", err)
		return err
	}

	return nil
}