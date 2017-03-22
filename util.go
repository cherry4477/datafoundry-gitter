package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	mathrand "math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/zonesan/clog"
)

func setBaseUrl(urlStr string) string {
	// Make sure the given URL end with a slash
	if strings.HasSuffix(urlStr, "/") {
		return setBaseUrl(strings.TrimSuffix(urlStr, "/"))
	}
	return urlStr
}

func randToken() string {
	b := make([]byte, 40)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

// func redirectUrl(oauthConf *oauth2.Config) string {
// 	return ""
// }

func debug(v interface{}) {
	// return
	d, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Printf("json.MarshlIndent() failed with %s\n", err)
	}
	fmt.Println(string(d))
}

func RespOK(w http.ResponseWriter, data interface{}) {
	if data == nil {
		data = genRespJson(nil)
	}

	if body, err := json.MarshalIndent(data, "", "  "); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	}
}

func RespError(w http.ResponseWriter, err error) {
	resp := genRespJson(err)

	if body, err := json.MarshalIndent(resp, "", "  "); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.status)
		w.Write(body)
	}

}

func genRespJson(err error) *APIResponse {
	resp := new(APIResponse)

	if err == nil {
		resp.Code = ErrCodeOK
		resp.status = http.StatusOK
	} else {
		if e, ok := err.(*ErrorMessage); ok {
			resp.Code = e.Code
			resp.status = trickCode2Status(resp.Code) //http.StatusBadRequest
			resp.Message = e.Message
		} else if e, ok := err.(*StatusError); ok {
			resp.Code = int(e.ErrStatus.Code)

			// frontend can't handle 403, he will panic...
			{
				if resp.Code == http.StatusForbidden {
					resp.Code = http.StatusBadRequest
				}
			}
			resp.status = resp.Code
			resp.Message = e.ErrStatus.Message

		} else {
			resp.Code = ErrCodeBadRequest
			resp.Message = err.Error()
			resp.status = trickCode2Status(resp.Code) //http.StatusBadRequest
		}
	}

	resp.Reason = http.StatusText(resp.status)

	return resp
}

func trickCode2Status(errCode int) int {
	var statusCode int
	if errCode < 10000 {
		statusCode = errCode % 1000
	} else {
		statusCode = trickCode2Status(errCode / 10)
	}

	return statusCode
}

func parseRequestBody(r *http.Request, v interface{}) error {
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		return err
	}
	clog.Debug("Request Body:", string(b))
	if err := json.Unmarshal(b, v); err != nil {
		return err
	}

	return nil
}

var letters = []rune("abcdefghijklmnopqrstuvwxyz0123456789")
var numLetters = len(letters)
var rng = struct {
	sync.Mutex
	rand *mathrand.Rand
}{
	rand: mathrand.New(mathrand.NewSource(time.Now().UTC().UnixNano())),
}

// intn generates an integer in range 0->max.
// By design this should panic if input is invalid, <= 0.
func intn(max int) int {
	rng.Lock()
	defer rng.Unlock()
	return rng.rand.Intn(max)
}

// String generates a random alphanumeric string n characters long.  This will
// panic if n is less than zero.
func randomStr(length int) string {
	b := make([]rune, length)
	for i := range b {
		b[i] = letters[intn(numLetters)]
	}
	return string(b)
}

func httpAddr(addr string) string {

	if !strings.HasPrefix(strings.ToLower(addr), "http://") &&
		!strings.HasPrefix(strings.ToLower(addr), "https://") {
		return fmt.Sprintf("http://%s", addr)
	}

	return setBaseUrl(addr)
}

func httpsAddr(addr string) string {

	if !strings.HasPrefix(strings.ToLower(addr), "http://") &&
		!strings.HasPrefix(strings.ToLower(addr), "https://") {
		return fmt.Sprintf("https://%s", addr)
	}

	return setBaseUrl(addr)
}
