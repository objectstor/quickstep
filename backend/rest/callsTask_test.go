package qrouter

import (

	//"log"

	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTask(t *testing.T) {
	server, superToken, err := authAndGetToken("super", "password")
	assert.Nil(t, err)
	client := &http.Client{}
	defer server.Close()
	jsonStr := []byte("")
	taskURL := fmt.Sprintf("%s/task", server.URL)
	req, _ := http.NewRequest("PUT", taskURL, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Golden-Ticket", superToken.Token)
	acl := fmt.Sprintf("%s:%s:%s,%s:%s:%s", "someUser", "blah.org", "cr", "otherUser", "foo.org", "r")
	req.Header.Set("X-Task-ACL", acl)
	resp, err := client.Do(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode) // auth succeed
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))
	resp.Body.Close()
}
