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

type JsonToken struct {
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
	loginUrl := fmt.Sprintf("%s/login", server.URL)
	res, err := http.Get(loginUrl)
	assert.Equal(t, 404, res.StatusCode)
	var jsonStr = []byte(`{"title":"Bummer"}`)
	req, err := http.NewRequest("POST", loginUrl, bytes.NewBuffer(jsonStr))
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
	var myToken JsonToken
	// test login cases
	super := new(User)
	db := new(qdb.Qdb)
	db.Type = "mongodb"
	db.Timeout = time.Second * 10
	db.Url = "localhost"
	session, err := db.Open()
	assert.Nil(t, err)
	router, err := New("localhost:9090", session)
	assert.Nil(t, err)
	err = router.Enable()
	assert.Nil(t, err)
	server := httptest.NewServer(router.Mux)
	//server.Start()
	defer server.Close()
	loginUrl := fmt.Sprintf("%s/login", server.URL)
	res, err := http.Get(loginUrl)
	assert.Equal(t, 404, res.StatusCode)
	var jsonStr = []byte(`{"title":"Buy cheese and bread for breakfast."}`)
	req, err := http.NewRequest("POST", loginUrl, bytes.NewBuffer(jsonStr))
	//req.Header.Set("X-Custom-Header", "myvalue")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	resp.Body.Close()
	assert.Nil(t, err)
	assert.Equal(t, 403, resp.StatusCode)

	super.Name = "super"
	jsonStr, err = json.Marshal(super)
	assert.Nil(t, err)
	req, err = http.NewRequest("POST", loginUrl, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)
	assert.Nil(t, err)
	assert.Equal(t, 403, resp.StatusCode)
	resp.Body.Close()

	super.Name = ""
	super.Password = "LamePassword"
	jsonStr, err = json.Marshal(super)
	assert.Nil(t, err)
	req, err = http.NewRequest("POST", loginUrl, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)
	assert.Nil(t, err)
	assert.Equal(t, 403, resp.StatusCode)
	resp.Body.Close()

	super.Name = "super"
	super.Password = "LamePassword"
	jsonStr, err = json.Marshal(super)
	assert.Nil(t, err)
	req, err = http.NewRequest("POST", loginUrl, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)
	defer resp.Body.Close()
	assert.Equal(t, 201, resp.StatusCode)
	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(body, &myToken)
	fmt.Println("Token:", myToken.Token)
	assert.NotEmpty(t, myToken.Token)
}
