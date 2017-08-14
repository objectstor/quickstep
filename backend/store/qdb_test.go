package qdb

import (
	"fmt"
	"testing"
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/stretchr/testify/assert"
)

func TestDbOpen(t *testing.T) {
	db := new(Qdb)
	_, err := db.Open()
	assert.Error(t, err)
	db.Type = "Blah"
	_, err = db.Open()
	assert.Error(t, err)
	db.URL = "localhostBadName"
	// don't wait use short timeout
	db.Timeout = time.Second * 1
	_, err = db.Open()
	assert.Error(t, err)
	db.Type = "Mysql" // this should fail for now
	_, err = db.Open()
	assert.Error(t, err)
	db.Type = "mongodb"
	_, err = db.Open()
	assert.Error(t, err)
	db.Timeout = time.Second * 10
	db.URL = "localhost"
	session, err := db.Open()
	assert.Nil(t, err)
	defer session.Close()
	newSession := session.New()
	assert.NotNil(t, newSession)
	defer newSession.Close()
	err = db.Close()
	assert.Nil(t, err)
}

func TestDBUserInsertFind(t *testing.T) {
	db := new(Qdb)
	db.Type = "mongodb"
	db.Timeout = time.Second * 10
	db.URL = "localhost"
	session, err := db.Open()
	assert.Nil(t, err)
	user := new(User)
	user.ID = bson.NewObjectId()
	user.Name = "super"
	user.Password = "password"
	err = session.InsertUser(user)
	assert.Nil(t, err)
	fmt.Printf("INSERTED\n")
	newUser := session.FindUser("super")
	assert.NotNil(t, newUser)
	assert.Equal(t, user.Name, newUser.Name)
	emptyUser := session.FindUser("super_one")
	assert.Empty(t, emptyUser.Name)
	defer session.Close()
	defer db.Close()
}
