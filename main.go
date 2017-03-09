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

	router.GET("/authorize/:source", handleGitterAuthorize)

	// router.GET("/login", handleGitHubLogin)

	// github callback handler
	router.GET("/github_oauth_cb", handleGitHubCallback)

	//gitlab callback handler
	router.GET("/gitlab_oauth_cb", handleGitLabCallback)

	clog.Debug("listening on port 7000 ...")
	clog.Fatal(http.ListenAndServe(":7000", router))

}

func init() {
	store = NewRedisStorage()
}
