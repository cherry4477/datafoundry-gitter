package main

import (
	"net/http"
	"os"
	"strings"
	"encoding/json"

	"github.com/julienschmidt/httprouter"
	"github.com/zonesan/clog"
)

var (
	store Storage
)

// http://localhost:18080/?code=958b1416f6362d24229ea051debeaa5256db9539ff12655f55aa0afc989429af&state=your_unique_state_hash
// https://gitlab.example.com/oauth/authorize?client_id=APP_ID&redirect_uri=REDIRECT_URI&response_type=code&state=your_unique_state_hash
func main() {

	router := httprouter.New()
	router.GET("/", handleMain)

	// authoriza handler
	router.GET("/authorize/:source", handleGitterAuthorize)

	// callback handler
	router.GET("/github_oauth_cb", handleGitHubCallback)
	router.GET("/gitlab_oauth_cb", handleGitLabCallback)

	// list repos handler
	router.GET("/repos/github", handleGitHubRepos)
	router.GET("/repos/gitlab", handleGitLabRepos)

	//list repo branches handler
	router.GET("/repos/github/:namespace/:repo/branches", handleGitHubRepoBranches)
	router.GET("/repos/gitlab/projects/:repoid/branches", handleGitLabRepoBranches)

	// router.GET("/repos/:source/:repo/webhook", handleCheckWebhook)
	// router.POST("/repos/:source/:repo/webhook", handleCreateWebhook)
	// router.DELETE("/repos/:source/:repo/webhook", handleRemoveWebhook)

	clog.Debug("listening on port 7000 ...")
	clog.Fatal(http.ListenAndServe(":7000", router))

}

func init() {

	store = NewStorage(NewMemoryKeyValueStorager())
	return

	// redis
	var redisStorager KeyValueStorager

	var redisParams = os.Getenv("REDIS_SERVER_PARAMS")
	if redisParams != "" {
		// host+port+password
		var words = strings.Split(redisParams, "+")
		if len(words) < 3 {
			clog.Fatalf("REDIS_SERVER_PARAMS (%s) should have 3 params, now: %d", redisParams, len(words))
		}

		redisStorager = NewRedisKeyValueStorager(
			words[0] + ":" +  words[1],
			"", // blank clusterName means no sentinel servers
			strings.Join(words[2:], "+", // password
			),
		)

		clog.Info("redis storage created with REDIS_SERVER_PARAMS:", redisParams)
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
			clog.Fatalf("unmarshal VCAP_SERVICES error: %f\n%s", err, vcapServices)
		}

		var redisServices = services[RedisServiceKindName]
		if len(redisServices) == 0 {
			clog.Fatal("no redis services found in VCAP_SERVICES")
		}

		var credential = &redisServices[0].Credential
		redisStorager = NewRedisKeyValueStorager(
			credential.Host + ":" + credential.Port,
			credential.Name,
			credential.Password,
		)

		clog.Info("redis storage created with VCAP_SERVICES:", credential)
	}

	store = NewStorage(redisStorager)
}
