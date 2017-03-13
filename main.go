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
	store = NewRedisStorage()
}
