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
	err = router.EnableRest()
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

	bacl := qdb.CreateACL("", "blah.org", "crud")
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
	facl := qdb.CreateACL("", "foo.org", "crud")
	fooOrgUser.ACL = append(fooOrgUser.ACL, *facl) // can create modify read users in blah.org and foo.org domain
	facl = qdb.CreateACL("", "blah.org", "crud")
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
	//fmt.Printf("blah %s\n", string(body))
	assert.Equal(t, http.StatusOK, resp.StatusCode) // auth succeed

	req, _ = http.NewRequest("POST", userURL, bytes.NewBuffer(fooJSON))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Golden-Ticket", superToken.Token)
	resp, err = client.Do(req)
	assert.Equal(t, http.StatusOK, resp.StatusCode) // auth succeed
	// get token for both
	blahToken, err := GetToken(server, blahOrgUser.Name, blahOrgUser.Org, blahOrgUser.Password)
	assert.Nil(t, err)
	fooToken, err := GetToken(server, fooOrgUser.Name, fooOrgUser.Org, fooOrgUser.Password)
	assert.Nil(t, err)
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
	verUser := new(qdb.User)
	assert.Equal(t, http.StatusOK, resp.StatusCode) // auth succeed
	assert.Nil(t, json.Unmarshal(body, verUser))
	assert.Equal(t, blahOrgUser.Name, verUser.Name)
	assert.Equal(t, blahOrgUser.Org, verUser.Org)
	resp.Body.Close()

	req, _ = http.NewRequest("GET", userURL, bytes.NewBuffer(fooJSON))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Golden-Ticket", blahToken.Token)
	resp, err = client.Do(req)
	body, _ = ioutil.ReadAll(resp.Body)
	assert.Equal(t, http.StatusOK, resp.StatusCode) // auth succeed
	assert.Nil(t, json.Unmarshal(body, verUser))
	assert.Equal(t, blahOrgUser.Name, verUser.Name)
	assert.Equal(t, blahOrgUser.Org, verUser.Org)
	resp.Body.Close()
	//user blah  try get foo domain - this should fail
	qUser.Name = fooOrgUser.Name
	qUser.Org = fooOrgUser.Org
	fooJSON, err = json.Marshal(qUser)
	assert.Nil(t, err)
	req, _ = http.NewRequest("GET", userURL, bytes.NewBuffer(fooJSON))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Golden-Ticket", blahToken.Token)
	resp, err = client.Do(req)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode) //failed - OK have no access to that domain
	// do the same with foo user

	qUser.Name = fooOrgUser.Name
	qUser.Org = fooOrgUser.Org
	fooJSON, err = json.Marshal(qUser)

	req, _ = http.NewRequest("GET", userURL, bytes.NewBuffer(fooJSON))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Golden-Ticket", fooToken.Token) // auth as foo user
	resp, err = client.Do(req)
	body, _ = ioutil.ReadAll(resp.Body)
	assert.Equal(t, http.StatusOK, resp.StatusCode) // auth succeed
	assert.Nil(t, json.Unmarshal(body, verUser))
	assert.Equal(t, fooOrgUser.Name, verUser.Name)
	assert.Equal(t, fooOrgUser.Org, verUser.Org)
	resp.Body.Close()
	//user foo try get blah domain - this should succeed
	qUser.Name = blahOrgUser.Name
	qUser.Org = blahOrgUser.Org
	fooJSON, err = json.Marshal(qUser)
	assert.Nil(t, err)
	req, _ = http.NewRequest("GET", userURL, bytes.NewBuffer(fooJSON))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Golden-Ticket", fooToken.Token)
	resp, err = client.Do(req)
	body, _ = ioutil.ReadAll(resp.Body)
	assert.Equal(t, http.StatusOK, resp.StatusCode) //failed - OK have no access to that domain
	assert.Nil(t, json.Unmarshal(body, verUser))
	assert.Equal(t, blahOrgUser.Name, verUser.Name)
	assert.Equal(t, blahOrgUser.Org, verUser.Org)
	resp.Body.Close()

	//try to list for both in both domains one should faile
	// create user for both in both doamins - one should fail
}

func TestCreateUserOther(t *testing.T) {
	server, superToken, err := authAndGetToken("super", "password")
	assert.Nil(t, err)
	client := &http.Client{}
	defer server.Close()
	// bad json ok token
	userURL := fmt.Sprintf("%s/user/admin", server.URL)
	req, _ := http.NewRequest("POST", userURL, bytes.NewBufferString("BUMMER"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Golden-Ticket", superToken.Token)
	resp, err := client.Do(req)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode) // auth fail
	qUser := new(qdb.User)
	qUser.Name = "someUser"
	qUser.Org = ""
	jsonBytes, err := json.Marshal(qUser)
	assert.Nil(t, err)
	// bad json values ok token
	req, _ = http.NewRequest("POST", userURL, bytes.NewBuffer(jsonBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Golden-Ticket", superToken.Token)
	resp, err = client.Do(req)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode) // auth fail
	qUser.Name = "admin"
	jsonBytes, err = json.Marshal(qUser)
	assert.Nil(t, err)
	// bad json (org missing) - ok token
	req, _ = http.NewRequest("POST", userURL, bytes.NewBuffer(jsonBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Golden-Ticket", superToken.Token)
	resp, err = client.Do(req)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode) // auth fail
	db := new(qdb.Qdb)
	db.Type = "mongodb"
	db.Timeout = time.Second * 10
	db.URL = "localhost"
	session, err := db.Open()
	assert.Nil(t, err)
	bacl := qdb.CreateACL("", "blah.org", "crud")
	blahOrgUser := new(qdb.User)
	blahOrgUser.Name = "admin"
	blahOrgUser.Password = "password"
	blahOrgUser.Org = "blah.org"
	blahOrgUser.ACL = append(blahOrgUser.ACL, *bacl) // can create modify list users in blah domain
	err = session.DeleteUser(blahOrgUser.Name, blahOrgUser.Org)
	assert.Nil(t, err)
	jsonBytes, err = json.Marshal(blahOrgUser)
	assert.Nil(t, err)
	// create blah admin with crud perm
	req, _ = http.NewRequest("POST", userURL, bytes.NewBuffer(jsonBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Golden-Ticket", superToken.Token)
	resp, err = client.Do(req)
	assert.Equal(t, http.StatusOK, resp.StatusCode) // ok

	// change directly in db blah admin perm to rd
	blahOrgUser, err = session.FindUser(blahOrgUser.Name, blahOrgUser.Org)
	bacl = qdb.CreateACL("", "blah.org", "rd") // we have now just delete and read access
	blahOrgUser.ACL[0] = *bacl                 // can create modify list users in blah domain
	err = session.InsertUser(blahOrgUser)
	assert.Nil(t, err)
	blahToken, err := GetToken(server, blahOrgUser.Name, blahOrgUser.Org, blahOrgUser.Password)
	assert.Nil(t, err)

	// create new user using blah admin token
	// it will fail as blah admin have rd perm only
	otherUserURL := fmt.Sprintf("%s/user/nuser", server.URL)
	blahOrgUser.Name = "nuser"
	jsonBytes, err = json.Marshal(blahOrgUser)
	assert.Nil(t, err)
	//try to update - it should failed because acl and we are using blah token
	req, _ = http.NewRequest("POST", otherUserURL, bytes.NewBuffer(jsonBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Golden-Ticket", blahToken.Token)
	resp, err = client.Do(req)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode) // auth fail

	// now get back crud perm blah
	blahOrgUser.Name = "admin"
	bacl = qdb.CreateACL("", "blah.org", "crud")
	blahOrgUser.ACL[0] = *bacl
	jsonBytes, err = json.Marshal(blahOrgUser)
	assert.Nil(t, err)

	// modify blah admin perm to crud using super token
	// so it will succeed
	req, _ = http.NewRequest("POST", userURL, bytes.NewBuffer(jsonBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Golden-Ticket", superToken.Token)
	// use super token - so blah admin update should be fine
	resp, err = client.Do(req)
	assert.Equal(t, http.StatusOK, resp.StatusCode) // ok
	//try again this should be fine blah admin have proper crud permission

	//create nuser with ru permission using blah admin ( now have proper permissions)
	blahOrgUser.ID = "" // remove admin ID
	blahOrgUser.Name = "nuser"
	bacl = qdb.CreateACL("", "blah.org", "ru")
	blahOrgUser.ACL[0] = *bacl
	jsonBytes, err = json.Marshal(blahOrgUser)
	assert.Nil(t, err)
	req, _ = http.NewRequest("POST", otherUserURL, bytes.NewBuffer(jsonBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Golden-Ticket", blahToken.Token)
	resp, err = client.Do(req)
	assert.Equal(t, http.StatusOK, resp.StatusCode) // ok
}

func TestGetUserOther(t *testing.T) {
	server, superToken, err := authAndGetToken("super", "password")
	assert.Nil(t, err)
	client := &http.Client{}
	defer server.Close()
	// bad json for GET token ok
	userURL := fmt.Sprintf("%s/user/admin", server.URL)
	req, _ := http.NewRequest("GET", userURL, bytes.NewBufferString("BUMMER"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Golden-Ticket", superToken.Token)
	resp, err := client.Do(req)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode) // auth fail
	qUser := new(qdb.User)
	qUser.Name = "someUser"
	qUser.Org = ""
	jsonBytes, err := json.Marshal(qUser)
	assert.Nil(t, err)
	// json with bad values token ok
	req, _ = http.NewRequest("GET", userURL, bytes.NewBuffer(jsonBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Golden-Ticket", superToken.Token)
	resp, err = client.Do(req)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode) // auth fail

	db := new(qdb.Qdb)
	db.Type = "mongodb"
	db.Timeout = time.Second * 10
	db.URL = "localhost"
	session, err := db.Open()
	assert.Nil(t, err)
	blahOrgUser := new(qdb.User)
	blahOrgUser.Name = "nuser"
	blahOrgUser.Password = "password"
	blahOrgUser.Org = "blah.org"
	blahToken, err := GetToken(server, blahOrgUser.Name, blahOrgUser.Org, blahOrgUser.Password)
	assert.Nil(t, err)

	//blahOrgUser.ACL = append(blahOrgUser.ACL, *bacl) // can create modify list users in blah domain
	// delete user from db
	err = session.DeleteUser(blahOrgUser.Name, blahOrgUser.Org)
	assert.Nil(t, err)
	jsonBytes, err = json.Marshal(blahOrgUser)
	assert.Nil(t, err)

	// self search not correct token - should be nuser
	otherUserURL := fmt.Sprintf("%s/user/nuser", server.URL)
	req, _ = http.NewRequest("GET", otherUserURL, bytes.NewBufferString(""))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Golden-Ticket", blahToken.Token)
	resp, err = client.Do(req)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode) // auth fail

	// user not exists in db content  = 0 ( self search)
	req, _ = http.NewRequest("GET", otherUserURL, bytes.NewBufferString(""))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Golden-Ticket", superToken.Token)
	resp, err = client.Do(req)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode) // auth fail

	//user not exists in db but json is ok token is ok
	jsonBytes, err = json.Marshal(blahOrgUser)
	assert.Nil(t, err)
	req, _ = http.NewRequest("GET", otherUserURL, bytes.NewBuffer(jsonBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Golden-Ticket", superToken.Token)
	resp, err = client.Do(req)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode) // auth fail
	assert.Nil(t, session.InsertUser(blahOrgUser))

}

func TestStat(t *testing.T) {
	server, superToken, err := authAndGetToken("super", "password")
	assert.Nil(t, err)
	client := &http.Client{}
	defer server.Close()
	// bad json ok token
	statURL := fmt.Sprintf("%s/stat", server.URL)
	req, _ := http.NewRequest("GET", statURL, bytes.NewBufferString(""))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Golden-Ticket", superToken.Token)
	resp, err := client.Do(req)
	assert.Nil(t, err)
	body, _ := ioutil.ReadAll(resp.Body)
	assert.True(t, len(string(body)) > 10)
	resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode) // auth fail

}
