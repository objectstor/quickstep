package qrouter

import (

	//"log"

	"bytes"
	"encoding/json"
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

func TestTask(t *testing.T) {
	var task qdb.Task
	server, superToken, err := authAndGetToken("super", "password")
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

func TestGetAllTasks(t *testing.T) {
	var task qdb.Task
	var jsonStr []byte
	var tasks []qdb.Task
	server, superToken, err := authAndGetToken("super", "password")
	assert.Nil(t, err)
	client := &http.Client{}
	defer server.Close()
	taskURL := fmt.Sprintf("%s/task", server.URL)
	for i := 0; i < 10; i++ {
		tm := time.Now()
		task.Name = fmt.Sprintf("super_task_%s_%d", tm.Format("20060102150405"), i)
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
