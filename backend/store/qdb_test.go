package qdb

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	//mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
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
	defer db.Close()
	user := new(User)
	user.ID = bson.NewObjectId()
	user.Name = "super"
	user.Password = "password"
	user.ACL = ":crw"
	err = session.InsertUser(user)
	assert.Nil(t, err)
	newUser, nerr := session.FindUser("super")
	assert.NotNil(t, newUser)
	assert.Nil(t, nerr)
	assert.Equal(t, strings.Compare(user.Name, newUser.Name), 0)
	_, dberr := session.FindUser("super_one")
	assert.Equal(t, EntryNotFound(dberr), true)
	defer session.Close()
	_, dberr = session.FindUser("")
	assert.Error(t, dberr)
	emptyuser := new(User)
	err = session.InsertUser(emptyuser)
	assert.Error(t, err)
	if session != nil {
		session.mgoSession = nil
		err = session.InsertUser(emptyuser)
		assert.Error(t, err)
	}

}
