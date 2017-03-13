package main

import (
	"net/http"

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

	// list repo branches handler
	router.GET("/repos/github/branches", handleGitHubRepoBranches)
	router.GET("/repos/gitlab/branches", handleGitLabRepoBranches)

	// webhhook handler
	router.GET("/repos/github/webhook", handleGitHubCheckWebhook)
	router.POST("/repos/github/webhook", handleGitHubCreateWebhook)
	router.DELETE("/repos/github/webhook/:hookid", handleGitHubRemoveWebhook)
	router.GET("/repos/gitlab/webhook", handleGitLabCheckWebhook)
	router.POST("/repos/gitlab/webhook", handleGitLabCreateWebhook)
	router.DELETE("/repos/gitlab/webhook/:hookid", handleGitLabRemoveWebhook)

	// router.GET("/repos/:source", handleGitLabRepos)
	// router.GET("/repos/:source/branch", handleGitLabRepos)
	// router.GET("/repos/:source/webhook", handleGitLabRepos)

	clog.Debug("listening on port 7000 ...")
	clog.Fatal(http.ListenAndServe(":7000", router))

}

func init() {
	store = NewRedisStorage()
}
