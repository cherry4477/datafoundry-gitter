package main

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/zonesan/clog"
)

var (
	store               Storage
	DataFoundryHostAddr string
)

func main() {

	router := httprouter.New()
	router.GET("/", handleMain)
	router.NotFound = &Mux{}

	// authoriza handler
	router.GET("/authorize/:source", authorize(handleGitterAuthorize))

	// callback handler
	router.GET("/github_oauth_cb", handleGitHubCallback)
	router.GET("/gitlab_oauth_cb", handleGitLabCallback)

	router.GET("/repos/:source", authorize(handleRepos))
	router.GET("/repos/:source/branches", authorize(handleRepoBranches))

	router.GET("/repos/:source/secret", authorize(handleSecret))

	router.GET("/repos/:source/webhook", authorize(handleCheckWebhook))
	router.POST("/repos/:source/webhook", authorize(handleCreateWebhook))
	router.DELETE("/repos/:source/webhook/:hookid", authorize(handleRemoveWebhook))

	clog.Debug("listening on port 7000 ...")
	clog.Fatal(http.ListenAndServe(":7000", router))

}

func init() {

	DataFoundryHostAddr = os.Getenv("DATAFOUNDRY_API_SERVER")
	if len(DataFoundryHostAddr) == 0 {
		clog.Fatal("DATAFOUNDRY_API_SERVER must be specified.")
	}
	DataFoundryHostAddr = httpsAddr(DataFoundryHostAddr)
	clog.Debug("datafoundry api server:", DataFoundryHostAddr)

	// redis

	var redisStorager KeyValueStorager

	func() {
		var redisParams = os.Getenv("REDIS_SERVER_PARAMS")
		if redisParams != "" {
			// host+port+password
			var words = strings.Split(redisParams, "+")
			if len(words) < 3 {
				clog.Warnf("REDIS_SERVER_PARAMS (%s) should have 3 params, now: %d", redisParams, len(words))
				return
			}

			redisStorager = NewRedisKeyValueStorager(
				words[0]+":"+words[1],
				"", // blank clusterName means no sentinel servers
				strings.Join(words[2:], "+"),// password

			)

			clog.Info("redis storage created with REDIS_SERVER_PARAMS:", redisParams)
		} else {
			var vcapServices = os.Getenv("VCAP_SERVICES")
			if vcapServices == "" {
				clog.Warn("neither REDIS_SERVER_PARAMS nor VCAP_SERVICES env is not set")
				return
			}
			var redisBsiName = os.Getenv("Redis_BackingService_Name")
			if redisBsiName == "" {
				clog.Warn("Redis_BackingService_Name env is not set")
				return
			}

			type Credential struct {
				Host     string `json:"Host"`
				Name     string `json:"Name"`
				Password string `json:"Password"`
				Port     string `json:"Port"`
				Uri      string `json:"Uri"`
				Username string `json:"Username"`
				VHost    string `json:"Vhost"`
			}
			type Service struct {
				Name       string     `json:"name"`
				Label      string     `json:"label"`
				Plan       string     `json:"plan"`
				Credential Credential `json:"credentials"`
			}

			var services = map[string][]Service{}
			if err := json.Unmarshal([]byte(vcapServices), &services); err != nil {
				clog.Warnf("unmarshal VCAP_SERVICES error: %f\n%s", err, vcapServices)
				return
			}

			var redisServices = services["Redis"]
			if len(redisServices) == 0 {
				clog.Warn("no redis services found in VCAP_SERVICES")
				return
			}

			var credential = &redisServices[0].Credential
			redisStorager = NewRedisKeyValueStorager(
				credential.Host+":"+credential.Port,
				credential.Name,
				credential.Password,
			)

			clog.Info("redis storage created with VCAP_SERVICES:", credential)
		}
	}()

	if redisStorager != nil {
		store = NewStorage(redisStorager)
	} else {
		store = NewStorage(NewMemoryKeyValueStorager())

		clog.Warn("redis storage is not reachable, use memory storage instead.")
		clog.Warn("REDIS STORAGE IS NOT REACHABLE, USE MEMORY STORAGE INSTEAD.")
	}
}
