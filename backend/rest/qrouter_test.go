package qrouter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"
	//"log"
	"net/http"
	"net/http/httptest"
	"quickstep/backend/store"
	"testing"

	"github.com/stretchr/testify/assert"
)

/* JsonToken - temp struct keeping json token */
type JSONToken struct {
	Token string `json:"token"`
}

func TestNullSession(t *testing.T) {
	// test database connection error
	router, err := New("localhost:9090", nil)
	assert.Nil(t, err)
	err = router.Enable()
	assert.Nil(t, err)
	server := httptest.NewServer(router.Mux)
	//server.Start()
	defer server.Close()
	loginURL := fmt.Sprintf("%s/login", server.URL)
	res, err := http.Get(loginURL)
	assert.Nil(t, err)
	assert.Equal(t, 404, res.StatusCode)
	var jsonStr = []byte(`{"title":"Bummer"}`)
	req, err := http.NewRequest("POST", loginURL, bytes.NewBuffer(jsonStr))
	assert.Nil(t, err)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	assert.Equal(t, 500, resp.StatusCode)
}

func TestRouterLogin(t *testing.T) {
	var myToken JSONToken
	// test login cases
	super := new(UserAuth)
	db := new(qdb.Qdb)
	db.Type = "mongodb"
	db.Timeout = time.Second * 10
	db.URL = "localhost"
	session, err := db.Open()
	assert.Nil(t, err)
	router, err := New("localhost:9090", session)
	assert.Nil(t, err)
	err = router.Enable()
	assert.Nil(t, err)
	server := httptest.NewServer(router.Mux)
	//server.Start()
	defer server.Close()
	loginURL := fmt.Sprintf("%s/login", server.URL)
	res, err := http.Get(loginURL)
	assert.Nil(t, err)
	assert.Equal(t, 404, res.StatusCode)
	var jsonStr = []byte(`{"title":"Buy cheese and bread for breakfast."}`)
	req, err := http.NewRequest("POST", loginURL, bytes.NewBuffer(jsonStr))
	assert.Nil(t, err)
	//req.Header.Set("X-Custom-Header", "myvalue")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	resp.Body.Close()
	assert.Nil(t, err)
	assert.Equal(t, 403, resp.StatusCode)
	// bad json
	jsonStr = []byte(`this is not jeson string`)
	req, err = http.NewRequest("POST", loginURL, bytes.NewBuffer(jsonStr))
	assert.Nil(t, err)
	//req.Header.Set("X-Custom-Header", "myvalue")
	req.Header.Set("Content-Type", "application/json")

	client = &http.Client{}
	resp, err = client.Do(req)
	resp.Body.Close()
	assert.Nil(t, err)
	assert.Equal(t, 403, resp.StatusCode)

	super.Name = "super"
	jsonStr, err = json.Marshal(super)
	assert.Nil(t, err)
	req, err = http.NewRequest("POST", loginURL, bytes.NewBuffer(jsonStr))
	assert.Nil(t, err)
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)
	assert.Nil(t, err)
	assert.Equal(t, 403, resp.StatusCode)
	resp.Body.Close()

	super.Name = ""
	super.Password = "LamePassword"
	super.Org = "org"
	jsonStr, err = json.Marshal(super)
	assert.Nil(t, err)
	req, _ = http.NewRequest("POST", loginURL, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	resp, _ = client.Do(req)
	assert.Nil(t, err)
	assert.Equal(t, 403, resp.StatusCode)
	resp.Body.Close()

	super.Name = "super"
	super.Password = "password"
	jsonStr, err = json.Marshal(super)
	assert.Nil(t, err)
	req, _ = http.NewRequest("POST", loginURL, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	resp, _ = client.Do(req)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(body, &myToken)
	assert.NotEmpty(t, myToken.Token)
}

func TestRouterLoginPlugin(t *testing.T) {
	var myToken JSONToken
	// test login cases
	super := new(UserAuth)
	db := new(qdb.Qdb)
	client := &http.Client{}

	db.Type = "mongodb"
	db.Timeout = time.Second * 10
	db.URL = "localhost"
	session, err := db.Open()
	assert.Nil(t, err)
	router, err := New("localhost:9090", session)
	assert.Nil(t, err)
	err = router.EnablePlugins("rauth")
	assert.Nil(t, err)
	err = router.Enable()
	assert.Nil(t, err)

	server := httptest.NewServer(router.Mux)
	defer server.Close()
	loginURL := fmt.Sprintf("%s/login", server.URL)

	super.Name = "super"
	super.Password = "password"
	super.Org = "org"
	jsonStr, err := json.Marshal(super)
	assert.Nil(t, err)
	req, _ := http.NewRequest("POST", loginURL, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(body, &myToken)
	resp.Body.Close()
	assert.NotEmpty(t, myToken.Token)
	statURL := fmt.Sprintf("%s/stat", server.URL)
	req, _ = http.NewRequest("GET", statURL, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)
	assert.Nil(t, err)
	body, _ = ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	assert.Equal(t, 403, resp.StatusCode)
	req, _ = http.NewRequest("GET", statURL, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Golden-Ticket", myToken.Token)
	resp, err = client.Do(req)
	assert.Nil(t, err)
	body, _ = ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode) // auth succeed
	super.Name = "super"
	super.Password = "bad_password"
	jsonStr, err = json.Marshal(super)
	assert.Nil(t, err)
	req, _ = http.NewRequest("POST", loginURL, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)
	assert.Nil(t, err)
	assert.Equal(t, 403, resp.StatusCode)
	router.s.SigningKey = []byte("BAD_SIGN_KEY")
	req, _ = http.NewRequest("POST", loginURL, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)
	assert.Nil(t, err)
	assert.Equal(t, 403, resp.StatusCode)
}

func authAndGetToken(user string, passwd string) (*httptest.Server, *JSONToken, error) {
	super := new(UserAuth)
	db := new(qdb.Qdb)
	token := new(JSONToken)
	client := &http.Client{}
	db.Type = "mongodb"
	db.Timeout = time.Second * 10
	db.URL = "localhost"
	session, err := db.Open()
	if err != nil {
		return nil, nil, err
	}
	router, err := New("localhost:9090", session)
	if err != nil {
		return nil, nil, err
	}
	err = router.EnablePlugins("rauth")
	if err != nil {
		return nil, nil, err
	}
	err = router.Enable()
	if err != nil {
		return nil, nil, err
	}
	server := httptest.NewServer(router.Mux)
	loginURL := fmt.Sprintf("%s/login", server.URL)
	super.Name = user
	super.Password = passwd
	super.Org = "org"
	jsonStr, err := json.Marshal(super)
	if err != nil {
		return nil, nil, err
	}
	req, _ := http.NewRequest("POST", loginURL, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	body, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(body, token)
	resp.Body.Close()
	return server, token, nil
}

func GetToken(server *httptest.Server, user string, org string, passwd string) (*JSONToken, error) {
	token := new(JSONToken)
	loginURL := fmt.Sprintf("%s/login", server.URL)
	super := new(qdb.User)
	super.Name = user
	super.Password = passwd
	super.Org = org
	jsonStr, err := json.Marshal(super)
	if err != nil {
		return nil, err
	}
	client := &http.Client{}
	req, _ := http.NewRequest("POST", loginURL, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	body, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(body, token)
	resp.Body.Close()
	return token, nil
}

func TestCreateGetUser(t *testing.T) {
	server, superToken, err := authAndGetToken("super", "password")
	assert.Nil(t, err)
	client := &http.Client{}
	defer server.Close()
	//jsonStr := "blah" //json.Marshal(super)
	jsonStr := []byte("")
	userURL := fmt.Sprintf("%s/user/super", server.URL)
	req, _ := http.NewRequest("GET", userURL, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Golden-Ticket", superToken.Token)
	resp, err := client.Do(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode) // auth succeed

	bacl := qdb.CreateACL("blah.org", "crud")
	blahOrgUser := new(qdb.User)
	blahOrgUser.Name = "admin"
	blahOrgUser.Password = "password"
	blahOrgUser.Org = "blah.org"
	blahOrgUser.ACL = append(blahOrgUser.ACL, *bacl) // can create modify list users in blah domain
	blahJSON, err := json.Marshal(blahOrgUser)
	assert.Nil(t, err)

	fooOrgUser := new(qdb.User)
	fooOrgUser.Name = "admin"
	fooOrgUser.Password = "password"
	fooOrgUser.Org = "foo.org"
	facl := qdb.CreateACL("foo.org", "crud")
	fooOrgUser.ACL = append(fooOrgUser.ACL, *facl) // can create modify read users in blah.org and foo.org domain
	facl = qdb.CreateACL("blah.org", "crud")
	fooOrgUser.ACL = append(fooOrgUser.ACL, *facl) // can create modify read users in blah.org and foo.org domain
	fooJSON, err := json.Marshal(fooOrgUser)
	assert.Nil(t, err)

	// create 2 users
	userURL = fmt.Sprintf("%s/user/admin", server.URL)
	req, _ = http.NewRequest("POST", userURL, bytes.NewBuffer(blahJSON))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Golden-Ticket", superToken.Token)
	resp, err = client.Do(req)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Printf("blah %s\n", string(body))
	assert.Equal(t, http.StatusOK, resp.StatusCode) // auth succeed

	req, _ = http.NewRequest("POST", userURL, bytes.NewBuffer(fooJSON))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Golden-Ticket", superToken.Token)
	resp, err = client.Do(req)
	body, _ = ioutil.ReadAll(resp.Body)
	fmt.Printf("foo %s\n", string(body))
	assert.Equal(t, http.StatusOK, resp.StatusCode) // auth succeed
	// get token for both
	blahToken, err := GetToken(server, blahOrgUser.Name, blahOrgUser.Org, blahOrgUser.Password)
	assert.Nil(t, err)
	fooToken, err := GetToken(server, fooOrgUser.Name, fooOrgUser.Org, fooOrgUser.Password)
	assert.Nil(t, err)
	fmt.Println(blahToken.Token)
	fmt.Println(fooToken.Token)
	// do get for BLAH on blah.org an foo.org domains
	qUser := new(qdb.User)
	qUser.Name = blahOrgUser.Name
	qUser.Org = blahOrgUser.Org
	fooJSON, err = json.Marshal(qUser)
	assert.Nil(t, err)
	//fooJson = fmt.Sprintf("{\"name\"}")
	//	req, _ = http.NewRequest("GET", userURL, bytes.NewBuffer(fooJSON))
	req, _ = http.NewRequest("GET", userURL, bytes.NewBufferString(""))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Golden-Ticket", blahToken.Token)
	resp, err = client.Do(req)
	body, _ = ioutil.ReadAll(resp.Body)
	verBlahUser := new(qdb.User)
	assert.Nil(t, json.Unmarshal(body, verBlahUser))
	assert.Equal(t, blahOrgUser.Name, verBlahUser.Name)
	assert.Equal(t, blahOrgUser.Org, verBlahUser.Org)
	resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode) // auth succeed

	req, _ = http.NewRequest("GET", userURL, bytes.NewBuffer(fooJSON))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Golden-Ticket", blahToken.Token)
	resp, err = client.Do(req)
	body, _ = ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))
	resp.Body.Close()

	//try to list for both in both domains one should faile
	// create user for both in both doamins - one should fail
}
