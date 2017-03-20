package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	mathrand "math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/zonesan/clog"
	"golang.org/x/crypto/ssh"
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
	// if data == nil {
	// 	data = genRespJson(nil)
	// }

	if body, err := json.MarshalIndent(data, "", "  "); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	}
}

// rsa public and private keys
func generateKeyPair() (privateKey, publicKey string, err error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return
	}
	err = priv.Validate()
	if err != nil {
		return
	}

	priv_der := x509.MarshalPKCS1PrivateKey(priv)

	// pem.Block
	// blk pem.Block
	priv_blk := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   priv_der,
	}

	// Resultant private key in PEM format.
	// priv_pem string
	privateKey = string(pem.EncodeToMemory(&priv_blk))
	//println("private:", privateKey)

	// ...

	pub := priv.PublicKey

	// pub.pem
	//pub_der, err := x509.MarshalPKIXPublicKey(&pub)
	//if err != nil {
	//	return
	//}
	//
	//pub_blk := pem.Block {
	//	Type: "PUBLIC KEY",
	//	Headers: nil,
	//	Bytes: pub_der,
	//}
	//publicKey = string(pem.EncodeToMemory(&pub_blk))
	//println("public:", publicKey)

	sshpub, err := ssh.NewPublicKey(&pub)
	if err != nil {
		return
	}
	publicKey = string(ssh.MarshalAuthorizedKey(sshpub))
	publicKey = strings.TrimRight(publicKey, "\n")
	publicKey = fmt.Sprintf("%s rsa-key-%s", publicKey, time.Now().Format("20060102"))
	//println("public:", publicKey)

	return
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
