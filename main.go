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
	var redisParams = os.Getenv("REDIS_SERVER_PARAMS")
	if redisParams != "" {
		// host+port+password
		var words = strings.Split(redisParams, "+")
		if len(words) < 3 {
			clog.Fatalf("REDIS_SERVER_PARAMS (%s) should have 3 params, now: %d", redisParams, len(words))
		}

		store = NewRedisStorage(
			words[0]+":"+words[1],
			"", // blank clusterName means no sentinel servers
			strings.Join(words[2:], "+"), // password

		)
	} else {
		const RedisServiceKindName = "Redis"
		var vcapServices = os.Getenv("VCAP_SERVICES")
		if vcapServices == "" {
			clog.Fatal("VCAP_SERVICES env is not set")
		}
		var redisBsiName = os.Getenv("Redis_BackingService_Name")
		if redisBsiName == "" {
			clog.Fatal("Redis_BackingService_Name env is not set")
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
			clog.Fatalf("unmarshal VCAP_SERVICES error: %v\n%s", err, vcapServices)
		}

		var redisServices = services[RedisServiceKindName]
		if len(redisServices) == 0 {
			clog.Fatal("no redis services found in VCAP_SERVICES")
		}

		var credential = &redisServices[0].Credential
		store = NewRedisStorage(
			credential.Host+":"+credential.Port,
			credential.Name,
			credential.Password,
		)
	}

}
