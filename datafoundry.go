package main

import (
	"bytes"
	"crypto/tls"
	// "encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"sync/atomic"

	"github.com/zonesan/clog"
	"k8s.io/kubernetes/pkg/api/unversioned"
	kapi "k8s.io/kubernetes/pkg/api/v1"
)

type DataFoundryClient struct {
	host        string
	oapiURL     string
	kapiURL     string
	bearerToken atomic.Value
}

type User struct {
	unversioned.TypeMeta `json:",inline"`
	// Standard object's metadata.
	kapi.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// FullName is the full name of user
	FullName string `json:"fullName,omitempty" protobuf:"bytes,2,opt,name=fullName"`

	// Identities are the identities associated with this user
	Identities []string `json:"identities" protobuf:"bytes,3,rep,name=identities"`

	// Groups are the groups that this user is a member of
	Groups []string `json:"groups" protobuf:"bytes,4,rep,name=groups"`
}

func NewDataFoundryTokenClient(bearerToken string) *DataFoundryClient {
	// host = setBaseUrl(host)
	client := &DataFoundryClient{
		host:    dataFoundryHostAddr,
		oapiURL: dataFoundryHostAddr + "/oapi/v1",
		kapiURL: dataFoundryHostAddr + "/api/v1",
	}

	client.setBearerToken(bearerToken)

	return client
}

func (c *DataFoundryClient) setBearerToken(token string) {
	c.bearerToken.Store(token)
}

func (c *DataFoundryClient) BearerToken() string {
	//return oc.bearerToken
	return c.bearerToken.Load().(string)
}

func (c *DataFoundryClient) CreateSecret(ns, name string, data map[string]string) (*kapi.Secret, error) {
	uri := "/namespaces/" + ns + "/secrets"

	sReq := new(kapi.Secret)
	sReq.Name = name
	sReq.Data = make(map[string][]byte)

	// clog.Debugf("%v", data)

	for key, value := range data {
		sReq.Data[key] = []byte(value)
	}
	secret := new(kapi.Secret)
	err := c.KPost(uri, sReq, secret)
	return secret, err
}

func (c *DataFoundryClient) OGet(uri string, into interface{}) error {
	return c.doRequest("GET", c.oapiURL+uri, nil, into)
}

func (c *DataFoundryClient) OPost(uri string, body, into interface{}) error {
	return c.doRequest("POST", c.oapiURL+uri, body, into)
}

func (c *DataFoundryClient) KGet(uri string, into interface{}) error {
	return c.doRequest("GET", c.kapiURL+uri, nil, into)
}

func (c *DataFoundryClient) KPost(uri string, body, into interface{}) error {
	return c.doRequest("POST", c.kapiURL+uri, body, into)
}

func (c *DataFoundryClient) doRequest(method, url string, bodyParams interface{}, v interface{}) (err error) {
	var reqbody []byte
	if bodyParams != nil {
		reqbody, err = json.Marshal(bodyParams)
		if err != nil {
			return err
		}
	}

	resp, err := c.request(method, url, reqbody)

	if err != nil {
		return err
	}

	defer func() {
		// Drain up to 512 bytes and close the body to let the Transport reuse the connection
		io.CopyN(ioutil.Discard, resp.Body, 512)
		resp.Body.Close()
	}()

	err = checkRespStatus(resp)
	if err != nil {
		// even though there was an error, we still return the response
		// in case the caller wants to inspect it further
		return err
	}

	if v != nil {
		if w, ok := v.(io.Writer); ok {
			io.Copy(w, resp.Body)
		} else {
			err = json.NewDecoder(resp.Body).Decode(v)
			if err == io.EOF {
				err = nil // ignore EOF errors caused by empty response body
			}
			//clog.Tracef("%#v", v)
		}
	}
	return err

}

func (c *DataFoundryClient) request(method string, url string, body []byte) (*http.Response, error) {
	token := c.BearerToken()
	if token == "" {
		return nil, errors.New("token is blank")
	}

	clog.Trace("request url:", url)
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

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)

	transCfg := &http.Transport{
		DisableKeepAlives: true,
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: transCfg,
		// Timeout:   timeout,
	}
	return client.Do(req)
}

// StatusError is an error intended for consumption by a REST API server; it can also be
// reconstructed by clients from a REST response. Public to allow easy type switches.
type StatusError struct {
	ErrStatus unversioned.Status
}

// Error implements the Error interface.
func (e *StatusError) Error() string {
	return e.ErrStatus.Message
}

// CheckResponse checks the API response for errors, and returns them if
// present.  A response is considered an error if it has a status code outside
// the 200 range.  API error responses are expected to have either no response
// body, or a JSON response body that maps to ErrorResponse.  Any other
// response body will be silently ignored.
//
// The error type will be *RateLimitError for rate limit exceeded errors,
// and *TwoFactorAuthError for two-factor authentication errors.
func checkRespStatus(r *http.Response) error {
	if c := r.StatusCode; 200 <= c && c <= 299 {
		return nil
	}
	// openshift returns 401 with a plain text but not ErrStatus json, so we hacked this response text.
	if r.StatusCode == http.StatusUnauthorized {
		errorResponse := &StatusError{}
		errorResponse.ErrStatus.Code = http.StatusUnauthorized
		errorResponse.ErrStatus.Message = http.StatusText(http.StatusUnauthorized)
		return errorResponse
	}

	errorResponse := &StatusError{}
	data, err := ioutil.ReadAll(r.Body)
	if err == nil && data != nil {
		clog.Errorf("%v,%s", r.StatusCode, data)
		json.Unmarshal(data, &errorResponse.ErrStatus)
	}

	return errorResponse
}
