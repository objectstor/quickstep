package qrouter

import (

	//"log"

	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"quickstep/backend/store"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gopkg.in/mgo.v2/bson"
)

type RespTask struct {
	ID string `json:"taskid"`
}

func CreateRestTask(serverURL string, acl string, token *JSONToken, task *qdb.Task) error {
	taskURL := fmt.Sprintf("%s/task", serverURL)
	client := &http.Client{}
	jsonPayload, err := json.Marshal(task)
	if err != nil {
		return err
	}
	req, _ := http.NewRequest("PUT", taskURL, bytes.NewBuffer(jsonPayload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Golden-Ticket", token.Token)
	if len(acl) > 0 {
		req.Header.Set("X-Task-ACL", acl)
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}
	return nil
}

func CreateRestUser(serverURL string, token *JSONToken, user *qdb.User) error {
	// create 2 users
	userURL := fmt.Sprintf("%s/user/%s", serverURL, user.Name)
	client := &http.Client{}
	jsonPayload, err := json.Marshal(user)
	if err != nil {
		return err
	}
	req, _ := http.NewRequest("POST", userURL, bytes.NewBuffer(jsonPayload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Golden-Ticket", token.Token)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}
	return nil
}

func TestTask(t *testing.T) {
	var task qdb.Task
	server, superToken, err := authAndGetToken("super", "secret")
	assert.Nil(t, err)
	client := &http.Client{}
	defer server.Close()
	jsonStr := []byte("")
	taskURL := fmt.Sprintf("%s/task", server.URL)
	req, _ := http.NewRequest("PUT", taskURL, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Golden-Ticket", superToken.Token)
	resp, err := client.Do(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode) // auth succeed
	task.Name = "New_task"
	task.Private = false
	task.Status = "NEW"
	task.DeadLineTime = time.Now().Add(time.Hour * 24 * 7)
	// bad ID's
	task.ParentID = "AAA"
	jsonStr, err = json.Marshal(task)
	assert.Nil(t, err)
	req, _ = http.NewRequest("PUT", taskURL, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Golden-Ticket", superToken.Token)
	resp, err = client.Do(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode) // auth succeed

	task.ParentID = ""
	task.ChildID = append(task.ChildID, "AJA")
	jsonStr, err = json.Marshal(task)
	assert.Nil(t, err)
	req, _ = http.NewRequest("PUT", taskURL, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Golden-Ticket", superToken.Token)
	resp, err = client.Do(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode) // auth succeed

	//bad name
	task.ChildID = append(task.ChildID[:0]) // remove id's
	task.Name = ""
	jsonStr, err = json.Marshal(task)
	assert.Nil(t, err)
	req, _ = http.NewRequest("PUT", taskURL, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Golden-Ticket", superToken.Token)
	resp, err = client.Do(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode) // auth succeed

	//bad time
	task.Name = "blah_one"
	task.DeadLineTime = time.Time{}
	jsonStr, err = json.Marshal(task)
	assert.Nil(t, err)
	req, _ = http.NewRequest("PUT", taskURL, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Golden-Ticket", superToken.Token)
	resp, err = client.Do(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode) // auth succeed

	//bad time
	task.Name = "blah_one"
	task.DeadLineTime = time.Now()
	task.CreationTime = task.DeadLineTime.Add(time.Microsecond + 1)
	jsonStr, err = json.Marshal(task)
	assert.Nil(t, err)
	req, _ = http.NewRequest("PUT", taskURL, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Golden-Ticket", superToken.Token)
	resp, err = client.Do(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode) // auth succeed

	//ok
	tm := time.Now()
	task.Name = fmt.Sprintf("blah_one_%s", tm.Format("20060102150405"))
	task.DeadLineTime = task.DeadLineTime.Add(time.Hour + 24)
	task.CreationTime = time.Now()
	task.ItemType = qdb.TaskItem
	jsonStr, err = json.Marshal(task)
	assert.Nil(t, err)
	req, _ = http.NewRequest("PUT", taskURL, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Golden-Ticket", superToken.Token)
	resp, err = client.Do(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode) // auth succeed
	// bad  acl
	acl := new(qdb.ACLPerm)
	//acl.User =
	acl.Domain = "blah.org"
	acl.Create = true
	acl.Delete = true
	acl.Read = true
	acl.Update = true
	aclStr, err := json.Marshal(acl)
	assert.Nil(t, err)
	req, _ = http.NewRequest("PUT", taskURL, bytes.NewBuffer(jsonStr))
	req.Header.Set("X-Task-ACL", string(aclStr))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Golden-Ticket", superToken.Token)
	resp, err = client.Do(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// good acl
	acl.User = "admin"
	aclStr, err = json.Marshal(acl)
	assert.Nil(t, err)
	req, _ = http.NewRequest("PUT", taskURL, bytes.NewBuffer(jsonStr))
	req.Header.Set("X-Task-ACL", string(aclStr))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Golden-Ticket", superToken.Token)
	resp, err = client.Do(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode) // auth succeed

	// no acl
	body, _ := ioutil.ReadAll(resp.Body)
	resptask := new(RespTask)
	err = json.Unmarshal(body, resptask)
	assert.Nil(t, err)
	resp.Body.Close()
	assert.True(t, bson.IsObjectIdHex(resptask.ID))

	// bad parrent ID - not exists
	ps := bson.NewObjectId()
	task.ParentID = ps.Hex()
	jsonStr, err = json.Marshal(task)
	req, _ = http.NewRequest("PUT", taskURL, bytes.NewBuffer(jsonStr))
	req.Header.Set("X-Task-ACL", string(aclStr))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Golden-Ticket", superToken.Token)
	resp, err = client.Do(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	// bad child - not exists

	task.ParentID = ""
	task.ChildID = append(task.ChildID, bson.NewObjectId().Hex())
	jsonStr, err = json.Marshal(task)
	req, _ = http.NewRequest("PUT", taskURL, bytes.NewBuffer(jsonStr))
	req.Header.Set("X-Task-ACL", string(aclStr))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Golden-Ticket", superToken.Token)
	resp, err = client.Do(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// bad childID
	task.ChildID[0] = "BLA"
	jsonStr, err = json.Marshal(task)
	req, _ = http.NewRequest("PUT", taskURL, bytes.NewBuffer(jsonStr))
	req.Header.Set("X-Task-ACL", string(aclStr))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Golden-Ticket", superToken.Token)
	resp, err = client.Do(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	//body, _ = ioutil.ReadAll(resp.Body)
	//fmt.Println(string(body))
	//resp.Body.Close()

}

func TestPuACL(t *testing.T) {
	server, superToken, err := authAndGetToken("super", "secret")

	fooOrgUser := new(qdb.User)
	fooOrgUser.Name = fmt.Sprintf("foo_admin_%s", time.Now().Format("20060102150405"))
	fooOrgUser.Secret = "secret"
	fooOrgUser.Org = "foo.org"
	facl := qdb.CreateACL("", "foo.org", "crud")
	fooOrgUser.ACL = append(fooOrgUser.ACL, *facl) // can create modify read users in blah.org and foo.org domain
	facl = qdb.CreateACL("", "boo.org", "crud")
	fooOrgUser.ACL = append(fooOrgUser.ACL, *facl) // can create modify read users in blah.org and foo.org domain

	booOrgUser := new(qdb.User)
	booOrgUser.Name = fmt.Sprintf("boo_admin_%s", time.Now().Format("20060102150405"))
	booOrgUser.Secret = "secret"
	booOrgUser.Org = "boo.org"
	facl = qdb.CreateACL("", "boo.org", "crud")
	booOrgUser.ACL = append(booOrgUser.ACL, *facl) // can create modify read users in blah.org and foo.org domain

	err = CreateRestUser(server.URL, superToken, fooOrgUser)
	assert.Nil(t, err)
	err = CreateRestUser(server.URL, superToken, booOrgUser)
	assert.Nil(t, err)
	tm := time.Now()

	btask := new(qdb.Task)
	btask.Name = fmt.Sprintf("task_for_boo_user_%s", tm.Format("20060102150405"))
	btask.Private = false
	btask.Status = "NEW"
	btask.ItemType = qdb.TaskItem
	btask.CreationTime = time.Now()
	btask.DeadLineTime = btask.CreationTime.Add(time.Hour + 24)
	booacl := new(qdb.ACLPerm)
	booacl.User = booOrgUser.Name
	booacl.Domain = booOrgUser.Org
	booacl.Create = true
	booacl.Delete = true
	booacl.Read = true
	booacl.Update = true
	booaclJSON, err := json.Marshal(booacl)
	assert.Nil(t, err)

	ftask := new(qdb.Task)
	ftask.Name = fmt.Sprintf("task_for_foo_user_%s", tm.Format("20060102150405"))
	ftask.Private = false
	ftask.Status = "NEW"
	btask.ItemType = qdb.TaskItem
	ftask.CreationTime = time.Now()
	ftask.DeadLineTime = ftask.CreationTime.Add(time.Hour + 24)
	fooacl := new(qdb.ACLPerm)
	fooacl.User = fooOrgUser.Name
	fooacl.Domain = fooOrgUser.Org
	fooacl.Create = true
	fooacl.Delete = true
	fooacl.Read = true
	fooacl.Update = true
	fooaclJSON, err := json.Marshal(fooacl)

	fooToken, err := GetToken(server, fooOrgUser.Name, fooOrgUser.Org, fooOrgUser.Secret)
	assert.Nil(t, err)
	booToken, err := GetToken(server, booOrgUser.Name, booOrgUser.Org, booOrgUser.Secret)
	//assert.Nil(t, err)
	//create task for boo by foo
	// this should succeed as foo have access to boo
	err = CreateRestTask(server.URL, string(booaclJSON), fooToken, btask)
	assert.Nil(t, err)

	// create task for foo by boo
	//this will fail as  boo dont have access to foo
	err = CreateRestTask(server.URL, string(fooaclJSON), booToken, ftask)
	assert.Error(t, err)

}

func TestGetAllTasks(t *testing.T) {
	var task qdb.Task
	var jsonStr []byte
	var tasks []qdb.Task
	server, superToken, err := authAndGetToken("super", "secret")
	assert.Nil(t, err)
	client := &http.Client{}
	defer server.Close()
	taskURL := fmt.Sprintf("%s/task", server.URL)
	for i := 0; i < 10; i++ {
		tm := time.Now()
		task.Name = fmt.Sprintf("super_task_%s_%d", tm.Format("20060102150405"), i)
		task.ItemType = qdb.TaskItem
		task.CreationTime = time.Now()
		task.DeadLineTime = task.CreationTime.Add(time.Hour + 24)
		jsonStr, err = json.Marshal(task)
		assert.Nil(t, err)
		req, _ := http.NewRequest("PUT", taskURL, bytes.NewBuffer(jsonStr))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Golden-Ticket", superToken.Token)
		resp, err := client.Do(req)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	}
	req, _ := http.NewRequest("GET", taskURL, bytes.NewBufferString(""))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Golden-Ticket", superToken.Token)
	resp, err := client.Do(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	decoder := json.NewDecoder(resp.Body)
	assert.Nil(t, decoder.Decode(&tasks))
	assert.True(t, len(tasks) > 10) // at lleast 10 + 1 from previous tests
	resp.Body.Close()
}

func TestUpdateTasks(t *testing.T) {
	var task qdb.Task
	var jsonStr []byte
	server, superToken, err := authAndGetToken("super", "secret")
	assert.Nil(t, err)
	client := &http.Client{}
	defer server.Close()
	taskURL := fmt.Sprintf("%s/task", server.URL)
	tm := time.Now()
	task.Name = fmt.Sprintf("super_new_task_%s", tm.Format("20060102150405"))
	task.ItemType = qdb.TaskItem
	task.CreationTime = time.Now()
	task.DeadLineTime = task.CreationTime.Add(time.Hour + 24)
	jsonStr, err = json.Marshal(task)
	assert.Nil(t, err)
	req, _ := http.NewRequest("PUT", taskURL, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Golden-Ticket", superToken.Token)
	resp, err := client.Do(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, _ := ioutil.ReadAll(resp.Body)
	resptask := new(RespTask)
	assert.Nil(t, json.Unmarshal(body, resptask))
	resp.Body.Close()
	postTaskURL := fmt.Sprintf("%s/task/%s", server.URL, resptask.ID)
	fmt.Println(postTaskURL)
	req, _ = http.NewRequest("POST", postTaskURL, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Golden-Ticket", superToken.Token)
	resp, err = client.Do(req)
	fmt.Println(resp.StatusCode)
}
