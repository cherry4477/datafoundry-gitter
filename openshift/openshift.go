package openshift

import (
	"errors"
	"bufio"
	"bytes"
	"strings"
	"time"
	"sync/atomic"
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"net/http"
	neturl "net/url"

	kclient "k8s.io/kubernetes/pkg/client/unversioned"
	kapi "k8s.io/kubernetes/pkg/api/v1"
	"github.com/openshift/origin/pkg/cmd/util/tokencmd"
)

//==============================================================
//
//==============================================================

func validateHttpsAddr(addr string) string {
	addr = strings.TrimSpace(addr)

	if strings.HasSuffix(addr, "/") {
		addr = strings.TrimRight(addr, "/")
	}

	if strings.HasPrefix(addr, "http://") {
		panic("sorry, http protocol is not supported")
	}

	if strings.HasPrefix(addr, "https://") {
		return addr
	}

	return "https://" + addr
}

//==============================================================
// 
//==============================================================

type OpenshiftClient struct {
	name string // for logging only

	host string
	oapiUrl string
	kapiUrl string

	username    string
	password    string
	//bearerToken string
	bearerToken atomic.Value
}

func NewClient(addr, token string) *OpenshiftClient {
	addr = validateHttpsAddr(addr)
	oc := &OpenshiftClient{
		host: addr,
		oapiUrl: addr + "/oapi/v1",
		kapiUrl: addr + "/api/v1",
	}

	oc.setBearerToken(token)

	return oc
}

// Derive another client.
// The input token must contains "Bearer "
func (baseOC *OpenshiftClient) Derive(token string) *OpenshiftClient {
	oc := &OpenshiftClient{
		host:    baseOC.host,
		oapiUrl: baseOC.oapiUrl,
		kapiUrl: baseOC.kapiUrl,
	}

	oc.setBearerToken(token)

	return oc
}

func (oc *OpenshiftClient) BearerToken() string {
	return oc.bearerToken.Load().(string)
}

func (oc *OpenshiftClient) setBearerToken(token string) {
	oc.bearerToken.Store(token)
}

func (oc *OpenshiftClient) updateBearerToken(durPhase time.Duration) {
	for {
		clientConfig := &kclient.Config{}
		clientConfig.Host = oc.host
		clientConfig.Insecure = true
		//clientConfig.Version =

		println("Request Token from: ", clientConfig.Host)

		token, err := tokencmd.RequestToken(clientConfig, nil, oc.username, oc.password)
		if err != nil {
			println("RequestToken error: ", err.Error())

			time.Sleep(15 * time.Second)
		} else {
			//clientConfig.BearerToken = token
			//oc.bearerToken = "Bearer " + token
			oc.setBearerToken("Bearer " + token)

			println(oc.name, ", RequestToken token: ", token)

			// durPhase is to avoid mulitple OCs updating tokens at the same time
			time.Sleep(3 * time.Hour + durPhase)
			durPhase = 0
		}
	}
}

func (oc *OpenshiftClient) request(method string, url string, body []byte, timeout time.Duration) (*http.Response, error) {
	//token := oc.bearerToken
	token := oc.BearerToken()
	if token == "" {
		return nil, errors.New("token is blank")
	}

	var req *http.Request
	var err error
	if len(body) == 0 {
		req, err = http.NewRequest(method, url, nil)
	} else {
		req, err = http.NewRequest(method, url, bytes.NewReader(body))
	}

	if err != nil {
		return nil, err
	}

	//for k, v := range headers {
	//	req.Header.Add(k, v)
	//}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)

	transCfg := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: transCfg,
		Timeout:   timeout,
	}
	return client.Do(req)
}

type WatchStatus struct {
	Info []byte
	Err  error
}

func (oc *OpenshiftClient) doWatch(url string) (<-chan WatchStatus, chan<- struct{}, error) {
	res, err := oc.request("GET", url, nil, 0)
	if err != nil {
		return nil, nil, err
	}
	//if res.Body == nil {
	//	return nil, nil, errors.New("response.body is nil")
	//}

	statuses := make(chan WatchStatus, 5)
	canceled := make(chan struct{}, 1)

	go func() {
		defer func() {
			close(statuses)
			res.Body.Close()
		}()

		reader := bufio.NewReader(res.Body)
		for {
			select {
			case <-canceled:
				return
			default:
			}

			line, err := reader.ReadBytes('\n')
			if err != nil {
				statuses <- WatchStatus{nil, err}
				return
			}

			statuses <- WatchStatus{line, nil}
		}
	}()

	return statuses, canceled, nil
}

func (oc *OpenshiftClient) OWatch(uri string) (<-chan WatchStatus, chan<- struct{}, error) {
	return oc.doWatch(oc.oapiUrl + "/watch" + uri)
}

func (oc *OpenshiftClient) KWatch(uri string) (<-chan WatchStatus, chan<- struct{}, error) {
	return oc.doWatch(oc.kapiUrl + "/watch" + uri)
}

const GeneralRequestTimeout = time.Duration(30) * time.Second

/*
func (oc *OpenshiftClient) doRequest (method, url string, body []byte) ([]byte, error) {
	res, err := oc.request(method, url, body, GeneralRequestTimeout)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	return ioutil.ReadAll(res.Body)
}

func (oc *OpenshiftClient) ORequest (method, uri string, body []byte) ([]byte, error) {
	return oc.doRequest(method, oc.oapiUrl + uri, body)
}

func (oc *OpenshiftClient) KRequest (method, uri string, body []byte) ([]byte, error) {
	return oc.doRequest(method, oc.kapiUrl + uri, body)
}
*/

type OpenshiftREST struct {
	oc  *OpenshiftClient
	Err error
	StatusCode int
}

func NewOpenshiftREST(client *OpenshiftClient) *OpenshiftREST {
	if client == nil {
		panic("client argument can't be nil")
	}
	return &OpenshiftREST{oc: client}
}

func (osr *OpenshiftREST) doRequest(method, url string, bodyParams interface{}, into interface{}) *OpenshiftREST {
	if osr.Err != nil {
		return osr
	}

	var body []byte
	if bodyParams != nil {
		body, osr.Err = json.Marshal(bodyParams)
		if osr.Err != nil {
			return osr
		}
	}

	//println("11111 method = ", method, ", url = ", url)

	//res, osr.Err := oc.request(method, url, body, GeneralRequestTimeout) // non-name error
	res, err := osr.oc.request(method, url, body, GeneralRequestTimeout)
	osr.Err = err
	if osr.Err != nil {
		return osr
	}
	osr.StatusCode = res.StatusCode
	defer res.Body.Close()

	var data []byte
	data, osr.Err = ioutil.ReadAll(res.Body)
	if osr.Err != nil {
		return osr
	}

	//println("22222 len(data) = ", len(data), " , res.StatusCode = ", res.StatusCode)

	if res.StatusCode < 200 || res.StatusCode >= 400 {
		osr.Err = errors.New(string(data))
	} else {
		if into != nil {
			//println("into data = ", string(data), "\n")

			osr.Err = json.Unmarshal(data, into)
		}
	}

	return osr
}

func buildUriWithSelector(uri string, selector map[string]string) string {
	var buf bytes.Buffer
	for k, v := range selector {
		if buf.Len() > 0 {
			buf.WriteByte(',')
		}
		buf.WriteString(k)
		buf.WriteByte('=')
		buf.WriteString(v)
	}

	if buf.Len() == 0 {
		return uri
	}

	values := neturl.Values{}
	values.Set("labelSelector", buf.String())

	if strings.IndexByte(uri, '?') < 0 {
		uri = uri + "?"
	}

	println("\n uri=", uri+values.Encode(), "\n")

	return uri + values.Encode()
}

// o

func (osr *OpenshiftREST) OList(uri string, selector map[string]string, into interface{}) *OpenshiftREST {

	return osr.doRequest("GET", osr.oc.oapiUrl+buildUriWithSelector(uri, selector), nil, into)
}

func (osr *OpenshiftREST) OGet(uri string, into interface{}) *OpenshiftREST {
	return osr.doRequest("GET", osr.oc.oapiUrl+uri, nil, into)
}

func (osr *OpenshiftREST) ODelete(uri string, into interface{}) *OpenshiftREST {
	return osr.doRequest("DELETE", osr.oc.oapiUrl+uri, &kapi.DeleteOptions{}, into)
}

func (osr *OpenshiftREST) OPost(uri string, body interface{}, into interface{}) *OpenshiftREST {
	return osr.doRequest("POST", osr.oc.oapiUrl+uri, body, into)
}

func (osr *OpenshiftREST) OPut(uri string, body interface{}, into interface{}) *OpenshiftREST {
	return osr.doRequest("PUT", osr.oc.oapiUrl+uri, body, into)
}

// k

func (osr *OpenshiftREST) KList(uri string, selector map[string]string, into interface{}) *OpenshiftREST {
	return osr.doRequest("GET", osr.oc.kapiUrl+buildUriWithSelector(uri, selector), nil, into)
}

func (osr *OpenshiftREST) KGet(uri string, into interface{}) *OpenshiftREST {
	return osr.doRequest("GET", osr.oc.kapiUrl+uri, nil, into)
}

func (osr *OpenshiftREST) KDelete(uri string, into interface{}) *OpenshiftREST {
	return osr.doRequest("DELETE", osr.oc.kapiUrl+uri, &kapi.DeleteOptions{}, into)
}

func (osr *OpenshiftREST) KPost(uri string, body interface{}, into interface{}) *OpenshiftREST {
	return osr.doRequest("POST", osr.oc.kapiUrl+uri, body, into)
}

func (osr *OpenshiftREST) KPut(uri string, body interface{}, into interface{}) *OpenshiftREST {
	return osr.doRequest("PUT", osr.oc.kapiUrl+uri, body, into)
}

