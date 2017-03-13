package main

import (
	"fmt"
	"time"
	"github.com/zonesan/clog"
	"golang.org/x/oauth2"
	"github.com/garyburd/redigo/redis"
)

var labStore = make(map[string]*oauth2.Token)
var hubStore = make(map[string]*oauth2.Token)

type Storage interface {
	LoadTokenGitlab(user string) *oauth2.Token
	SaveTokenGitlab(user string, tok *oauth2.Token) error
	LoadTokenGithub(user string) *oauth2.Token
	SaveTokenGithub(user string, tok *oauth2.Token) error
}

type RedisStore struct {
	pool    *redis.Pool
}

// addr format is host:port.
// clusterName is blank means addr is the master address, otherwise addr is sentinel address.
func NewRedisStorage(addr, clusterName, password string) *RedisStore {
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
					clog.Errorf("redis master (%s) AUTH error: %s", addr, err)
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
	
	return &RedisStore{pool: p}
}

func (rs *RedisStore) LoadTokenGitlab(user string) *oauth2.Token {
	clog.Debug("loading user:", user)
	return labStore[user]
}

func (rs *RedisStore) SaveTokenGitlab(user string, tok *oauth2.Token) error {
	clog.Debugf("%v: %#v", user, tok)
	labStore[user] = tok
	return nil
}

func (rs *RedisStore) LoadTokenGithub(user string) *oauth2.Token {
	clog.Debug("called.")
	return hubStore[user]
}

func (rs *RedisStore) SaveTokenGithub(user string, tok *oauth2.Token) error {
	clog.Debug("called.")
	hubStore[user] = tok
	return nil
}
